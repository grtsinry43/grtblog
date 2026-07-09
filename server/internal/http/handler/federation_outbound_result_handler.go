package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	fedinfra "github.com/grtsinry43/grtblog-v2/server/internal/infra/federation"
)

type FederationOutboundResultHandler struct {
	deliverySvc *appfed.DeliveryService
	verifier    *fedinfra.Verifier
}

func NewFederationOutboundResultHandler(deliverySvc *appfed.DeliveryService, verifier *fedinfra.Verifier) *FederationOutboundResultHandler {
	return &FederationOutboundResultHandler{
		deliverySvc: deliverySvc,
		verifier:    verifier,
	}
}

// ResultCallback 远端回调本地出站结果。
// @Summary 联合出站结果回调
// @Tags Federation
// @Accept json
// @Produce json
// @Param request body contract.FederationOutboundResultReq true "回调结果"
// @Success 200 {object} contract.FederationOutboundResultResp
// @Router /api/federation/outbound/result [post]
func (h *FederationOutboundResultHandler) ResultCallback(c *fiber.Ctx) error {
	if h.deliverySvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "联邦服务未初始化")
	}
	body := c.Body()
	req, err := parseFederationRequest(c)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求解析失败", err)
	}
	callerBaseURL := ""
	if h.verifier != nil {
		signature, err := h.verifier.VerifyRequest(c.Context(), req, body)
		if err != nil {
			return response.NewBizErrorWithMsg(response.Unauthorized, "签名校验失败")
		}
		if signature != nil {
			callerBaseURL = signature.BaseURL
		}
	}

	var payload contract.FederationOutboundResultReq
	if err := c.BodyParser(&payload); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	requestID := strings.TrimSpace(payload.RequestID)
	if requestID == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "request_id 不能为空")
	}
	var processedAt *time.Time
	if raw := strings.TrimSpace(payload.ProcessedAt); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return response.NewBizErrorWithMsg(response.ParamsError, "processed_at 必须为 RFC3339")
		}
		processedAt = &parsed
	}
	item, err := h.deliverySvc.HandleCallback(c.Context(), appfed.CallbackResultCmd{
		RequestID:      requestID,
		Type:           strings.TrimSpace(payload.Type),
		Status:         strings.TrimSpace(payload.Status),
		RemoteTicketID: strings.TrimSpace(payload.RemoteTicketID),
		Reason:         strings.TrimSpace(payload.Reason),
		ProcessedAt:    processedAt,
		CallerBaseURL:  callerBaseURL,
	})
	if err != nil {
		if errors.Is(err, appfed.ErrCallbackSourceMismatch) {
			return response.NewBizErrorWithMsg(response.Unauthorized, "回调来源与投递目标不一致")
		}
		return response.NewBizErrorWithCause(response.ServerError, "回调处理失败", err)
	}
	return response.Success(c, contract.FederationOutboundResultResp{
		RequestID: item.RequestID,
		Status:    item.Status,
	})
}
