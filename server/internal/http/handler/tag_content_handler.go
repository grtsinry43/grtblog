package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type TagContentHandler struct {
	articleHandler *ArticleHandler
	momentHandler  *MomentHandler
	contentRepo    content.Repository
}

func NewTagContentHandler(articleHandler *ArticleHandler, momentHandler *MomentHandler, contentRepo content.Repository) *TagContentHandler {
	return &TagContentHandler{
		articleHandler: articleHandler,
		momentHandler:  momentHandler,
		contentRepo:    contentRepo,
	}
}

// ListByTagID godoc
// @Summary 根据标签聚合返回相关文章和手记
// @Tags Tag
// @Produce json
// @Param id path int true "标签ID"
// @Success 200 {object} contract.TagRelatedContentsResp
// @Router /tags/{id}/contents [get]
func (h *TagContentHandler) ListByTagID(c *fiber.Ctx) error {
	tagID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || tagID <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的标签ID")
	}

	if _, err := h.contentRepo.GetTagByID(c.Context(), tagID); err != nil {
		if errors.Is(err, content.ErrTagNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "标签不存在")
		}
		return err
	}

	published := true
	articles, err := h.listAllTaggedArticles(c.Context(), tagID, &published)
	if err != nil {
		return err
	}
	moments, err := h.listAllTaggedMoments(c.Context(), tagID, &published)
	if err != nil {
		return err
	}

	articleItems := make([]contract.ArticleListItemResp, len(articles))
	for i, item := range articles {
		respItem, err := h.articleHandler.toArticleListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		articleItems[i] = *respItem
	}

	momentItems := make([]contract.MomentListItemResp, len(moments))
	for i, item := range moments {
		respItem, err := h.momentHandler.toMomentListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		momentItems[i] = *respItem
	}

	return response.Success(c, contract.TagRelatedContentsResp{
		Articles: articleItems,
		Moments:  momentItems,
	})
}

func (h *TagContentHandler) listAllTaggedArticles(ctx context.Context, tagID int64, published *bool) ([]*content.Article, error) {
	const pageSize = 100
	var all []*content.Article
	for page := 1; ; page++ {
		items, total, err := h.articleHandler.svc.ListArticles(ctx, content.ArticleListOptionsInternal{
			Page:      page,
			PageSize:  pageSize,
			TagID:     &tagID,
			Published: published,
		})
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}
		all = append(all, items...)
		if int64(len(all)) >= total {
			break
		}
	}
	return all, nil
}

func (h *TagContentHandler) listAllTaggedMoments(ctx context.Context, tagID int64, published *bool) ([]*content.Moment, error) {
	const pageSize = 100
	var all []*content.Moment
	for page := 1; ; page++ {
		items, total, err := h.momentHandler.svc.ListMoments(ctx, content.MomentListOptionsInternal{
			Page:      page,
			PageSize:  pageSize,
			TopicID:   &tagID,
			Published: published,
		})
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}
		all = append(all, items...)
		if int64(len(all)) >= total {
			break
		}
	}
	return all, nil
}
