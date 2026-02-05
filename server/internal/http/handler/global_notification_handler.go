package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/globalnotification"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type GlobalNotificationHandler struct {
	svc *globalnotification.Service
}

func NewGlobalNotificationHandler(svc *globalnotification.Service) *GlobalNotificationHandler {
	return &GlobalNotificationHandler{svc: svc}
}

// ListAdmin godoc
// @Summary 获取全站通知列表（管理端）
// @Tags GlobalNotification
// @Produce json
// @Param status query string false "状态过滤: active|upcoming|expired"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} contract.GlobalNotificationListResp
// @Security BearerAuth
// @Router /admin/global-notifications [get]
// @Security JWTAuth
func (h *GlobalNotificationHandler) ListAdmin(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "pageSize", 10)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	items, total, err := h.svc.List(c.Context(), globalnotification.ListOptions{
		Status:   strings.TrimSpace(c.Query("status")),
		Page:     page,
		PageSize: size,
	})
	if err != nil {
		return err
	}
	respItems := make([]contract.GlobalNotificationResp, len(items))
	for i, item := range items {
		respItems[i] = contract.ToGlobalNotificationResp(item)
	}
	return response.Success(c, contract.GlobalNotificationListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// GetAdmin godoc
// @Summary 获取全站通知详情（管理端）
// @Tags GlobalNotification
// @Produce json
// @Param id path int64 true "通知ID"
// @Success 200 {object} contract.GlobalNotificationResp
// @Security BearerAuth
// @Router /admin/global-notifications/{id} [get]
// @Security JWTAuth
func (h *GlobalNotificationHandler) GetAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的通知ID")
	}
	item, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		return h.mapGlobalNotificationError(err)
	}
	return response.Success(c, contract.ToGlobalNotificationResp(*item))
}

// Create godoc
// @Summary 创建全站通知
// @Tags GlobalNotification
// @Accept json
// @Produce json
// @Param request body contract.GlobalNotificationCreateReq true "创建参数"
// @Success 200 {object} contract.GlobalNotificationResp
// @Security BearerAuth
// @Router /admin/global-notifications [post]
// @Security JWTAuth
func (h *GlobalNotificationHandler) Create(c *fiber.Ctx) error {
	var req contract.GlobalNotificationCreateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	created, err := h.svc.Create(c.Context(), globalnotification.CreateCmd{
		Content:    req.Content,
		PublishAt:  req.PublishAt,
		ExpireAt:   req.ExpireAt,
		AllowClose: req.AllowClose,
	})
	if err != nil {
		return h.mapGlobalNotificationError(err)
	}
	Audit(c, "global_notification.create", map[string]any{"id": created.ID})
	return response.SuccessWithMessage(c, contract.ToGlobalNotificationResp(*created), "全站通知创建成功")
}

// Update godoc
// @Summary 更新全站通知
// @Tags GlobalNotification
// @Accept json
// @Produce json
// @Param id path int64 true "通知ID"
// @Param request body contract.GlobalNotificationUpdateReq true "更新参数"
// @Success 200 {object} contract.GlobalNotificationResp
// @Security BearerAuth
// @Router /admin/global-notifications/{id} [put]
// @Security JWTAuth
func (h *GlobalNotificationHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的通知ID")
	}
	var req contract.GlobalNotificationUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	updated, err := h.svc.Update(c.Context(), globalnotification.UpdateCmd{
		ID:         id,
		Content:    req.Content,
		PublishAt:  req.PublishAt,
		ExpireAt:   req.ExpireAt,
		AllowClose: req.AllowClose,
	})
	if err != nil {
		return h.mapGlobalNotificationError(err)
	}
	Audit(c, "global_notification.update", map[string]any{"id": updated.ID})
	return response.SuccessWithMessage(c, contract.ToGlobalNotificationResp(*updated), "全站通知更新成功")
}

// Delete godoc
// @Summary 删除全站通知
// @Tags GlobalNotification
// @Produce json
// @Param id path int64 true "通知ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /admin/global-notifications/{id} [delete]
// @Security JWTAuth
func (h *GlobalNotificationHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的通知ID")
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return h.mapGlobalNotificationError(err)
	}
	Audit(c, "global_notification.delete", map[string]any{"id": id})
	return response.SuccessWithMessage[any](c, nil, "全站通知删除成功")
}

// ListPublicActive godoc
// @Summary 公开获取当前生效的全站通知
// @Tags GlobalNotification
// @Produce json
// @Success 200 {object} []contract.GlobalNotificationResp
// @Router /public/global-notifications [get]
func (h *GlobalNotificationHandler) ListPublicActive(c *fiber.Ctx) error {
	items, err := h.svc.ListActive(c.Context())
	if err != nil {
		return err
	}
	respItems := make([]contract.GlobalNotificationResp, len(items))
	for i, item := range items {
		respItems[i] = contract.ToGlobalNotificationResp(item)
	}
	return response.Success(c, respItems)
}

func (h *GlobalNotificationHandler) mapGlobalNotificationError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, globalnotification.ErrContentRequired),
		errors.Is(err, globalnotification.ErrInvalidPublishWindow),
		errors.Is(err, globalnotification.ErrInvalidNotificationID):
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	case errors.Is(err, social.ErrGlobalNotificationNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, err.Error())
	default:
		return err
	}
}
