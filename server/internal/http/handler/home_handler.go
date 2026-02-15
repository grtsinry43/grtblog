package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/home"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type HomeHandler struct {
	svc *home.Service
}

func NewHomeHandler(svc *home.Service) *HomeHandler {
	return &HomeHandler{svc: svc}
}

// GetActivityPulse godoc
// @Summary 获取首页创作律动数据
// @Tags Home
// @Produce json
// @Param days query string false "天数（默认365，最大730，传 all 返回全量）"
// @Success 200 {object} contract.GenericMessageEnvelope
// @Router /public/home/activity-pulse [get]
func (h *HomeHandler) GetActivityPulse(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "home service 未初始化")
	}
	daysQuery := strings.TrimSpace(c.Query("days"))
	days := 365
	if strings.EqualFold(daysQuery, "all") {
		days = -1
	} else {
		days = parseIntQuery(c, "days", 365)
	}
	result, err := h.svc.GetActivityPulse(c.UserContext(), days)
	if err != nil {
		return err
	}
	return response.Success(c, result)
}

// GetInspirationStats godoc
// @Summary 获取首页灵感模块统计数据
// @Tags Home
// @Produce json
// @Param githubUsername query string false "GitHub 用户名（可选）"
// @Success 200 {object} contract.GenericMessageEnvelope
// @Router /public/home/inspiration-stats [get]
func (h *HomeHandler) GetInspirationStats(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "home service 未初始化")
	}
	githubUsername := strings.TrimSpace(c.Query("githubUsername"))
	result, err := h.svc.GetInspirationStats(c.UserContext(), githubUsername)
	if err != nil {
		return err
	}
	return response.Success(c, result)
}

// GetTimelineByYear godoc
// @Summary 获取时间轴按年份聚合数据
// @Tags Home
// @Produce json
// @Success 200 {object} contract.GenericMessageEnvelope
// @Router /public/home/timeline-by-year [get]
func (h *HomeHandler) GetTimelineByYear(c *fiber.Ctx) error {
	if h == nil || h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "home service 未初始化")
	}
	result, err := h.svc.GetTimelineByYear(c.UserContext())
	if err != nil {
		return err
	}
	return response.Success(c, result)
}
