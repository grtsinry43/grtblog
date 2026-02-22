package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/ownerstatus"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type OwnerStatusHandler struct {
	svc *ownerstatus.Service
}

func NewOwnerStatusHandler(svc *ownerstatus.Service) *OwnerStatusHandler {
	return &OwnerStatusHandler{svc: svc}
}

func (h *OwnerStatusHandler) GetStatus(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "站长状态服务未初始化")
	}
	return response.Success(c, h.svc.Get())
}

func (h *OwnerStatusHandler) UpdateStatus(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "站长状态服务未初始化")
	}

	var req ownerstatus.UpdateInput
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
		}
	}

	if req.OK != nil && *req.OK != 0 && *req.OK != 1 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ok 仅支持 0 或 1")
	}
	if req.Timestamp != nil && *req.Timestamp <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "timestamp 必须大于 0")
	}

	next := h.svc.Update(req)
	Audit(c, "owner.status.update", map[string]any{
		"ok":               next.OK,
		"process":          next.Process,
		"adminPanelOnline": next.AdminPanelOnline,
	})
	return response.Success(c, next)
}

func (h *OwnerStatusHandler) PanelHeartbeat(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "站长状态服务未初始化")
	}

	status := h.svc.TouchAdminPanel()
	Audit(c, "owner.status.panel_heartbeat", map[string]any{
		"adminPanelOnline": status.AdminPanelOnline,
	})
	return response.Success(c, status)
}
