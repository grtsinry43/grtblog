package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	applike "github.com/grtsinry43/grtblog-v2/server/internal/app/like"
	domainlike "github.com/grtsinry43/grtblog-v2/server/internal/domain/like"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type LikeHandler struct {
	svc *applike.Service
}

func NewLikeHandler(svc *applike.Service) *LikeHandler {
	return &LikeHandler{svc: svc}
}

type TrackLikeEnvelope struct {
	Code   int                    `json:"code"`
	BizErr string                 `json:"bizErr"`
	Msg    string                 `json:"msg"`
	Data   contract.TrackLikeResp `json:"data"`
	Meta   response.Meta          `json:"meta"`
}

// TrackLike godoc
// @Summary 点赞埋点（去重）
// @Tags Analytics
// @Accept json
// @Produce json
// @Param request body contract.TrackLikeReq true "点赞参数"
// @Success 200 {object} TrackLikeEnvelope
// @Router /public/analytics/like [post]
func (h *LikeHandler) TrackLike(c *fiber.Ctx) error {
	var req contract.TrackLikeReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	result, err := h.svc.TrackLike(c.UserContext(), applike.TrackLikeCmd{
		ContentType: req.ContentType,
		ContentID:   req.ContentID,
		VisitorID:   req.VisitorID,
	}, applike.RequestMeta{
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent", ""),
	})
	if err != nil {
		return h.mapLikeError(err)
	}

	return response.Success(c, contract.TrackLikeResp{
		VisitorID: result.VisitorID,
		Affected:  result.Affected,
	})
}

func (h *LikeHandler) mapLikeError(err error) error {
	switch {
	case errors.Is(err, domainlike.ErrInvalidTargetType), errors.Is(err, domainlike.ErrInvalidTargetID):
		return response.NewBizErrorWithCause(response.ParamsError, "点赞参数无效", err)
	case errors.Is(err, domainlike.ErrTargetNotFound):
		return response.NewBizErrorWithCause(response.NotFound, "点赞目标不存在", err)
	default:
		return err
	}
}
