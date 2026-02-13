package handler

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"

	"github.com/gofiber/fiber/v2"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/article"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type ArticleHandler struct {
	svc         *article.Service
	contentRepo content.Repository
	commentRepo domaincomment.CommentRepository
	userRepo    identity.Repository
}

func NewArticleHandler(svc *article.Service, contentRepo content.Repository, commentRepo domaincomment.CommentRepository, userRepo identity.Repository) *ArticleHandler {
	return &ArticleHandler{
		svc:         svc,
		contentRepo: contentRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
	}
}

// CreateArticle godoc
// @Summary 创建文章
// @Tags Article
// @Accept json
// @Produce json
// @Param request body contract.CreateArticleReq true "创建文章参数"
// @Success 200 {object} contract.ArticleResp
// @Security BearerAuth
// @Router /articles [post]
// @Security JWTAuth
func (h *ArticleHandler) CreateArticle(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.CreateArticleReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	extInfo, err := parseExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	var cmd article.CreateArticleCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}
	if cmd.AllowComment == nil {
		defaultAllow := true
		cmd.AllowComment = &defaultAllow
	}
	cmd.ExtInfo = extInfo

	createdArticle, err := h.svc.CreateArticle(c.Context(), claims.UserID, cmd)
	if err != nil {
		if errors.Is(err, content.ErrArticleShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		if errors.Is(err, content.ErrCategoryNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "分类不存在")
		}
		if errors.Is(err, content.ErrTagNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "标签不存在")
		}
		return err
	}

	articleResponse, err := h.toArticleResp(c.Context(), createdArticle)
	if err != nil {
		return err
	}

	Audit(c, "article.create", map[string]any{
		"articleId": createdArticle.ID,
		"title":     createdArticle.Title,
		"userId":    claims.UserID,
	})

	return response.SuccessWithMessage(c, articleResponse, "文章创建成功")
}

// UpdateArticle godoc
// @Summary 更新文章
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param request body contract.UpdateArticleReq true "更新文章参数"
// @Success 200 {object} contract.ArticleResp
// @Security BearerAuth
// @Router /articles/{id} [put]
// @Security JWTAuth
func (h *ArticleHandler) UpdateArticle(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的文章ID")
	}

	var req contract.UpdateArticleReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	extInfo, err := parseExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	var cmd article.UpdateArticleCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}
	cmd.ID = id
	cmd.ExtInfo = extInfo

	updatedArticle, err := h.svc.UpdateArticle(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, content.ErrArticleShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		if errors.Is(err, content.ErrCategoryNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "分类不存在")
		}
		if errors.Is(err, content.ErrTagNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "标签不存在")
		}
		return err
	}

	articleResponse, err := h.toArticleResp(c.Context(), updatedArticle)
	if err != nil {
		return err
	}

	Audit(c, "article.update", map[string]any{
		"articleId": updatedArticle.ID,
		"title":     updatedArticle.Title,
		"userId":    claims.UserID,
	})

	return response.SuccessWithMessage(c, articleResponse, "文章更新成功")
}

// BatchSetArticlePublished godoc
// @Summary 批量设置文章发布状态（管理端）
// @Tags Article
// @Accept json
// @Produce json
// @Param request body contract.BatchSetArticlePublishedReq true "批量发布状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/articles/published [put]
// @Security JWTAuth
func (h *ArticleHandler) BatchSetArticlePublished(c *fiber.Ctx) error {
	var req contract.BatchSetArticlePublishedReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}
	for _, id := range req.IDs {
		if id <= 0 {
			return response.NewBizErrorWithMsg(response.ParamsError, "ids 必须为正整数")
		}
	}

	if err := h.svc.BatchSetPublished(c.Context(), article.BatchSetPublishedCmd{
		IDs:         req.IDs,
		IsPublished: req.IsPublished,
	}); err != nil {
		return err
	}

	if req.IsPublished {
		return response.SuccessWithMessage[any](c, nil, "文章发布状态已批量更新为已发布")
	}
	return response.SuccessWithMessage[any](c, nil, "文章发布状态已批量更新为未发布")
}

// BatchSetArticleTop godoc
// @Summary 批量设置文章置顶状态（管理端）
// @Tags Article
// @Accept json
// @Produce json
// @Param request body contract.BatchSetArticleTopReq true "批量置顶状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/articles/top [put]
// @Security JWTAuth
func (h *ArticleHandler) BatchSetArticleTop(c *fiber.Ctx) error {
	var req contract.BatchSetArticleTopReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}
	for _, id := range req.IDs {
		if id <= 0 {
			return response.NewBizErrorWithMsg(response.ParamsError, "ids 必须为正整数")
		}
	}

	if err := h.svc.BatchSetTop(c.Context(), article.BatchSetTopCmd{
		IDs:   req.IDs,
		IsTop: req.IsTop,
	}); err != nil {
		return err
	}

	if req.IsTop {
		return response.SuccessWithMessage[any](c, nil, "文章置顶状态已批量更新为置顶")
	}
	return response.SuccessWithMessage[any](c, nil, "文章置顶状态已批量更新为取消置顶")
}

// GetArticle godoc
// @Summary 获取文章详情
// @Tags Article
// @Produce json
// @Param id path int true "文章ID"
// @Security BearerAuth
// @Success 200 {object} contract.ArticleResp
// @Router /articles/{id} [get]
func (h *ArticleHandler) GetArticle(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的文章ID")
	}

	article, err := h.svc.GetArticleByID(c.Context(), id)
	if err != nil {
		return err
	}

	articleResponse, err := h.toArticleResp(c.Context(), article)
	if err != nil {
		return err
	}

	return response.Success(c, articleResponse)
}

// ListSamePeriodMoments godoc
// @Summary 获取文章同一时间的手记（两周内）
// @Tags Article
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} contract.SamePeriodMomentListResp
// @Router /articles/{id}/same-period-moments [get]
func (h *ArticleHandler) ListSamePeriodMoments(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的文章ID")
	}

	articleItem, err := h.svc.GetArticleByID(c.Context(), id)
	if err != nil {
		return err
	}

	const windowDays = 14
	const limit = 2
	start := articleItem.CreatedAt.AddDate(0, 0, -windowDays)
	end := articleItem.CreatedAt.AddDate(0, 0, windowDays)

	moments, err := h.contentRepo.ListPublishedMomentsByCreatedAtRange(c.Context(), start, end, limit)
	if err != nil {
		return err
	}

	items := make([]contract.SamePeriodMomentItemResp, 0, len(moments))
	for _, item := range moments {
		resp := contract.SamePeriodMomentItemResp{
			ID:        item.ID,
			Title:     item.Title,
			ShortURL:  item.ShortURL,
			Summary:   item.Summary,
			CreatedAt: item.CreatedAt,
		}
		if item.Image != nil {
			resp.Image = splitImages(item.Image)
		}
		items = append(items, resp)
	}

	return response.Success(c, contract.SamePeriodMomentListResp{
		Items: items,
	})
}

// GetArticleByShortURL godoc
// @Summary 根据短链接获取文章
// @Tags Article
// @Produce json
// @Param shortUrl path string true "短链接"
// @Success 200 {object} contract.ArticleResp
// @Router /articles/short/{shortUrl} [get]
func (h *ArticleHandler) GetArticleByShortURL(c *fiber.Ctx) error {
	shortURL := c.Params("shortUrl")
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "短链接不能为空")
	}

	article, err := h.svc.GetArticleByShortURL(c.Context(), shortURL)
	if err != nil {
		return err
	}

	articleResponse, err := h.toArticleResp(c.Context(), article)
	if err != nil {
		return err
	}

	return response.Success(c, articleResponse)
}

// ListArticles godoc
// @Summary 获取文章列表
// @Tags Article
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param categoryId query int false "分类ID"
// @Param tagId query int false "标签ID"
// @Param search query string false "搜索关键词"
// @Success 200 {object} contract.ArticleListResp
// @Router /articles [get]
func (h *ArticleHandler) ListArticles(c *fiber.Ctx) error {
	query := buildArticleListQuery(c)

	// 公共接口只返回已发布
	published := true
	query.Published = &published

	return h.listArticlesWithQuery(c, query)
}

// ListArticlesByCategoryShortURL godoc
// @Summary 根据分类短链接获取文章列表
// @Tags Article
// @Produce json
// @Param shortUrl path string true "分类短链接"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} contract.ArticleListResp
// @Router /categories/short/{shortUrl}/articles [get]
func (h *ArticleHandler) ListArticlesByCategoryShortURL(c *fiber.Ctx) error {
	shortURL := strings.TrimSpace(c.Params("shortUrl"))
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "分类短链接不能为空")
	}

	category, err := h.contentRepo.GetCategoryByShortURL(c.Context(), shortURL)
	if err != nil {
		if errors.Is(err, content.ErrCategoryNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "分类不存在")
		}
		return err
	}

	query := buildArticleListQuery(c)
	query.CategoryID = &category.ID
	published := true
	query.Published = &published

	return h.listArticlesWithQuery(c, query)
}

// ListArticlesAdmin godoc
// @Summary 获取文章列表（管理员）
// @Tags Article
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param categoryId query int false "分类ID"
// @Param tagId query int false "标签ID"
// @Param authorId query int false "作者ID"
// @Param published query bool false "是否发布"
// @Param search query string false "搜索关键词"
// @Success 200 {object} contract.ArticleListResp
// @Security BearerAuth
// @Router /admin/articles [get]
func (h *ArticleHandler) ListArticlesAdmin(c *fiber.Ctx) error {
	query := buildArticleListQuery(c)
	return h.listArticlesWithQuery(c, query)
}

func (h *ArticleHandler) listArticlesWithQuery(c *fiber.Ctx, query contract.ListArticlesReq) error {
	articles, total, err := h.svc.ListArticles(c.Context(), content.ArticleListOptionsInternal(query))
	if err != nil {
		return err
	}

	articleResponses := make([]contract.ArticleListItemResp, len(articles))
	for i, art := range articles {
		resp, err := h.toArticleListItemResp(c.Context(), art)
		if err != nil {
			return err
		}
		articleResponses[i] = *resp
	}

	listResponse := contract.ArticleListResp{
		Items: articleResponses,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}

	return response.Success(c, listResponse)
}

func buildArticleListQuery(c *fiber.Ctx) contract.ListArticlesReq {
	query := contract.ListArticlesReq{
		Page:     1,
		PageSize: 10,
	}

	if page, err := strconv.Atoi(c.Query("page", "1")); err == nil && page > 0 {
		query.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("pageSize", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		query.PageSize = pageSize
	}
	if categoryID, err := strconv.ParseInt(c.Query("categoryId"), 10, 64); err == nil {
		query.CategoryID = &categoryID
	}
	if tagID, err := strconv.ParseInt(c.Query("tagId"), 10, 64); err == nil {
		query.TagID = &tagID
	}
	if authorID, err := strconv.ParseInt(c.Query("authorId"), 10, 64); err == nil {
		query.AuthorID = &authorID
	}
	if publishedStr := c.Query("published"); publishedStr != "" {
		if published, err := strconv.ParseBool(publishedStr); err == nil {
			query.Published = &published
		}
	}
	if search := c.Query("search"); search != "" {
		query.Search = &search
	}

	return query
}

// ListRecentPublicArticles godoc
// @Summary 获取最近公开文章
// @Tags Public
// @Produce json
// @Success 200 {object} contract.ArticleListResp
// @Router /public/articles/recent [get]
func (h *ArticleHandler) ListRecentPublicArticles(c *fiber.Ctx) error {
	const page = 1
	const size = 5
	published := true

	articles, total, err := h.svc.ListArticles(c.Context(), content.ArticleListOptionsInternal{
		Page:      page,
		PageSize:  size,
		Published: &published,
	})
	if err != nil {
		return err
	}

	articleResponses := make([]contract.ArticleListItemResp, len(articles))
	for i, art := range articles {
		resp, err := h.toArticleListItemResp(c.Context(), art)
		if err != nil {
			return err
		}
		articleResponses[i] = *resp
	}

	return response.Success(c, contract.ArticleListResp{
		Items: articleResponses,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// CheckArticleLatest godoc
// @Summary 校验文章是否最新
// @Tags Article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param request body contract.CheckArticleLatestReq true "文章版本校验参数"
// @Success 200 {object} contract.CheckArticleLatestResp
// @Router /articles/{id}/latest [post]
func (h *ArticleHandler) CheckArticleLatest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的文章ID")
	}

	var req contract.CheckArticleLatestReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	article, err := h.svc.GetArticleByID(c.Context(), id)
	if errors.Is(err, content.ErrArticleNotFound) {
		return response.NewBizErrorWithMsg(response.NotFound, "文章不存在")
	} else if err != nil {
		return err
	}

	if req.Hash == article.ContentHash {
		return response.Success(c, contract.CheckArticleLatestResp{
			Latest: true,
			ArticleContentPayload: contract.ArticleContentPayload{
				ContentHash: article.ContentHash,
			},
		})
	}

	return response.Success(c, contract.CheckArticleLatestResp{
		Latest: false,
		ArticleContentPayload: contract.ArticleContentPayload{
			ContentHash: article.ContentHash,
			Title:       article.Title,
			LeadIn:      article.LeadIn,
			TOC:         mapTOCNodes(article.TOC),
			Content:     article.Content,
		},
	})
}

// DeleteArticle godoc
// @Summary 删除文章
// @Tags Article
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /articles/{id} [delete]
// @Security JWTAuth
func (h *ArticleHandler) DeleteArticle(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的文章ID")
	}

	if err := h.svc.DeleteArticle(c.Context(), id); err != nil {
		return err
	}

	Audit(c, "article.delete", map[string]any{
		"articleId": id,
		"userId":    claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "文章删除成功")
}

// BatchDeleteArticles godoc
// @Summary 批量删除文章（管理端）
// @Tags Article
// @Accept json
// @Produce json
// @Param request body contract.BatchDeleteArticleReq true "批量删除参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/articles/batch-delete [post]
// @Security JWTAuth
func (h *ArticleHandler) BatchDeleteArticles(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.BatchDeleteArticleReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}
	for _, id := range req.IDs {
		if id <= 0 {
			return response.NewBizErrorWithMsg(response.ParamsError, "ids 必须为正整数")
		}
	}

	if err := h.svc.BatchDelete(c.Context(), article.BatchDeleteCmd{IDs: req.IDs}); err != nil {
		return err
	}

	Audit(c, "article.batch_delete", map[string]any{
		"articleIds": req.IDs,
		"userId":     claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "文章批量删除成功")
}

func (h *ArticleHandler) toArticleResp(ctx context.Context, article *content.Article) (*contract.ArticleResp, error) {
	tags, err := h.svc.GetArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	metrics, err := h.svc.GetArticleMetrics(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	var resp contract.ArticleResp
	if err := copier.Copy(&resp, article); err != nil {
		return nil, err
	}
	resp.TOC = mapTOCNodes(article.TOC)
	resp.ExtInfo = jsonRawFromBytes(article.ExtInfo)
	resp.AllowComment = h.allowCommentByAreaID(ctx, article.CommentID)

	if len(tags) > 0 {
		resp.Tags = make([]contract.TagResp, len(tags))
		for i, tag := range tags {
			if err := copier.Copy(&resp.Tags[i], tag); err != nil {
				return nil, err
			}
		}
	}

	if metrics != nil {
		var metricsResp contract.MetricsResp
		if err := copier.Copy(&metricsResp, metrics); err != nil {
			return nil, err
		}
		resp.Metrics = &metricsResp
	}

	return &resp, nil
}

func (h *ArticleHandler) toArticleListItemResp(ctx context.Context, article *content.Article) (*contract.ArticleListItemResp, error) {
	tags, err := h.svc.GetArticleTags(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	metrics, err := h.svc.GetArticleMetrics(ctx, article.ID)
	if err != nil {
		return nil, err
	}

	resp := contract.ArticleListItemResp{
		ID:           article.ID,
		Title:        article.Title,
		ShortURL:     article.ShortURL,
		Summary:      article.Summary,
		IsTop:        article.IsTop,
		IsHot:        article.IsHot,
		AllowComment: h.allowCommentByAreaID(ctx, article.CommentID),
		IsOriginal:   article.IsOriginal,
		IsPublished:  article.IsPublished,
		CreatedAt:    article.CreatedAt,
		UpdatedAt:    article.UpdatedAt,
		Tags:         []string{},
	}
	resp.CommentID = article.CommentID

	if article.Cover != nil {
		resp.Cover = *article.Cover
	}

	if metrics != nil {
		resp.Views = metrics.Views
		resp.Likes = metrics.Likes
		resp.Comments = metrics.Comments
	}

	if len(tags) > 0 {
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		resp.Tags = tagNames
	}

	if article.CategoryID != nil {
		category, err := h.contentRepo.GetCategoryByID(ctx, *article.CategoryID)
		if err != nil {
			if !errors.Is(err, content.ErrCategoryNotFound) {
				return nil, err
			}
		} else if category != nil {
			resp.CategoryName = category.Name
			if category.ShortURL != nil {
				resp.CategoryShortURL = *category.ShortURL
			}
		}
	}

	if h.userRepo != nil {
		user, err := h.userRepo.FindByID(ctx, article.AuthorID)
		if err != nil {
			if !errors.Is(err, identity.ErrUserNotFound) {
				return nil, err
			}
		} else if user != nil {
			resp.AuthorName = user.Nickname
			resp.Avatar = user.Avatar
		}
	}

	return &resp, nil
}

func (h *ArticleHandler) allowCommentByAreaID(ctx context.Context, areaID *int64) bool {
	if h.commentRepo == nil || areaID == nil || *areaID <= 0 {
		return true
	}
	area, err := h.commentRepo.GetAreaByID(ctx, *areaID)
	if err != nil || area == nil {
		return false
	}
	return !area.IsClosed
}
func mapTOCNodes(nodes []content.TOCNode) []contract.TOCNode {
	result := make([]contract.TOCNode, len(nodes))
	for i, node := range nodes {
		result[i] = contract.TOCNode{
			Name:     node.Name,
			Anchor:   node.Anchor,
			Children: mapTOCNodes(node.Children),
		}
	}
	return result
}
