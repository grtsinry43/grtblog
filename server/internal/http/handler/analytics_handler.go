package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/analytics"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AnalyticsHandler struct {
	svc *analytics.Service
}

func NewAnalyticsHandler(svc *analytics.Service) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

type TrackViewEnvelope struct {
	Code   int                    `json:"code"`
	BizErr string                 `json:"bizErr"`
	Msg    string                 `json:"msg"`
	Data   contract.TrackViewResp `json:"data"`
	Meta   response.Meta          `json:"meta"`
}

// TrackView godoc
// @Summary 记录详情页访问埋点
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body contract.TrackViewReq true "埋点参数"
// @Success 200 {object} TrackViewEnvelope
// @Router /public/analytics/view [post]
func (h *AnalyticsHandler) TrackView(c *fiber.Ctx) error {
	var req contract.TrackViewReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	res, err := h.svc.TrackView(c.UserContext(), analytics.ViewTrackInput{
		ContentType: req.ContentType,
		ContentID:   req.ContentID,
		VisitorID:   req.VisitorID,
		IP:          c.IP(),
		UserAgent:   c.Get("User-Agent", ""),
	})
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "埋点请求无效", err)
	}

	return response.Success(c, contract.TrackViewResp{
		VisitorID: res.VisitorID,
		Queued:    res.Queued,
	})
}
