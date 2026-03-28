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
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
)

type FederationMentionHandler struct {
	cfgSvc       *sysconfig.Service
	instanceRepo federation.FederationInstanceRepository
	mentionRepo  federation.FederatedMentionRepository
	linkRepo     social.FriendLinkRepository
	userRepo     identity.Repository
	resolver     *fedinfra.Resolver
	verifier     *fedinfra.Verifier
	rateLimiter  fedinfra.RateLimiter
	events       appEvent.Bus
}

func NewFederationMentionHandler(
	cfgSvc *sysconfig.Service,
	instanceRepo federation.FederationInstanceRepository,
	mentionRepo federation.FederatedMentionRepository,
	linkRepo social.FriendLinkRepository,
	userRepo identity.Repository,
	resolver *fedinfra.Resolver,
	verifier *fedinfra.Verifier,
	rateLimiter fedinfra.RateLimiter,
	events appEvent.Bus,
) *FederationMentionHandler {
	if events == nil {
		events = appEvent.NopBus{}
	}
	return &FederationMentionHandler{
		cfgSvc:       cfgSvc,
		instanceRepo: instanceRepo,
		mentionRepo:  mentionRepo,
		linkRepo:     linkRepo,
		userRepo:     userRepo,
		resolver:     resolver,
		verifier:     verifier,
		rateLimiter:  rateLimiter,
		events:       events,
	}
}

// NotifyMention handles cross-site mention notifications.
// @Summary 联合提及通知（入站）
// @Tags Federation
// @Accept json
// @Produce json
// @Param request body contract.FederationMentionNotifyReq true "提及通知参数"
// @Success 200 {object} contract.FederationMentionNotifyResp
// @Router /api/federation/mentions/notify [post]
func (h *FederationMentionHandler) NotifyMention(c *fiber.Ctx) error {
	body := c.Body()
	req, err := parseFederationRequest(c)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求解析失败", err)
	}

	signature, err := h.verifier.VerifyRequest(c.Context(), req, body)
	if err != nil {
		log.Printf("[federation] 入站 提及通知 校验失败 ip=%s err=%v", c.IP(), err)
		_ = h.events.Publish(c.Context(), appEvent.Generic{
			EventName: "federation.signature.verify_failed",
			At:        time.Now(),
			Payload:   map[string]any{"action": "mention", "ip": c.IP()},
		})
		return response.NewBizErrorWithMsg(response.Unauthorized, "签名校验失败")
	}

	var payload contract.FederationMentionNotifyReq
	if err := json.Unmarshal(body, &payload); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(payload.SourceInstanceURL) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "source_instance_url 不能为空")
	}
	if strings.TrimSpace(payload.SourcePost.URL) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "source_post.url 不能为空")
	}
	if strings.TrimSpace(payload.MentionedUser) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "mentioned_user 不能为空")
	}
	if strings.TrimSpace(payload.MentionContext) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "mention_context 不能为空")
	}
	if signature != nil && signature.BaseURL != "" && !sameBaseURL(signature.BaseURL, payload.SourceInstanceURL) {
		return response.NewBizErrorWithMsg(response.Unauthorized, "签名来源与请求不一致")
	}

	settings, err := h.cfgSvc.FederationSettings(c.Context())
	if err != nil || !settings.Enabled {
		return response.NewBizErrorWithMsg(response.Unauthorized, "联合未启用")
	}
	policy := parseFederationPolicy(settings)
	if !policyBool(policy.AllowMention, true) {
		return response.NewBizErrorWithMsg(response.Unauthorized, "未允许被提及")
	}
	if !settings.AllowInbound {
		return response.NewBizErrorWithMsg(response.Unauthorized, "已关闭入站请求")
	}
	if err := enforceFederationInboundRateLimit(c.Context(), h.rateLimiter, payload.SourceInstanceURL, "mention", settings.RateLimits); err != nil {
		_ = h.events.Publish(c.Context(), appEvent.Generic{
			EventName: "federation.inbound.rate_limited",
			At:        time.Now(),
			Payload:   map[string]any{"action": "mention", "source": payload.SourceInstanceURL},
		})
		return err
	}

	user, err := h.userRepo.FindByUsername(c.Context(), payload.MentionedUser)
	if err != nil {
		if errors.Is(err, identity.ErrUserNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "用户不存在")
		}
		return response.NewBizErrorWithCause(response.ServerError, "用户查询失败", err)
	}

	instance, err := ensureFederationInstance(c.Context(), payload.SourceInstanceURL, h.resolver, h.instanceRepo)
	if err != nil {
		return err
	}

	mentionType := strings.TrimSpace(payload.MentionType)
	if mentionType == "" {
		mentionType = "discussion"
	}
	status := "pending"
	if policyBool(policy.AutoApproveFriendlinkCitation, false) && h.isFriendLink(c.Context(), payload.SourceInstanceURL) {
		status = "approved"
	}

	// Idempotency: skip if we already processed this request.
	if reqID := strings.TrimSpace(payload.RequestID); reqID != "" {
		if existing, err := h.mentionRepo.FindBySourceRequestID(c.Context(), reqID); err == nil && existing != nil {
			return response.Success(c, contract.FederationMentionNotifyResp{
				MentionID: existing.ID,
				Delivered: existing.Status == "approved",
			})
		}
	}

	mention := &federation.FederatedMention{
		SourceInstanceID: instance.ID,
		SourceRequestID:  toOptionalString(payload.RequestID),
		SourcePostURL:    payload.SourcePost.URL,
		SourcePostTitle:  toOptionalString(payload.SourcePost.Title),
		MentionedUserID:  user.ID,
		MentionContext:   payload.MentionContext,
		MentionType:      mentionType,
		Status:           status,
		IsRead:           false,
		CreatedAt:        time.Now().UTC(),
	}
	if err := h.mentionRepo.Create(c.Context(), mention); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "写入提及失败", err)
	}

	resp := contract.FederationMentionNotifyResp{
		MentionID: mention.ID,
		Delivered: status == "approved",
	}
	log.Printf("[federation] 入站 提及通知 source=%s mentioned=%s mention_id=%d key_id=%s", payload.SourceInstanceURL, payload.MentionedUser, mention.ID, signature.KeyID)
	_ = h.events.Publish(c.Context(), appEvent.Generic{
		EventName: "federation.mention.received",
		At:        time.Now(),
		Payload: map[string]any{
			"MentionID":         mention.ID,
			"SourceInstanceURL": payload.SourceInstanceURL,
			"MentionedUser":     payload.MentionedUser,
			"Status":            status,
			"KeyID":             signature.KeyID,
		},
	})
	return response.Success(c, resp)
}

func (h *FederationMentionHandler) isFriendLink(ctx context.Context, baseURL string) bool {
	if h.linkRepo == nil {
		return false
	}
	_, err := h.linkRepo.FindByURL(ctx, strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	return err == nil
}
