package handler

import (
	"github.com/gofiber/fiber/v2"

	apprss "github.com/grtsinry43/grtblog-v2/server/internal/app/rss"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type RSSAdminHandler struct {
	svc *apprss.AccessAnalyticsService
}

func NewRSSAdminHandler(svc *apprss.AccessAnalyticsService) *RSSAdminHandler {
	return &RSSAdminHandler{svc: svc}
}

// GetAccessStats godoc
// @Summary 获取 RSS 访问统计（管理端）
// @Tags RSSAdmin
// @Produce json
// @Param days query int false "统计天数（默认7，最大90）"
// @Param top query int false "Top数量（默认12，最大50）"
// @Success 200 {object} apprss.AccessStats
// @Security BearerAuth
// @Router /admin/rss/access-stats [get]
func (h *RSSAdminHandler) GetAccessStats(c *fiber.Ctx) error {
	days := parseIntQuery(c, "days", 7)
	top := parseIntQuery(c, "top", 12)
	stats, err := h.svc.GetStats(c.UserContext(), days, top)
	if err != nil {
		return err
	}
	return response.Success(c, stats)
}

