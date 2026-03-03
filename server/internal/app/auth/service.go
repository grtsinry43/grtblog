package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"

	"github.com/grtsinry43/grtblog-v2/server/internal/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/security/jwt"
)

var (
	ErrProviderNotConfigured = errors.New("oauth provider not configured")
	ErrRegisterClosed        = errors.New("register is only allowed for initial admin setup")
	ErrLastOAuthBinding      = errors.New("cannot unbind last oauth binding for non-admin user")
	ErrUserDisabled          = errors.New("user disabled")
	ErrPasswordTooWeak       = errors.New("password must be at least 8 characters")
	ErrAdminOnly             = errors.New("login is restricted to admin users")
	ErrInvalidOAuthIdentity  = errors.New("oauth provider identity is invalid")
)

// ExternalProvider 用于未来扩展 OAuth/OIDC, 当前仅定义接口。
type ExternalProvider interface {
	Name() string
	Exchange(ctx context.Context, code string) (*ExternalIdentity, error)
}

type ExternalIdentity struct {
	Provider   string
	ProviderID string
	Email      string
	Username   string
	Name       string
	Avatar     string
}

type Service struct {
	users      identity.Repository
	oauthRepo  identity.OAuthProviderRepository
	stateStore StateStore
	manager    *jwt.Manager
	providers  map[string]ExternalProvider
}

func NewService(repo identity.Repository, oauthRepo identity.OAuthProviderRepository, manager *jwt.Manager, stateStore StateStore, authCfg config.AuthConfig) *Service {
	return &Service{
		users:      repo,
		oauthRepo:  oauthRepo,
		stateStore: stateStore,
		manager:    manager,
		providers:  make(map[string]ExternalProvider),
	}
}

func (s *Service) RegisterProvider(provider ExternalProvider) {
	if provider == nil {
		return
	}
	s.providers[strings.ToLower(provider.Name())] = provider
}

type RegisterCmd struct {
	Username string
	Nickname string
	Email    string
	Password string
}

type LoginCmd struct {
	Credential string
	Password   string
}

type LoginResult struct {
	Token  string
	User   identity.User
	Claims *jwt.Claims
}

type UpdateProfileCmd struct {
	UserID   int64
	Nickname string
	Avatar   string
	Email    string
}

type ChangePasswordCmd struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

type AccessInfo struct {
	User identity.User
}

func (s *Service) Register(ctx context.Context, cmd RegisterCmd) (*identity.User, error) {
	cmd.Username = strings.TrimSpace(cmd.Username)
	cmd.Email = strings.TrimSpace(cmd.Email)
	if cmd.Username == "" || cmd.Email == "" || cmd.Password == "" {
		return nil, identity.ErrInvalidCredentials
	}
	if len(cmd.Password) < 8 {
		return nil, ErrPasswordTooWeak
	}
	if _, err := mail.ParseAddress(cmd.Email); err != nil {
		return nil, identity.ErrInvalidCredentials
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &identity.User{
		Username: cmd.Username,
		Nickname: firstNonEmpty(cmd.Nickname, cmd.Username),
		Email:    cmd.Email,
		Password: string(hashed),
		IsActive: true,
	}
	if total, err := s.users.CountUsers(ctx); err != nil {
		return nil, err
	} else if total == 0 {
		user.IsAdmin = true
	} else {
		return nil, ErrRegisterClosed
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *Service) Login(ctx context.Context, cmd LoginCmd) (*LoginResult, error) {
	if cmd.Credential == "" || cmd.Password == "" {
		return nil, identity.ErrInvalidCredentials
	}
	user, err := s.users.FindByCredential(ctx, cmd.Credential)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.Password)) != nil {
		return nil, identity.ErrInvalidCredentials
	}
	if !user.IsActive {
		return nil, ErrUserDisabled
	}
	if !user.IsAdmin {
		return nil, ErrAdminOnly
	}
	token, claims, err := s.manager.Generate(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}
	claims.Subject = user.Username
	user.Password = ""
	return &LoginResult{
		Token:  token,
		User:   *user,
		Claims: claims,
	}, nil
}

type OAuthLoginCmd struct {
	Provider     string
	Code         string
	State        string
	Redirect     string
	ContextNonce string
}

type OAuthAuthorizeResult struct {
	AuthURL       string
	State         string
	CodeChallenge string
}

func (s *Service) LoginWithProvider(ctx context.Context, cmd OAuthLoginCmd) (*LoginResult, error) {
	if s.oauthRepo == nil {
		return nil, ErrProviderNotConfigured
	}
	providerCfg, err := s.oauthRepo.GetByKey(ctx, cmd.Provider)
	if err != nil {
		return nil, err
	}
	if s.stateStore == nil {
		return nil, errors.New("state store not configured")
	}
	stateData, err := s.stateStore.Load(ctx, cmd.State)
	if err != nil {
		return nil, err
	}
	defer s.stateStore.Delete(ctx, cmd.State)
	if err := validateOAuthState(stateData, cmd); err != nil {
		return nil, err
	}

	conf := buildOAuth2Config(providerCfg)
	options := []oauth2.AuthCodeOption{}
	if providerCfg.PKCERequired && stateData.CodeVerifier != "" {
		options = append(options, oauth2.SetAuthURLParam("code_verifier", stateData.CodeVerifier))
	}
	token, err := conf.Exchange(ctx, cmd.Code, options...)
	if err != nil {
		return nil, err
	}

	external, err := fetchExternalIdentity(ctx, providerCfg, token)
	if err != nil {
		return nil, err
	}

	// 映射 / 注册本地用户
	user, err := s.users.FindByOAuth(ctx, providerCfg.ProviderKey, external.ProviderID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			user, err = s.registerOAuthUser(ctx, external)
			if err != nil {
				return nil, err
			}
			var expPtr *time.Time
			if !token.Expiry.IsZero() {
				exp := token.Expiry
				expPtr = &exp
			}
			if bindErr := s.users.BindOAuthByProvider(ctx, identity.UserOAuth{
				UserID:       user.ID,
				ProviderKey:  providerCfg.ProviderKey,
				OAuthID:      external.ProviderID,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				ExpiresAt:    expPtr,
			}); bindErr != nil {
				return nil, bindErr
			}
		} else {
			return nil, err
		}
	}
	if !user.IsActive {
		return nil, ErrUserDisabled
	}

	jwtToken, claims, err := s.manager.Generate(user.ID, user.IsAdmin)
	if err != nil {
		return nil, err
	}
	claims.Subject = user.Username
	user.Password = ""
	return &LoginResult{
		Token:  jwtToken,
		User:   *user,
		Claims: claims,
	}, nil
}

// AccessInfo 返回最新的用户、角色与权限信息。
func (s *Service) AccessInfo(ctx context.Context, claims *jwt.Claims) (*AccessInfo, error) {
	if claims == nil {
		return nil, identity.ErrInvalidCredentials
	}
	user, err := s.users.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return &AccessInfo{
		User: *user,
	}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, cmd UpdateProfileCmd) (*identity.User, error) {
	if cmd.UserID == 0 {
		return nil, identity.ErrInvalidCredentials
	}
	cmd.Nickname = strings.TrimSpace(cmd.Nickname)
	cmd.Email = strings.TrimSpace(cmd.Email)
	cmd.Avatar = strings.TrimSpace(cmd.Avatar)
	updated, err := s.users.UpdateProfile(ctx, cmd.UserID, cmd.Nickname, cmd.Avatar, cmd.Email)
	if err != nil {
		return nil, err
	}
	updated.Password = ""
	return updated, nil
}

// CurrentUser 返回当前用户信息。
func (s *Service) CurrentUser(ctx context.Context, userID int64) (*identity.User, error) {
	if userID == 0 {
		return nil, identity.ErrInvalidCredentials
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *Service) ChangePassword(ctx context.Context, cmd ChangePasswordCmd) error {
	if cmd.UserID == 0 || cmd.NewPassword == "" {
		return identity.ErrInvalidCredentials
	}
	if len(cmd.NewPassword) < 8 {
		return ErrPasswordTooWeak
	}
	user, err := s.users.FindByID(ctx, cmd.UserID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cmd.OldPassword)) != nil {
		return identity.ErrInvalidCredentials
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(cmd.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, cmd.UserID, string(hashed))
}

func (s *Service) ListOAuthBindings(ctx context.Context, userID int64) ([]identity.UserOAuthBinding, error) {
	if userID == 0 {
		return nil, identity.ErrInvalidCredentials
	}
	return s.users.ListOAuthBindings(ctx, userID)
}

func (s *Service) BindOAuth(ctx context.Context, userID int64, cmd OAuthLoginCmd) error {
	if userID == 0 || cmd.Provider == "" || cmd.Code == "" || cmd.State == "" {
		return identity.ErrInvalidCredentials
	}
	if s.oauthRepo == nil || s.stateStore == nil {
		return ErrProviderNotConfigured
	}
	providerCfg, err := s.oauthRepo.GetByKey(ctx, cmd.Provider)
	if err != nil {
		return err
	}
	stateData, err := s.stateStore.Load(ctx, cmd.State)
	if err != nil {
		return err
	}
	defer s.stateStore.Delete(ctx, cmd.State)
	if err := validateOAuthState(stateData, cmd); err != nil {
		return err
	}

	conf := buildOAuth2Config(providerCfg)
	options := []oauth2.AuthCodeOption{}
	if providerCfg.PKCERequired && stateData.CodeVerifier != "" {
		options = append(options, oauth2.SetAuthURLParam("code_verifier", stateData.CodeVerifier))
	}
	token, err := conf.Exchange(ctx, cmd.Code, options...)
	if err != nil {
		return err
	}
	external, err := fetchExternalIdentity(ctx, providerCfg, token)
	if err != nil {
		return err
	}
	owner, err := s.users.FindByOAuth(ctx, providerCfg.ProviderKey, external.ProviderID)
	if err == nil && owner.ID != userID {
		return identity.ErrOAuthAlreadyBound
	}
	if err != nil && !errors.Is(err, identity.ErrUserNotFound) {
		return err
	}

	var expPtr *time.Time
	if !token.Expiry.IsZero() {
		exp := token.Expiry
		expPtr = &exp
	}
	return s.users.BindOAuthByProvider(ctx, identity.UserOAuth{
		UserID:       userID,
		ProviderKey:  providerCfg.ProviderKey,
		OAuthID:      external.ProviderID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    expPtr,
	})
}

func (s *Service) UnbindOAuth(ctx context.Context, userID int64, provider string) error {
	if userID == 0 || strings.TrimSpace(provider) == "" {
		return identity.ErrInvalidCredentials
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.IsAdmin {
		bindings, err := s.users.ListOAuthBindings(ctx, userID)
		if err != nil {
			return err
		}
		if len(bindings) <= 1 {
			return ErrLastOAuthBinding
		}
	}
	return s.users.UnbindOAuth(ctx, userID, provider)
}

// IsInitialized 用于判断是否已完成初始化（存在至少一个用户）。
func (s *Service) IsInitialized(ctx context.Context) (bool, error) {
	total, err := s.users.CountUsers(ctx)
	if err != nil {
		return false, err
	}
	return total > 0, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// ListProviders 列出启用的 OAuth 提供方。
func (s *Service) ListProviders(ctx context.Context) ([]identity.OAuthProvider, error) {
	if s.oauthRepo == nil {
		return nil, ErrProviderNotConfigured
	}
	items, err := s.oauthRepo.ListEnabled(ctx)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// Authorize 生成授权 URL 和 state（可含 PKCE）。
func (s *Service) Authorize(ctx context.Context, providerKey, redirect, contextNonce string, stateTTL time.Duration) (*OAuthAuthorizeResult, error) {
	if s.oauthRepo == nil || s.stateStore == nil {
		return nil, ErrProviderNotConfigured
	}
	contextNonceHash := HashContextNonce(contextNonce)
	if contextNonceHash == "" {
		return nil, errors.New("missing oauth state context")
	}
	cfg, err := s.oauthRepo.GetByKey(ctx, providerKey)
	if err != nil {
		return nil, err
	}
	oauthCfg := buildOAuth2Config(cfg)

	state, err := GenerateState()
	if err != nil {
		return nil, err
	}
	var codeVerifier, codeChallenge string
	if cfg.PKCERequired {
		codeVerifier, codeChallenge, err = GenerateCodeVerifier()
		if err != nil {
			return nil, err
		}
	}

	if err := s.stateStore.Save(ctx, state, OAuthState{
		Provider:         providerKey,
		Redirect:         redirect,
		CodeVerifier:     codeVerifier,
		ContextNonceHash: contextNonceHash,
		CreatedAt:        time.Now(),
	}, stateTTL); err != nil {
		return nil, err
	}

	authOpts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}
	if cfg.PKCERequired && codeChallenge != "" {
		authOpts = append(authOpts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}
	authURL := oauthCfg.AuthCodeURL(state, authOpts...)
	return &OAuthAuthorizeResult{
		AuthURL:       authURL,
		State:         state,
		CodeChallenge: codeChallenge,
	}, nil
}

func buildOAuth2Config(p *identity.OAuthProvider) *oauth2.Config {
	redirect := p.RedirectURITemplate
	redirect = strings.ReplaceAll(redirect, "{provider}", p.ProviderKey)
	return &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.AuthorizationEndpoint,
			TokenURL: p.TokenEndpoint,
		},
		Scopes:      splitScopes(p.Scopes),
		RedirectURL: redirect,
	}
}

func splitScopes(sc string) []string {
	parts := strings.Fields(sc)
	return parts
}

func validateOAuthState(stateData *OAuthState, cmd OAuthLoginCmd) error {
	if stateData == nil {
		return errors.New("state not found")
	}
	if stateData.Provider != cmd.Provider {
		return errors.New("state/provider mismatch")
	}

	expectedRedirect := strings.TrimSpace(stateData.Redirect)
	actualRedirect := strings.TrimSpace(cmd.Redirect)
	if expectedRedirect != actualRedirect {
		return errors.New("state/redirect mismatch")
	}
	if !VerifyContextNonce(cmd.ContextNonce, stateData.ContextNonceHash) {
		return errors.New("state/context mismatch")
	}
	return nil
}

type ExternalProfile struct {
	ID       string
	Email    string
	Username string
	Name     string
	Avatar   string
}

func fetchExternalIdentity(ctx context.Context, cfg *identity.OAuthProvider, token *oauth2.Token) (*ExternalIdentity, error) {
	profile := ExternalProfile{}
	if cfg.UserinfoEndpoint != "" {
		if err := fetchUserInfo(ctx, cfg.UserinfoEndpoint, token, &profile); err != nil {
			return nil, err
		}
	}
	id := profile.ID
	if id == "" {
		// 回退使用 AccessToken 哈希避免空 ID
		id = token.AccessToken
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidOAuthIdentity
	}
	return &ExternalIdentity{
		Provider:   cfg.ProviderKey,
		ProviderID: id,
		Email:      profile.Email,
		Username:   firstNonEmpty(profile.Username, profile.Email, profile.Name, id),
		Name:       profile.Name,
		Avatar:     profile.Avatar,
	}, nil
}

func fetchUserInfo(ctx context.Context, endpoint string, token *oauth2.Token, out *ExternalProfile) error {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	resp, err := client.Get(endpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("userinfo status: %s", resp.Status)
	}
	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return err
	}
	out.ID = firstString(raw, "sub", "id")
	out.Email = firstString(raw, "email")
	out.Username = firstString(raw, "preferred_username", "username", "login", "name")
	out.Name = firstString(raw, "name")
	out.Avatar = firstString(raw, "avatar_url", "picture")
	return nil
}

func firstString(raw map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := raw[k]; ok {
			if s := stringValue(v); s != "" {
				return s
			}
		}
	}
	return ""
}

func stringValue(v any) string {
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val)
	case json.Number:
		return strings.TrimSpace(val.String())
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	default:
		return ""
	}
}

// registerOAuthUser 根据外部信息注册本地用户。
// 只保证 username 唯一；email 允许重复，直接使用 OAuth 提供的邮箱。
// 绝不复用已有账号，始终创建新用户。
func (s *Service) registerOAuthUser(ctx context.Context, ext *ExternalIdentity) (*identity.User, error) {
	username := firstNonEmpty(ext.Username, ext.Email, ext.Provider+"_"+ext.ProviderID)
	username, err := s.nextAvailableOAuthUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	user := &identity.User{
		Username: username,
		Nickname: firstNonEmpty(ext.Name, username),
		Email:    strings.TrimSpace(ext.Email),
		Avatar:   ext.Avatar,
		IsActive: true,
	}
	if total, err := s.users.CountUsers(ctx); err != nil {
		return nil, err
	} else if total == 0 {
		user.IsAdmin = true
	}
	if err := s.users.Create(ctx, user); err != nil {
		// username 并发冲突时重新生成用户名重试一次
		if errors.Is(err, identity.ErrUserExists) {
			retryUsername, findErr := s.nextAvailableOAuthUsername(ctx, username)
			if findErr != nil {
				return nil, findErr
			}
			user.Username = retryUsername
			if retryErr := s.users.Create(ctx, user); retryErr == nil {
				return user, nil
			} else if !errors.Is(retryErr, identity.ErrUserExists) {
				return nil, retryErr
			}
		}
		return nil, err
	}
	return user, nil
}

func (s *Service) nextAvailableOAuthUsername(ctx context.Context, seed string) (string, error) {
	base := strings.TrimSpace(seed)
	if base == "" {
		base = "oauth_user"
	}
	const maxLen = 45
	if len(base) > maxLen {
		base = base[:maxLen]
	}

	if _, err := s.users.FindByUsername(ctx, base); err == nil {
		// occupied, continue
	} else if errors.Is(err, identity.ErrUserNotFound) {
		return base, nil
	} else {
		return "", err
	}

	for i := 1; i <= 64; i++ {
		suffix := "_" + strconv.Itoa(i)
		prefix := base
		if len(prefix)+len(suffix) > maxLen {
			prefix = prefix[:maxLen-len(suffix)]
		}
		candidate := prefix + suffix
		if _, err := s.users.FindByUsername(ctx, candidate); err == nil {
			continue
		} else if errors.Is(err, identity.ErrUserNotFound) {
			return candidate, nil
		} else {
			return "", err
		}
	}

	nonce, err := randomString(8)
	if err != nil {
		return "", err
	}
	suffix := "_" + strings.ToLower(strings.TrimRight(nonce, "="))
	if len(suffix) >= maxLen {
		suffix = suffix[:maxLen-1]
	}
	prefix := base
	if len(prefix)+len(suffix) > maxLen {
		prefix = prefix[:maxLen-len(suffix)]
	}
	return prefix + suffix, nil
}
