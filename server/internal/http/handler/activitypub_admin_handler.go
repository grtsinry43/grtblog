package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appap "github.com/grtsinry43/grtblog-v2/server/internal/app/activitypub"
	domainap "github.com/grtsinry43/grtblog-v2/server/internal/domain/activitypub"
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
		SourceType:    sourceType,
		SourceID:      req.SourceID,
		Summary:       strings.TrimSpace(req.Summary),
		TriggerSource: "manual",
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
	return response.Success(c, contract.FederationActivityPubFollowerListResp{Items: resp, Total: total, Page: page, Size: size})
}

func (h *ActivityPubAdminHandler) ListOutbox(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "ActivityPub 服务未初始化")
	}
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "pageSize", 20)
	opts := domainap.OutboxListOptions{
		Page:       page,
		PageSize:   size,
		Status:     strings.TrimSpace(c.Query("status")),
		SourceType: strings.TrimSpace(c.Query("sourceType")),
		Search:     strings.TrimSpace(c.Query("search")),
	}
	items, total, err := h.svc.ListOutbox(c.Context(), opts)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "ActivityPub 出站查询失败", err)
	}
	resp := make([]contract.ActivityPubOutboxItemResp, 0, len(items))
	for _, item := range items {
		resp = append(resp, mapActivityPubOutboxItemResp(item, false))
	}
	return response.Success(c, contract.ActivityPubOutboxListResp{Items: resp, Total: total, Page: page, Size: size})
}

func (h *ActivityPubAdminHandler) GetOutbox(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "ActivityPub 服务未初始化")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的 outbox id")
	}
	item, err := h.svc.GetOutbox(c.Context(), id)
	if err != nil {
		if errors.Is(err, domainap.ErrOutboxItemNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "ActivityPub 出站记录不存在")
		}
		return response.NewBizErrorWithCause(response.ServerError, "查询 ActivityPub 出站详情失败", err)
	}
	return response.Success(c, mapActivityPubOutboxItemResp(*item, true))
}

func (h *ActivityPubAdminHandler) RetryOutbox(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "ActivityPub 服务未初始化")
	}
	id, err := strconv.ParseInt(strings.TrimSpace(c.Params("id")), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的 outbox id")
	}
	item, err := h.svc.RetryFailedDeliveries(c.Context(), resolveActivityPubBaseURL(c), id)
	if err != nil {
		if errors.Is(err, domainap.ErrOutboxItemNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "ActivityPub 出站记录不存在")
		}
		if errors.Is(err, domainap.ErrOutboxItemNotRetryable) {
			return response.NewBizErrorWithMsg(response.ParamsError, "当前状态不允许重试")
		}
		return response.NewBizErrorWithCause(response.ServerError, "重试 ActivityPub 出站失败", err)
	}
	return response.Success(c, mapActivityPubOutboxItemResp(*item, true))
}

func mapActivityPubOutboxItemResp(item domainap.OutboxItem, withDeliveries bool) contract.ActivityPubOutboxItemResp {
	resp := contract.ActivityPubOutboxItemResp{
		ID:            item.ID,
		ActivityID:    item.ActivityID,
		ObjectID:      item.ObjectID,
		SourceType:    item.SourceType,
		SourceID:      item.SourceID,
		SourceURL:     item.SourceURL,
		Summary:       item.Summary,
		Status:        item.Status,
		TriggerSource: item.TriggerSource,
		TotalTargets:  item.TotalTargets,
		SuccessCount:  item.SuccessCount,
		FailureCount:  item.FailureCount,
		DurationMs:    item.DurationMs,
		PublishedAt:   item.PublishedAt.UTC().Format(time.RFC3339),
		CreatedAt:     item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     item.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if item.StartedAt != nil {
		v := item.StartedAt.UTC().Format(time.RFC3339)
		resp.StartedAt = &v
	}
	if item.FinishedAt != nil {
		v := item.FinishedAt.UTC().Format(time.RFC3339)
		resp.FinishedAt = &v
	}
	if withDeliveries {
		resp.Activity = string(item.Activity)
		resp.Deliveries = make([]contract.ActivityPubDeliveryDetailResp, 0, len(item.Deliveries))
		for _, d := range item.Deliveries {
			var deliveredAt *string
			if d.DeliveredAt != nil {
				v := d.DeliveredAt.UTC().Format(time.RFC3339)
				deliveredAt = &v
			}
			resp.Deliveries = append(resp.Deliveries, contract.ActivityPubDeliveryDetailResp{
				Inbox:       d.Inbox,
				ActorID:     d.ActorID,
				Status:      d.Status,
				HTTPStatus:  d.HTTPStatus,
				Error:       d.Error,
				DeliveredAt: deliveredAt,
			})
		}
	}
	return resp
}
