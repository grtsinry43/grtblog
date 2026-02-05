package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/federationconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
)

type FederationFriendLinkHandler struct {
	cfgSvc          *federationconfig.Service
	instanceRepo    federation.FederationInstanceRepository
	linkRepo        social.FriendLinkRepository
	applicationRepo social.FriendLinkApplicationRepository
	resolver        *fedinfra.Resolver
	verifier        *fedinfra.Verifier
	rateLimiter     fedinfra.RateLimiter
	events          appEvent.Bus
}

func NewFederationFriendLinkHandler(
	cfgSvc *federationconfig.Service,
	instanceRepo federation.FederationInstanceRepository,
	linkRepo social.FriendLinkRepository,
	applicationRepo social.FriendLinkApplicationRepository,
	resolver *fedinfra.Resolver,
	verifier *fedinfra.Verifier,
	rateLimiter fedinfra.RateLimiter,
	events appEvent.Bus,
) *FederationFriendLinkHandler {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &FederationFriendLinkHandler{
		cfgSvc:          cfgSvc,
		instanceRepo:    instanceRepo,
		linkRepo:        linkRepo,
		applicationRepo: applicationRepo,
		resolver:        resolver,
		verifier:        verifier,
		rateLimiter:     rateLimiter,
		events:          events,
	}
}

// RequestFriendLink handles signed federation friendlink requests.
// @Summary 联合友链申请（入站）
// @Tags Federation
// @Accept json
// @Produce json
// @Param request body contract.FederationFriendLinkRequestReq true "友链申请参数"
// @Success 200 {object} contract.FederationFriendLinkResponseResp
// @Router /api/federation/friendlinks/request [post]
func (h *FederationFriendLinkHandler) RequestFriendLink(c *fiber.Ctx) error {
	body := c.Body()
	req, err := parseFederationRequest(c)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求解析失败", err)
	}

	signature, err := h.verifier.VerifyRequest(c.Context(), req, body)
	if err != nil {
		log.Printf("[federation] 入站 友链申请 校验失败 ip=%s err=%v", c.IP(), err)
		_ = h.events.Publish(c.Context(), appEvent.Generic{
			EventName: "federation.signature.verify_failed",
			At:        time.Now(),
			Payload:   map[string]any{"action": "friendlink", "ip": c.IP()},
		})
		return response.NewBizErrorWithMsg(response.Unauthorized, "签名校验失败")
	}

	var payload contract.FederationFriendLinkRequestReq
	if err := json.Unmarshal(body, &payload); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	requesterURL := strings.TrimSpace(payload.RequesterURL)
	if requesterURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "requester_url 不能为空")
	}
	if signature != nil && signature.BaseURL != "" && !sameBaseURL(signature.BaseURL, requesterURL) {
		return response.NewBizErrorWithMsg(response.Unauthorized, "签名来源与请求不一致")
	}

	settings, err := h.cfgSvc.Settings(c.Context())
	if err != nil || !settings.Enabled {
		return response.NewBizErrorWithMsg(response.Unauthorized, "联合未启用")
	}
	if !settings.AllowInbound {
		return response.NewBizErrorWithMsg(response.Unauthorized, "已关闭入站请求")
	}
	if err := enforceFederationInboundRateLimit(c.Context(), h.rateLimiter, requesterURL, "friendlink", settings.RateLimits); err != nil {
		_ = h.events.Publish(c.Context(), appEvent.Generic{
			EventName: "federation.inbound.rate_limited",
			At:        time.Now(),
			Payload:   map[string]any{"action": "friendlink", "source": requesterURL},
		})
		return err
	}

	manifest, endpoints, publicKey, err := fetchFederationDocs(c.Context(), h.resolver, requesterURL)
	if err != nil {
		return err
	}

	instance, err := ensureInstanceFromDocs(c.Context(), requesterURL, manifest, endpoints, publicKey, h.instanceRepo)
	if err != nil {
		return err
	}

	app, created, err := h.upsertFriendLinkApplication(c.Context(), payload, requesterURL, manifest, signature.KeyID)
	if err != nil {
		return err
	}

	if autoApproveFriendlink(settings) {
		if err := h.ensureFriendLink(c.Context(), instance, payload.RSSURL); err != nil {
			return err
		}
		app.Status = social.FriendLinkAppStatusApproved
		_ = h.applicationRepo.Update(c.Context(), app)
		instance.Status = "active"
		_ = h.instanceRepo.Update(c.Context(), instance)
	}

	message := "友链申请已提交"
	if app.Status == social.FriendLinkAppStatusApproved {
		message = "友链申请已通过"
	} else if !created {
		message = "友链申请已更新"
	}

	resp := contract.FederationFriendLinkResponseResp{
		ApplicationID: app.ID,
		Status:        app.Status,
		Message:       message,
	}
	log.Printf("[federation] 入站 友链申请 base=%s app_id=%d status=%s key_id=%s", requesterURL, app.ID, app.Status, signature.KeyID)
	_ = h.events.Publish(c.Context(), appEvent.Generic{
		EventName: "federation.friendlink.received",
		At:        time.Now(),
		Payload: map[string]any{
			"RequesterURL":  requesterURL,
			"ApplicationID": app.ID,
			"Status":        app.Status,
			"KeyID":         signature.KeyID,
		},
	})
	return response.Success(c, resp)
}

func (h *FederationFriendLinkHandler) ensureFriendLink(ctx context.Context, instance *federation.FederationInstance, rssURL string) error {
	if instance == nil {
		return nil
	}
	_, err := h.linkRepo.FindByURL(ctx, instance.BaseURL)
	if err == nil {
		return nil
	}
	if !errors.Is(err, social.ErrFriendLinkNotFound) {
		return err
	}
	link := &social.FriendLink{
		Name:             safeString(instance.Name, instance.BaseURL),
		URL:              instance.BaseURL,
		Description:      instance.Description,
		RSSURL:           toOptionalString(rssURL),
		Kind:             "federation",
		SyncMode:         "federation",
		InstanceID:       &instance.ID,
		IsActive:         true,
		TotalPostsCached: 0,
	}
	return h.linkRepo.Create(ctx, link)
}

func (h *FederationFriendLinkHandler) upsertFriendLinkApplication(ctx context.Context, payload contract.FederationFriendLinkRequestReq, requesterURL string, manifest *fedinfra.Manifest, keyID string) (*social.FriendLinkApplication, bool, error) {
	url := strings.TrimSpace(payload.RequesterURL)
	app, err := h.applicationRepo.FindByURL(ctx, url)
	if err != nil && !errors.Is(err, social.ErrFriendLinkApplicationNotFound) {
		return nil, false, err
	}
	if app != nil && app.Status == social.FriendLinkAppStatusBlocked {
		return nil, false, response.NewBizErrorWithMsg(response.Unauthorized, "已被封禁")
	}

	manifestPayload := toJSON(manifest)
	if app == nil {
		app = &social.FriendLinkApplication{
			Name:              toOptionalString(manifest.Instance.Name),
			URL:               url,
			Description:       toOptionalString(manifest.Instance.Description),
			ApplyChannel:      social.FriendLinkApplyChannelFederation,
			RequestedSyncMode: social.FriendLinkSyncModeFederation,
			RSSURL:            toOptionalString(payload.RSSURL),
			InstanceURL:       toOptionalString(requesterURL),
			Manifest:          manifestPayload,
			SignatureKeyID:    toOptionalString(keyID),
			SignatureVerified: true,
			Status:            social.FriendLinkAppStatusPending,
		}
		if err := h.applicationRepo.Create(ctx, app); err != nil {
			return nil, false, err
		}
		return app, true, nil
	}

	app.Name = toOptionalString(manifest.Instance.Name)
	app.Description = toOptionalString(manifest.Instance.Description)
	app.ApplyChannel = social.FriendLinkApplyChannelFederation
	app.RequestedSyncMode = social.FriendLinkSyncModeFederation
	app.RSSURL = toOptionalString(payload.RSSURL)
	app.InstanceURL = toOptionalString(requesterURL)
	app.Manifest = manifestPayload
	app.SignatureKeyID = toOptionalString(keyID)
	app.SignatureVerified = true
	app.Status = social.FriendLinkAppStatusPending
	if err := h.applicationRepo.Update(ctx, app); err != nil {
		return nil, false, err
	}
	return app, false, nil
}

func safeString(val *string, fallback string) string {
	if val == nil || strings.TrimSpace(*val) == "" {
		return fallback
	}
	return *val
}

func autoApproveFriendlink(settings federationconfig.Settings) bool {
	policy := parseFederationPolicy(settings)
	return policyBool(policy.AutoApproveFriendlink, false)
}
