package handler

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appap "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypub"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type ActivityPubAdminHandler struct {
	svc *appap.Service
}

func NewActivityPubAdminHandler(svc *appap.Service) *ActivityPubAdminHandler {
	return &ActivityPubAdminHandler{svc: svc}
}

// Publish pushes local content to ActivityPub followers.
// @Summary 推送 ActivityPub 时间线
// @Tags ActivityPubAdmin
// @Accept json
// @Produce json
// @Param request body contract.FederationActivityPubPublishReq true "推送参数"
// @Success 200 {object} contract.FederationActivityPubPublishResp
// @Security BearerAuth
// @Router /admin/activitypub/publish [post]
// @Security JWTAuth
func (h *ActivityPubAdminHandler) Publish(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "ActivityPub 服务未初始化")
	}
	var req contract.FederationActivityPubPublishReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	sourceType := strings.ToLower(strings.TrimSpace(req.SourceType))
	switch sourceType {
	case "article", "moment", "thinking":
	default:
		return response.NewBizErrorWithMsg(response.ParamsError, "source_type 仅支持 article/moment/thinking")
	}
	if req.SourceID <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "source_id 必须大于 0")
	}
	baseURL := resolveActivityPubBaseURL(c)
	result, err := h.svc.Publish(c.Context(), baseURL, appap.PublishCmd{
		SourceType: sourceType,
		SourceID:   req.SourceID,
		Summary:    strings.TrimSpace(req.Summary),
	})
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "ActivityPub 推送失败", err)
	}
	return response.Success(c, contract.FederationActivityPubPublishResp{
		ActivityID:   result.Item.ActivityID,
		ObjectID:     result.Item.ObjectID,
		SourceType:   result.Item.SourceType,
		SourceID:     result.Item.SourceID,
		Deliveries:   result.Deliveries,
		SuccessCount: result.SuccessCount,
		FailureCount: result.FailureCount,
		FailedTarget: result.FailedTargets,
		PublishedAt:  result.Item.PublishedAt.UTC().Format(time.RFC3339),
	})
}

// ListFollowers lists stored ActivityPub followers.
// @Summary 查询 ActivityPub 关注者
// @Tags ActivityPubAdmin
// @Produce json
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} contract.FederationActivityPubFollowerListResp
// @Security BearerAuth
// @Router /admin/activitypub/followers [get]
// @Security JWTAuth
func (h *ActivityPubAdminHandler) ListFollowers(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "ActivityPub 服务未初始化")
	}
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "pageSize", 20)
	items, total, err := h.svc.ListFollowers(c.Context(), page, size)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "关注者查询失败", err)
	}
	resp := make([]contract.FederationActivityPubFollowerResp, len(items))
	for i, item := range items {
		var lastSeen *string
		if item.LastSeenAt != nil {
			val := item.LastSeenAt.UTC().Format(time.RFC3339)
			lastSeen = &val
		}
		resp[i] = contract.FederationActivityPubFollowerResp{
			ID:                item.ID,
			ActorID:           item.ActorID,
			InboxURL:          item.InboxURL,
			SharedInboxURL:    item.SharedInboxURL,
			PreferredUsername: item.PreferredUsername,
			DisplayName:       item.DisplayName,
			Status:            item.Status,
			FollowedAt:        item.FollowedAt.UTC().Format(time.RFC3339),
			LastSeenAt:        lastSeen,
			UpdatedAt:         item.UpdatedAt.UTC().Format(time.RFC3339),
		}
	}
	return response.Success(c, contract.FederationActivityPubFollowerListResp{
		Items: resp,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
