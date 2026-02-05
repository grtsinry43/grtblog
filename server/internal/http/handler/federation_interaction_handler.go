package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type FederationInteractionHandler struct {
	contentRepo  content.Repository
	citationRepo domainfed.FederatedCitationRepository
	outboundRepo domainfed.OutboundDeliveryRepository
}

func NewFederationInteractionHandler(contentRepo content.Repository, citationRepo domainfed.FederatedCitationRepository, outboundRepo domainfed.OutboundDeliveryRepository) *FederationInteractionHandler {
	return &FederationInteractionHandler{
		contentRepo:  contentRepo,
		citationRepo: citationRepo,
		outboundRepo: outboundRepo,
	}
}

// GetArticleInteractions 查询文章对应的联合互动。
// @Summary 文章联合互动
// @Tags Federation
// @Produce json
// @Param id path string true "文章 ID 或短链接"
// @Success 200 {object} contract.FederationArticleInteractionsResp
// @Security BearerAuth
// @Router /articles/{id}/federation/interactions [get]
// @Security JWTAuth
func (h *FederationInteractionHandler) GetArticleInteractions(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.NewBizError(response.NotLogin)
	}
	rawID := strings.TrimSpace(c.Params("id"))
	if rawID == "" {
		return response.NewBizError(response.ParamsError)
	}
	article, err := h.resolveArticle(c, rawID)
	if err != nil {
		if errors.Is(err, content.ErrArticleNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return response.NewBizErrorWithCause(response.ServerError, "文章查询失败", err)
	}
	if !claims.IsAdmin && claims.UserID != article.AuthorID {
		return response.NewBizErrorWithMsg(response.Unauthorized, "仅作者可查看联合互动")
	}

	citations, err := h.citationRepo.ListByTarget(c.Context(), article.ID, "")
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "引用记录查询失败", err)
	}
	outbounds, err := h.outboundRepo.ListBySourceArticle(c.Context(), article.ID, 100)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "出站记录查询失败", err)
	}

	resp := contract.FederationArticleInteractionsResp{
		ArticleID:        article.ID,
		InboundCitations: make([]contract.FederationCitationInteractionResp, len(citations)),
		Outbound:         make([]contract.FederationOutboundInteractionResp, len(outbounds)),
	}
	for i := range citations {
		resp.InboundCitations[i] = contract.FederationCitationInteractionResp{
			ID:               citations[i].ID,
			SourceInstanceID: citations[i].SourceInstanceID,
			SourcePostURL:    citations[i].SourcePostURL,
			SourcePostTitle:  citations[i].SourcePostTitle,
			CitationType:     citations[i].CitationType,
			Status:           citations[i].Status,
			RequestedAt:      citations[i].RequestedAt.UTC().Format(time.RFC3339),
		}
	}
	for i := range outbounds {
		resp.Outbound[i] = contract.FederationOutboundInteractionResp{
			ID:                outbounds[i].ID,
			RequestID:         outbounds[i].RequestID,
			Type:              outbounds[i].DeliveryType,
			TargetInstanceURL: outbounds[i].TargetInstanceURL,
			Status:            outbounds[i].Status,
			AttemptCount:      outbounds[i].AttemptCount,
			HTTPStatus:        outbounds[i].HTTPStatus,
			ErrorMessage:      outbounds[i].ErrorMessage,
			RemoteTicketID:    outbounds[i].RemoteTicketID,
			CreatedAt:         outbounds[i].CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:         outbounds[i].UpdatedAt.UTC().Format(time.RFC3339),
		}
	}
	return response.Success(c, resp)
}

func (h *FederationInteractionHandler) resolveArticle(c *fiber.Ctx, rawID string) (*content.Article, error) {
	if id, err := strconv.ParseInt(rawID, 10, 64); err == nil {
		return h.contentRepo.GetArticleByID(c.Context(), id)
	}
	return h.contentRepo.GetArticleByShortURL(c.Context(), rawID)
}
