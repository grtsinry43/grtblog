package handler

import (
	"errors"
	"fmt"
	"strconv"

	appsearch "github.com/grtsinry43/grtblog-v2/server/internal/app/search"
	domainsearch "github.com/grtsinry43/grtblog-v2/server/internal/domain/search"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"

	"github.com/gofiber/fiber/v2"
)

type SearchHandler struct {
	svc *appsearch.Service
}

func NewSearchHandler(svc *appsearch.Service) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// SiteSearch godoc
// @Summary 全站搜索
// @Tags Search
// @Produce json
// @Param q query string true "搜索关键词"
// @Param limit query int false "每个分组最多返回条数，默认8，最大20"
// @Success 200 {object} contract.SiteSearchResp
// @Router /public/search [get]
func (h *SearchHandler) SiteSearch(c *fiber.Ctx) error {
	query := c.Query("q")
	limit := c.QueryInt("limit", 8)

	result, err := h.svc.SearchSite(c.Context(), query, limit)
	if err != nil {
		if errors.Is(err, appsearch.ErrEmptyQuery) {
			return response.NewBizErrorWithMsg(response.ParamsError, "搜索关键词不能为空")
		}
		return err
	}

	resp := contract.SiteSearchResp{
		Query:     result.Query,
		Keywords:  result.Keywords,
		Cached:    result.Cached,
		Articles:  mapSearchHitsToResp(result.Articles),
		Moments:   mapSearchHitsToResp(result.Moments),
		Pages:     mapSearchHitsToResp(result.Pages),
		Thinkings: mapSearchHitsToResp(result.Thinkings),
	}
	return response.Success(c, resp)
}

func mapSearchHitsToResp(hits []domainsearch.Hit) []contract.SiteSearchItemResp {
	items := make([]contract.SiteSearchItemResp, 0, len(hits))
	for _, hit := range hits {
		items = append(items, contract.SiteSearchItemResp{
			ID:        hit.ID,
			Title:     hit.Title,
			Summary:   hit.Summary,
			Snippet:   hit.Snippet,
			ShortURL:  hit.ShortURL,
			Path:      buildSearchPath(hit.Kind, hit.ID, hit.ShortURL),
			Score:     hit.Score,
			CreatedAt: hit.CreatedAt,
		})
	}
	return items
}

func buildSearchPath(kind domainsearch.Kind, id int64, shortURL *string) string {
	switch kind {
	case domainsearch.KindArticle:
		if shortURL != nil && *shortURL != "" {
			return "/articles/short/" + *shortURL
		}
		return "/articles/" + strconv.FormatInt(id, 10)
	case domainsearch.KindMoment:
		if shortURL != nil && *shortURL != "" {
			return "/moments/short/" + *shortURL
		}
		return "/moments/" + strconv.FormatInt(id, 10)
	case domainsearch.KindPage:
		if shortURL != nil && *shortURL != "" {
			return "/pages/short/" + *shortURL
		}
		return "/pages/" + strconv.FormatInt(id, 10)
	case domainsearch.KindThinking:
		return "/thinkings/" + strconv.FormatInt(id, 10)
	default:
		return fmt.Sprintf("/content/%d", id)
	}
}
