package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/redis/go-redis/v9"
)

type TagContentHandler struct {
	articleHandler *ArticleHandler
	momentHandler  *MomentHandler
	contentRepo    content.Repository
	redis          *redis.Client
	redisPrefix    string
}

func NewTagContentHandler(articleHandler *ArticleHandler, momentHandler *MomentHandler, contentRepo content.Repository, redisClient *redis.Client, redisPrefix string) *TagContentHandler {
	return &TagContentHandler{
		articleHandler: articleHandler,
		momentHandler:  momentHandler,
		contentRepo:    contentRepo,
		redis:          redisClient,
		redisPrefix:    redisPrefix,
	}
}

func (h *TagContentHandler) cacheKey(tagID int64) string {
	return fmt.Sprintf("%stag:contents:%d", h.redisPrefix, tagID)
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

	// Try Redis cache first.
	if h.redis != nil {
		cached, err := h.redis.Get(c.Context(), h.cacheKey(tagID)).Bytes()
		if err == nil {
			var resp contract.TagRelatedContentsResp
			if json.Unmarshal(cached, &resp) == nil {
				return response.Success(c, resp)
			}
		}
	}

	articles, err := h.listAllTaggedArticles(c.Context(), tagID)
	if err != nil {
		return err
	}
	moments, err := h.listAllTaggedMoments(c.Context(), tagID)
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

	resp := contract.TagRelatedContentsResp{
		Articles: articleItems,
		Moments:  momentItems,
	}

	// Write to Redis cache (no TTL — event-driven invalidation only).
	if h.redis != nil {
		if data, err := json.Marshal(resp); err == nil {
			if err := h.redis.Set(c.Context(), h.cacheKey(tagID), data, 0).Err(); err != nil {
				log.Printf("tag content cache write error: %v", err)
			}
		}
	}

	return response.Success(c, resp)
}

func (h *TagContentHandler) listAllTaggedArticles(ctx context.Context, tagID int64) ([]*content.Article, error) {
	const pageSize = 100
	var all []*content.Article
	for page := 1; ; page++ {
		items, total, err := h.articleHandler.svc.ListPublicArticles(ctx, content.ArticleListOptions{
			Page:     page,
			PageSize: pageSize,
			TagID:    &tagID,
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

func (h *TagContentHandler) listAllTaggedMoments(ctx context.Context, tagID int64) ([]*content.Moment, error) {
	const pageSize = 100
	var all []*content.Moment
	for page := 1; ; page++ {
		items, total, err := h.momentHandler.svc.ListPublicMoments(ctx, content.MomentListOptions{
			Page:     page,
			PageSize: pageSize,
			TopicID:  &tagID,
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
