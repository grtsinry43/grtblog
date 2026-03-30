package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/telemetry"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

// AdminTelemetryHandler exposes error telemetry data to the admin dashboard.
type AdminTelemetryHandler struct {
	collector *telemetry.Collector
}

func NewAdminTelemetryHandler(collector *telemetry.Collector) *AdminTelemetryHandler {
	return &AdminTelemetryHandler{collector: collector}
}

// GetSnapshot returns the full telemetry snapshot for audit / preview.
// GET /api/v2/admin/telemetry/snapshot
func (h *AdminTelemetryHandler) GetSnapshot(c *fiber.Ctx) error {
	if h == nil || h.collector == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry collector 未初始化")
	}
	snap := telemetry.BuildSnapshot(h.collector)
	return response.Success(c, snap)
}

// GetStats returns lightweight summary numbers.
// GET /api/v2/admin/telemetry/stats
func (h *AdminTelemetryHandler) GetStats(c *fiber.Ctx) error {
	if h == nil || h.collector == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry collector 未初始化")
	}
	unique, total := h.collector.Stats()
	return response.Success(c, fiber.Map{
		"uniqueErrors": unique,
		"totalCount":   total,
	})
}

// ResetErrors clears all collected error digests.
// POST /api/v2/admin/telemetry/reset
func (h *AdminTelemetryHandler) ResetErrors(c *fiber.Ctx) error {
	if h == nil || h.collector == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "telemetry collector 未初始化")
	}
	h.collector.Reset()
	return response.SuccessWithMessage[any](c, nil, "error telemetry reset")
}
