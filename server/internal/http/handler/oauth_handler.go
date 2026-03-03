package handler

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/auth"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type OAuthHandler struct {
	svc      *auth.Service
	stateTTL time.Duration
}

func NewOAuthHandler(svc *auth.Service, stateTTL time.Duration) *OAuthHandler {
	return &OAuthHandler{svc: svc, stateTTL: stateTTL}
}

// ListProviders godoc
// @Summary 获取可用的 OAuth 登录提供方
// @Tags Auth
// @Produce json
// @Success 200 {object} contract.ProviderListRespEnvelope
// @Router /auth/providers [get]
func (h *OAuthHandler) ListProviders(c *fiber.Ctx) error {
	items, err := h.svc.ListProviders(c.Context())
	if err != nil {
		return err
	}
	resp := make([]contract.OAuthProviderResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, contract.OAuthProviderResp{
			Key:          item.ProviderKey,
			DisplayName:  item.DisplayName,
			Scopes:       splitScopes(item.Scopes),
			PKCERequired: item.PKCERequired,
		})
	}
	return response.Success(c, resp)
}

// Authorize godoc
// @Summary 获取指定 provider 的授权跳转地址
// @Tags Auth
// @Produce json
// @Param provider path string true "provider key"
// @Param redirect_uri query string false "登录成功后的前端跳转地址"
// @Success 200 {object} contract.AuthorizeRespEnvelope
// @Router /auth/providers/{provider}/authorize [get]
func (h *OAuthHandler) Authorize(c *fiber.Ctx) error {
	provider := c.Params("provider")
	redirect := c.Query("redirect_uri")
	contextNonce, err := auth.GenerateContextNonce()
	if err != nil {
		return err
	}
	res, err := h.svc.Authorize(c.Context(), provider, redirect, contextNonce, h.stateTTL)
	if err != nil {
		return err
	}
	setOAuthStateNonceCookie(c, contextNonce, h.stateTTL)
	return response.Success(c, contract.AuthorizeResp{
		AuthURL:       res.AuthURL,
		State:         res.State,
		CodeChallenge: res.CodeChallenge,
	})
}

// Callback godoc
// @Summary OAuth 回调，完成自动登录并签发 JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param provider path string true "provider key"
// @Param request body contract.OAuthCallbackReq true "回调参数"
// @Success 200 {object} contract.LoginRespEnvelope
// @Router /auth/providers/{provider}/callback [post]
func (h *OAuthHandler) Callback(c *fiber.Ctx) error {
	provider := c.Params("provider")
	var req contract.OAuthCallbackReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if req.Code == "" || req.State == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "code/state 不能为空")
	}
	contextNonce := readOAuthStateNonceCookie(c)
	clearOAuthStateNonceCookie(c)
	result, err := h.svc.LoginWithProvider(c.Context(), auth.OAuthLoginCmd{
		Provider:     provider,
		Code:         req.Code,
		State:        req.State,
		Redirect:     req.RedirectURI,
		ContextNonce: contextNonce,
	})
	if err != nil {
		if err == auth.ErrUserDisabled {
			return response.NewBizErrorWithMsg(response.Unauthorized, "账号已被禁用")
		}
		if err == auth.ErrInvalidOAuthIdentity {
			return response.NewBizErrorWithMsg(response.ParamsError, "OAuth 身份信息无效，请检查 provider 的用户信息映射配置")
		}
		return err
	}
	return response.Success(c, contract.LoginResp{
		Token: result.Token,
		User:  contract.ToUserResp(result.User),
	})
}

func splitScopes(sc string) []string {
	return strings.Fields(sc)
}

func setOAuthStateNonceCookie(c *fiber.Ctx, nonce string, ttl time.Duration) {
	maxAge := int(ttl.Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.Cookie(&fiber.Cookie{
		Name:     auth.OAuthStateNonceCookieName,
		Value:    nonce,
		Path:     "/",
		MaxAge:   maxAge,
		HTTPOnly: true,
		Secure:   strings.EqualFold(c.Protocol(), "https"),
		SameSite: "Lax",
	})
}

func clearOAuthStateNonceCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     auth.OAuthStateNonceCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   strings.EqualFold(c.Protocol(), "https"),
		SameSite: "Lax",
	})
}

func readOAuthStateNonceCookie(c *fiber.Ctx) string {
	return strings.TrimSpace(c.Cookies(auth.OAuthStateNonceCookieName))
}
