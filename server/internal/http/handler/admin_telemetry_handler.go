package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/telemetry"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

// AdminTelemetryHandler exposes error telemetry data to the admin dashboard.
type AdminTelemetryHandler struct {
	svc *telemetry.Service
}

func NewAdminTelemetryHandler(svc *telemetry.Service) *AdminTelemetryHandler {
	return &AdminTelemetryHandler{svc: svc}
}

// GetSnapshot returns the full telemetry snapshot (environment + metrics + errors).
// GET /api/v2/admin/telemetry/snapshot
func (h *AdminTelemetryHandler) GetSnapshot(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry service 未初始化")
	}
	snap := h.svc.FullSnapshot(c.UserContext())
	return response.Success(c, snap)
}

// GetStats returns lightweight summary numbers.
// GET /api/v2/admin/telemetry/stats
func (h *AdminTelemetryHandler) GetStats(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry service 未初始化")
	}
	collector := h.svc.Collector()
	unique, total := collector.Stats()
	return response.Success(c, fiber.Map{
		"uniqueErrors": unique,
		"totalCount":   total,
	})
}

// ResetErrors clears all collected error digests.
// POST /api/v2/admin/telemetry/reset
func (h *AdminTelemetryHandler) ResetErrors(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry service 未初始化")
	}
	h.svc.Collector().Reset()
	return response.SuccessWithMessage[any](c, nil, "error telemetry reset")
}

// GetReportHistory returns recent upload attempt records.
// GET /api/v2/admin/telemetry/report-history
func (h *AdminTelemetryHandler) GetReportHistory(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry service 未初始化")
	}
	reporter := h.svc.Reporter()
	return response.Success(c, fiber.Map{
		"history": reporter.History(),
	})
}

// ReportNow triggers an immediate telemetry upload.
// POST /api/v2/admin/telemetry/report-now
func (h *AdminTelemetryHandler) ReportNow(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry service 未初始化")
	}
	reporter := h.svc.Reporter()
	rec := reporter.ReportNow(c.UserContext())
	return response.Success(c, rec)
}
