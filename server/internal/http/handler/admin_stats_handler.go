package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminstats"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AdminStatsHandler struct {
	svc *adminstats.Service
}

func NewAdminStatsHandler(svc *adminstats.Service) *AdminStatsHandler {
	return &AdminStatsHandler{svc: svc}
}

type AdminDashboardStatsEnvelope struct {
	Code   int                       `json:"code"`
	BizErr string                    `json:"bizErr"`
	Msg    string                    `json:"msg"`
	Data   adminstats.DashboardStats `json:"data"`
	Meta   response.Meta             `json:"meta"`
}

// GetDashboard godoc
// @Summary 获取后台仪表盘统计
// @Description 获取后台总览、互动总量、趋势、分布与热门内容等统计数据（含 Redis 缓存）。
// @Tags Admin-Stats
// @Produce json
// @Success 200 {object} AdminDashboardStatsEnvelope
// @Security BearerAuth
// @Router /admin/stats/dashboard [get]
func (h *AdminStatsHandler) GetDashboard(c *fiber.Ctx) error {
	stats, err := h.svc.GetDashboardStats(c.UserContext())
	if err != nil {
		return err
	}
	return response.Success(c, stats)
}
