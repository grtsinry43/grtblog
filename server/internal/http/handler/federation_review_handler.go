package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appfed "github.com/grtsinry43/grtblog-v2/server/internal/app/federation"
	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type FederationReviewHandler struct {
	citationRepo domainfed.FederatedCitationRepository
	mentionRepo  domainfed.FederatedMentionRepository
	instanceRepo domainfed.FederationInstanceRepository
	outbound     *appfed.OutboundService
}

func NewFederationReviewHandler(
	citationRepo domainfed.FederatedCitationRepository,
	mentionRepo domainfed.FederatedMentionRepository,
	instanceRepo domainfed.FederationInstanceRepository,
	outbound *appfed.OutboundService,
) *FederationReviewHandler {
	return &FederationReviewHandler{
		citationRepo: citationRepo,
		mentionRepo:  mentionRepo,
		instanceRepo: instanceRepo,
		outbound:     outbound,
	}
}

// ListPendingReviews 查询待审核项。
// @Summary 联合待审核列表
// @Tags FederationAdmin
// @Produce json
// @Success 200 {object} contract.FederationReviewListResp
// @Security BearerAuth
// @Router /admin/federation/reviews/pending [get]
// @Security JWTAuth
func (h *FederationReviewHandler) ListPendingReviews(c *fiber.Ctx) error {
	citations, err := h.citationRepo.List(c.Context(), "pending", 100)
	if err != nil {
		return err
	}
	mentions, err := h.mentionRepo.List(c.Context(), "pending", 100)
	if err != nil {
		return err
	}
	items := make([]contract.FederationReviewItemResp, 0, len(citations)+len(mentions))
	for _, item := range citations {
		items = append(items, contract.FederationReviewItemResp{
			Type:             "citation",
			ID:               item.ID,
			Status:           item.Status,
			SourceInstanceID: item.SourceInstanceID,
			SourceRequestID:  item.SourceRequestID,
			Summary:          item.SourcePostURL,
			RequestedAt:      item.RequestedAt.UTC().Format(time.RFC3339),
		})
	}
	for _, item := range mentions {
		items = append(items, contract.FederationReviewItemResp{
			Type:             "mention",
			ID:               item.ID,
			Status:           item.Status,
			SourceInstanceID: item.SourceInstanceID,
			SourceRequestID:  item.SourceRequestID,
			Summary:          item.MentionContext,
			RequestedAt:      item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return response.Success(c, contract.FederationReviewListResp{Items: items})
}

// ReviewCitation 审核引用。
// @Summary 审核联合引用
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param id path int true "引用ID"
// @Param request body contract.FederationReviewDecisionReq true "审核结果"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /admin/federation/citations/{id}/review [put]
// @Security JWTAuth
func (h *FederationReviewHandler) ReviewCitation(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的引用ID")
	}
	var req contract.FederationReviewDecisionReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	status := normalizeReviewStatus(req.Status)
	if status == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "status 仅支持 approved/rejected")
	}
	var reason *string
	if trimmed := strings.TrimSpace(req.Reason); trimmed != "" {
		reason = &trimmed
	}
	if err := h.citationRepo.UpdateStatus(c.Context(), id, status, reason); err != nil {
		return err
	}
	item, err := h.citationRepo.GetByID(c.Context(), id)
	if err == nil {
		_ = h.sendResultCallback(c.Context(), "citation", item.SourceInstanceID, item.SourceRequestID, status, req.Reason)
	}
	return response.SuccessWithMessage[any](c, nil, "审核结果已更新")
}

// ReviewMention 审核提及。
// @Summary 审核联合提及
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param id path int true "提及ID"
// @Param request body contract.FederationReviewDecisionReq true "审核结果"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /admin/federation/mentions/{id}/review [put]
// @Security JWTAuth
func (h *FederationReviewHandler) ReviewMention(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的提及ID")
	}
	var req contract.FederationReviewDecisionReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	status := normalizeReviewStatus(req.Status)
	if status == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "status 仅支持 approved/rejected")
	}
	var reason *string
	if trimmed := strings.TrimSpace(req.Reason); trimmed != "" {
		reason = &trimmed
	}
	if err := h.mentionRepo.UpdateStatus(c.Context(), id, status, reason); err != nil {
		return err
	}
	item, err := h.mentionRepo.GetByID(c.Context(), id)
	if err == nil {
		_ = h.sendResultCallback(c.Context(), "mention", item.SourceInstanceID, item.SourceRequestID, status, req.Reason)
	}
	return response.SuccessWithMessage[any](c, nil, "审核结果已更新")
}

func (h *FederationReviewHandler) sendResultCallback(ctx context.Context, typ string, sourceInstanceID int64, sourceRequestID *string, status string, reason string) error {
	if h.outbound == nil || h.instanceRepo == nil || sourceRequestID == nil || strings.TrimSpace(*sourceRequestID) == "" {
		return nil
	}
	instance, err := h.instanceRepo.GetByID(ctx, sourceInstanceID)
	if err != nil || instance == nil {
		return err
	}
	processedAt := time.Now().UTC().Format(time.RFC3339)
	_, _, _, err = h.outbound.SendResultCallback(ctx, instance.BaseURL, contract.FederationOutboundResultReq{
		RequestID:      strings.TrimSpace(*sourceRequestID),
		Type:           typ,
		Status:         mapReviewToOutboundStatus(status),
		Reason:         strings.TrimSpace(reason),
		ProcessedAt:    processedAt,
		RemoteTicketID: "",
	})
	return err
}

func normalizeReviewStatus(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "approved":
		return "approved"
	case "rejected":
		return "rejected"
	default:
		return ""
	}
}

func mapReviewToOutboundStatus(status string) string {
	switch status {
	case "approved":
		return domainfed.DeliveryStatusApproved
	case "rejected":
		return domainfed.DeliveryStatusRejected
	default:
		return domainfed.DeliveryStatusFailed
	}
}
