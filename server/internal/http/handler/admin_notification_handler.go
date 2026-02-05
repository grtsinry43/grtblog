package handler

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AdminNotificationHandler struct {
	svc *adminnotification.Service
}

func NewAdminNotificationHandler(svc *adminnotification.Service) *AdminNotificationHandler {
	return &AdminNotificationHandler{svc: svc}
}

// ListMine 查询当前用户站内信。
// @Summary 查询我的站内信
// @Tags AdminNotification
// @Produce json
// @Param unreadOnly query bool false "是否仅未读"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.AdminNotificationListResp
// @Security BearerAuth
// @Router /notifications [get]
// @Security JWTAuth
func (h *AdminNotificationHandler) ListMine(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.NewBizError(response.NotLogin)
	}
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "pageSize", 20)
	unreadOnly := c.QueryBool("unreadOnly", false)
	items, total, err := h.svc.ListByUser(c.Context(), claims.UserID, unreadOnly, page, size)
	if err != nil {
		return h.mapError(err)
	}
	respItems := make([]contract.AdminNotificationResp, len(items))
	for i := range items {
		respItems[i] = toAdminNotificationResp(items[i])
	}
	return response.Success(c, contract.AdminNotificationListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// MarkRead 标记单条已读。
// @Summary 标记站内信已读
// @Tags AdminNotification
// @Produce json
// @Param id path int true "通知ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /notifications/{id}/read [post]
// @Security JWTAuth
func (h *AdminNotificationHandler) MarkRead(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.NewBizError(response.NotLogin)
	}
	id := parseInt64Path(c, "id")
	if id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的通知ID")
	}
	if err := h.svc.MarkRead(c.Context(), claims.UserID, id); err != nil {
		return h.mapError(err)
	}
	return response.SuccessWithMessage[any](c, nil, "已标记为已读")
}

// MarkAllRead 标记全部已读。
// @Summary 标记全部站内信已读
// @Tags AdminNotification
// @Produce json
// @Success 200 {object} any
// @Security BearerAuth
// @Router /notifications/read-all [post]
// @Security JWTAuth
func (h *AdminNotificationHandler) MarkAllRead(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.NewBizError(response.NotLogin)
	}
	if err := h.svc.MarkAllRead(c.Context(), claims.UserID); err != nil {
		return h.mapError(err)
	}
	return response.SuccessWithMessage[any](c, nil, "全部标记已读")
}

func (h *AdminNotificationHandler) mapError(err error) error {
	switch err {
	case adminnotification.ErrInvalidNotification:
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	case social.ErrAdminNotificationNotFound:
		return response.NewBizErrorWithMsg(response.NotFound, err.Error())
	default:
		return err
	}
}

func toAdminNotificationResp(item social.AdminNotification) contract.AdminNotificationResp {
	var payload any
	_ = json.Unmarshal(item.Payload, &payload)
	return contract.AdminNotificationResp{
		ID:        item.ID,
		Type:      item.NotifType,
		Title:     item.Title,
		Content:   item.Content,
		Payload:   payload,
		IsRead:    item.IsRead,
		ReadAt:    timePtrToString(item.ReadAt),
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func timePtrToString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	v := t.UTC().Format(time.RFC3339)
	return &v
}

func parseInt64Path(c *fiber.Ctx, name string) int64 {
	val, _ := c.ParamsInt(name, 0)
	return int64(val)
}
