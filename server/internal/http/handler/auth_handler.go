package handler

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/auth"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/setupstate"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/security/turnstile"
)

type AuthHandler struct {
	svc       *auth.Service
	setupSvc  *setupstate.Service
	sysCfg    *sysconfig.Service
	turnstile TurnstileVerifier
}

// TurnstileVerifier 便于替换实现/注入 mock。
type TurnstileVerifier interface {
	Verify(ctx context.Context, token, remoteIP string, settings turnstile.Settings) error
}

func NewAuthHandler(svc *auth.Service, setupSvc *setupstate.Service, sysCfg *sysconfig.Service, verifier TurnstileVerifier) *AuthHandler {
	return &AuthHandler{svc: svc, setupSvc: setupSvc, sysCfg: sysCfg, turnstile: verifier}
}

// Register godoc
// @Summary 用户注册
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body contract.RegisterReq true "注册参数"
// @Success 200 {object} contract.RegisterRespEnvelope
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req contract.RegisterReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.verifyTurnstile(c, req.TurnstileToken); err != nil {
		return err
	}
	var cmd auth.RegisterCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}
	user, err := h.svc.Register(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, auth.ErrRegisterClosed) {
			return response.NewBizErrorWithMsg(response.ParamsError, "仅允许初始化管理员账号，普通用户请使用 OAuth 登录")
		}
		if errors.Is(err, auth.ErrPasswordTooWeak) {
			return response.NewBizErrorWithMsg(response.ParamsError, "密码至少需要 8 位")
		}
		if errors.Is(err, identity.ErrInvalidCredentials) {
			return response.NewBizErrorWithMsg(response.ParamsError, "用户名、邮箱和密码不能为空，且邮箱格式必须正确")
		}
		if errors.Is(err, identity.ErrUserExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "用户名或邮箱已存在")
		}
		return err
	}
	Audit(c, "auth.register", map[string]any{
		"userId":   user.ID,
		"username": user.Username,
		"email":    user.Email,
		"isAdmin":  user.IsAdmin,
	})
	return response.SuccessWithMessage(c, contract.ToUserResp(*user), "注册成功")
}

// Login godoc
// @Summary 用户登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body contract.LoginReq true "登录参数"
// @Success 200 {object} contract.LoginRespEnvelope
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req contract.LoginReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.verifyTurnstile(c, req.TurnstileToken); err != nil {
		return err
	}
	var cmd auth.LoginCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}
	result, err := h.svc.Login(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, auth.ErrUserDisabled) {
			return response.NewBizErrorWithMsg(response.Unauthorized, "账号已被禁用")
		}
		if errors.Is(err, identity.ErrInvalidCredentials) {
			return response.NewBizError(response.InvalidCredential)
		}
		return err
	}
	Audit(c, "auth.login", map[string]any{
		"userId":   result.User.ID,
		"username": result.User.Username,
		"isAdmin":  result.User.IsAdmin,
	})
	return response.Success(c, contract.LoginResp{
		Token: result.Token,
		User:  contract.ToUserResp(result.User),
	})
}

func (h *AuthHandler) verifyTurnstile(c *fiber.Ctx, token string) error {
	if h.turnstile == nil || h.sysCfg == nil {
		return nil
	}
	settings, err := h.sysCfg.Turnstile(c.Context())
	if err != nil {
		return response.NewBizErrorWithMsg(response.ServerError, "获取系统配置失败")
	}
	if !settings.Enabled {
		return nil
	}
	if err := h.turnstile.Verify(c.Context(), token, c.IP(), settings); err != nil {
		if errors.Is(err, turnstile.ErrVerificationFailed) {
			return response.NewBizErrorWithMsg(response.ParamsError, "人机校验未通过")
		}
		if errors.Is(err, turnstile.ErrMissingSecret) {
			return response.NewBizErrorWithMsg(response.ServerError, "人机校验未配置，请联系管理员")
		}
		return response.NewBizErrorWithMsg(response.ServerError, "人机校验失败，请稍后重试")
	}
	return nil
}

// UpdateProfile 更新当前登录用户的昵称/头像/邮箱。
func (h *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	var req contract.UpdateProfileReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	var cmd auth.UpdateProfileCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}
	cmd.UserID = claims.UserID
	user, err := h.svc.UpdateProfile(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUserExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "邮箱已被使用")
		}
		return err
	}
	Audit(c, "auth.update_profile", map[string]any{"userId": claims.UserID})
	return response.SuccessWithMessage(c, contract.ToUserResp(*user), "更新成功")
}

// ChangePassword 修改当前登录用户密码。
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	var req contract.ChangePasswordReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if req.NewPassword == "" || req.OldPassword == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "密码不能为空")
	}
	var cmd auth.ChangePasswordCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}
	cmd.UserID = claims.UserID
	if err := h.svc.ChangePassword(c.Context(), cmd); err != nil {
		if errors.Is(err, auth.ErrPasswordTooWeak) {
			return response.NewBizErrorWithMsg(response.ParamsError, "密码至少需要 8 位")
		}
		if errors.Is(err, identity.ErrInvalidCredentials) {
			return response.NewBizError(response.InvalidCredential)
		}
		return err
	}
	Audit(c, "auth.change_password", map[string]any{"userId": claims.UserID})
	return response.SuccessWithMessage[any](c, nil, "密码已更新")
}

// ListOAuthBindings 返回当前用户绑定的 OAuth 账号。
func (h *AuthHandler) ListOAuthBindings(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	items, err := h.svc.ListOAuthBindings(c.Context(), claims.UserID)
	if err != nil {
		return err
	}
	return response.Success(c, items)
}

// BindOAuth 绑定当前用户的 OAuth 账号。
func (h *AuthHandler) BindOAuth(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	provider := c.Params("provider")
	var req contract.OAuthCallbackReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if req.Code == "" || req.State == "" || provider == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "provider/code/state 不能为空")
	}
	contextNonce := readOAuthStateNonceCookie(c)
	clearOAuthStateNonceCookie(c)
	if err := h.svc.BindOAuth(c.Context(), claims.UserID, auth.OAuthLoginCmd{
		Provider:     provider,
		Code:         req.Code,
		State:        req.State,
		Redirect:     req.RedirectURI,
		ContextNonce: contextNonce,
	}); err != nil {
		if errors.Is(err, identity.ErrOAuthAlreadyBound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "该第三方账号已绑定其他用户")
		}
		if errors.Is(err, auth.ErrInvalidOAuthIdentity) {
			return response.NewBizErrorWithMsg(response.ParamsError, "OAuth 身份信息无效，请检查 provider 的用户信息映射配置")
		}
		return err
	}
	Audit(c, "auth.bind_oauth", map[string]any{
		"userId":   claims.UserID,
		"provider": provider,
	})
	return response.SuccessWithMessage[any](c, nil, "绑定成功")
}

// UnbindOAuth 解绑当前用户的 OAuth 账号。
func (h *AuthHandler) UnbindOAuth(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	provider := c.Params("provider")
	if provider == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "provider 不能为空")
	}
	if err := h.svc.UnbindOAuth(c.Context(), claims.UserID, provider); err != nil {
		if errors.Is(err, auth.ErrLastOAuthBinding) {
			return response.NewBizErrorWithMsg(response.ParamsError, "至少保留一种登录方式，当前账号无法解绑最后一个 OAuth")
		}
		if errors.Is(err, identity.ErrOAuthBindingNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "该账号未绑定此提供方")
		}
		return err
	}
	Audit(c, "auth.unbind_oauth", map[string]any{
		"userId":   claims.UserID,
		"provider": provider,
	})
	return response.SuccessWithMessage[any](c, nil, "解绑成功")
}

// AccessInfo 返回当前登录用户的角色/权限。
func (h *AuthHandler) AccessInfo(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	info, err := h.svc.AccessInfo(c.Context(), claims)
	if err != nil {
		return err
	}
	return response.Success(c, contract.AccessInfoResp{
		User: contract.ToUserResp(info.User),
	})
}

// Profile godoc
// @Summary 获取当前用户信息
// @Tags Auth
// @Produce json
// @Success 200 {object} contract.UserRespEnvelope
// @Security BearerAuth
// @Router /auth/profile [get]
func (h *AuthHandler) Profile(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	user, err := h.svc.CurrentUser(c.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "用户不存在")
		}
		return err
	}
	return response.Success(c, contract.ToUserResp(*user))
}

// InitState 返回是否需要初始化（无用户时为 false）。
func (h *AuthHandler) InitState(c *fiber.Ctx) error {
	initialized, err := h.svc.IsInitialized(c.Context())
	if err != nil {
		return response.NewBizErrorWithMsg(response.ServerError, "获取初始化状态失败")
	}
	return response.Success(c, contract.InitStateResp{
		Initialized: initialized,
	})
}

// SetupState 返回初始化准备状态（用户、管理员、站点信息）。
func (h *AuthHandler) SetupState(c *fiber.Ctx) error {
	if h.setupSvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "初始化服务未配置")
	}
	state, err := h.setupSvc.Evaluate(c.Context())
	if err != nil {
		return response.NewBizErrorWithMsg(response.ServerError, "获取初始化状态失败")
	}
	return response.Success(c, contract.SetupStateResp{
		HasUser:                state.HasUser,
		HasAdmin:               state.HasAdmin,
		WebsiteInfoReady:       state.WebsiteInfoReady,
		MissingWebsiteInfoKeys: state.MissingWebsiteInfoKeys,
		NeedsSetup:             state.NeedsSetup,
	})
}

// TurnstileState godoc
// @Summary 获取 Turnstile 配置状态
// @Tags Auth
// @Produce json
// @Success 200 {object} contract.TurnstileStateRespEnvelope
// @Router /auth/turnstile [get]
// TurnstileState 返回 Turnstile 是否启用及站点信息。
func (h *AuthHandler) TurnstileState(c *fiber.Ctx) error {
	if h.sysCfg == nil {
		return response.Success(c, contract.TurnstileStateResp{
			Enabled: false,
		})
	}
	settings, err := h.sysCfg.Turnstile(c.Context())
	if err != nil {
		return response.NewBizErrorWithMsg(response.ServerError, "获取系统配置失败")
	}
	resp := contract.TurnstileStateResp{
		Enabled: settings.Enabled,
	}
	if settings.Enabled {
		resp.SiteKey = settings.SiteKey
	}
	return response.Success(c, resp)
}
