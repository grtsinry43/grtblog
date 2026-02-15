package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/observability"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AdminObservabilityHandler struct {
	svc *observability.Service
}

func NewAdminObservabilityHandler(svc *observability.Service) *AdminObservabilityHandler {
	return &AdminObservabilityHandler{svc: svc}
}

func (h *AdminObservabilityHandler) GetOverview(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	data, err := h.svc.GetOverview(c.UserContext())
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetControlPlane(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	window := parseWindowParam(c.Query("window"), 5*time.Minute)
	data, err := h.svc.GetControlPlane(c.UserContext(), window)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetRenderPlane(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	data, err := h.svc.GetRenderPlane(c.UserContext())
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetRealtime(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	data, err := h.svc.GetRealtime(c.UserContext())
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetFederation(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	window := parseWindowParam(c.Query("window"), 24*time.Hour)
	data, err := h.svc.GetFederation(c.UserContext(), window)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetStorage(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	data, err := h.svc.GetStorage(c.UserContext())
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetTimeline(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	since, _ := time.Parse(time.RFC3339, c.Query("since"))
	until, _ := time.Parse(time.RFC3339, c.Query("until"))
	groupBy := c.Query("group_by")
	data, err := h.svc.GetTimeline(c.UserContext(), since, until, groupBy)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "timeline 参数错误", err)
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetAlerts(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	limit := c.QueryInt("limit", 50)
	data, err := h.svc.GetAlerts(c.UserContext(), limit)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) GetPageState(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	trackedLimit := c.QueryInt("tracked_limit", 200)
	recentLimit := c.QueryInt("recent_limit", 30)
	routeLimit := c.QueryInt("route_limit", 500)
	data, err := h.svc.GetPageState(c.UserContext(), trackedLimit, recentLimit, routeLimit)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) BootstrapPages(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	data, err := h.svc.BootstrapPages(c.UserContext())
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func (h *AdminObservabilityHandler) InvalidatePages(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "observability service 未初始化")
	}
	var req observability.PageInvalidateRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return response.NewBizErrorWithCause(response.ParamsError, "invalid invalidate payload", err)
		}
	}
	data, err := h.svc.InvalidatePages(c.UserContext(), req)
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func parseWindowParam(raw string, fallback time.Duration) time.Duration {
	if raw == "" {
		return fallback
	}
	window, err := time.ParseDuration(raw)
	if err != nil || window <= 0 {
		return fallback
	}
	return window
}
