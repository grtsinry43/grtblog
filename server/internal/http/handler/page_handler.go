package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/page"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/content"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type PageHandler struct {
	svc         *page.Service
	commentRepo domaincomment.CommentRepository
}

func NewPageHandler(svc *page.Service, commentRepo domaincomment.CommentRepository) *PageHandler {
	return &PageHandler{
		svc:         svc,
		commentRepo: commentRepo,
	}
}

// CreatePage godoc
// @Summary 创建页面
// @Tags Page
// @Accept json
// @Produce json
// @Param request body contract.CreatePageReq true "创建页面参数"
// @Success 200 {object} contract.PageResp
// @Security BearerAuth
// @Router /pages [post]
// @Security JWTAuth
func (h *PageHandler) CreatePage(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.CreatePageReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if req.Views != nil && *req.Views < 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "views 不能为负数")
	}
	extInfo, err := parseExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	cmd := page.CreatePageCmd{
		Title:        req.Title,
		Description:  req.Description,
		AISummary:    req.AISummary,
		Content:      req.Content,
		ShortURL:     req.ShortURL,
		AllowComment: req.AllowComment,
		IsEnabled:    req.IsEnabled,
		IsBuiltin:    req.IsBuiltin,
		ExtInfo:      extInfo,
		CreatedAt:    req.CreatedAt,
		Views:        req.Views,
	}
	if cmd.AllowComment == nil {
		defaultAllow := true
		cmd.AllowComment = &defaultAllow
	}

	createdPage, err := h.svc.CreatePage(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, content.ErrPageShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		return err
	}

	pageResponse, err := h.toPageResp(c.Context(), createdPage)
	if err != nil {
		return err
	}

	Audit(c, "page.create", map[string]any{
		"pageId": createdPage.ID,
		"title":  createdPage.Title,
		"userId": claims.UserID,
	})

	return response.SuccessWithMessage(c, pageResponse, "页面创建成功")
}

// UpdatePage godoc
// @Summary 更新页面
// @Tags Page
// @Accept json
// @Produce json
// @Param id path int true "页面ID"
// @Param request body contract.UpdatePageReq true "更新页面参数"
// @Success 200 {object} contract.PageResp
// @Security BearerAuth
// @Router /pages/{id} [put]
// @Security JWTAuth
func (h *PageHandler) UpdatePage(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的页面ID")
	}

	var req contract.UpdatePageReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	extInfo, err := parseExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	cmd := page.UpdatePageCmd{
		ID:           id,
		Title:        req.Title,
		Description:  req.Description,
		AISummary:    req.AISummary,
		Content:      req.Content,
		ShortURL:     req.ShortURL,
		AllowComment: req.AllowComment,
		IsEnabled:    req.IsEnabled,
		IsBuiltin:    req.IsBuiltin,
		ExtInfo:      extInfo,
	}

	updatedPage, err := h.svc.UpdatePage(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, content.ErrPageShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		return err
	}

	pageResponse, err := h.toPageResp(c.Context(), updatedPage)
	if err != nil {
		return err
	}

	Audit(c, "page.update", map[string]any{
		"pageId": updatedPage.ID,
		"title":  updatedPage.Title,
		"userId": claims.UserID,
	})

	return response.SuccessWithMessage(c, pageResponse, "页面更新成功")
}

// BatchSetPageEnabled godoc
// @Summary 批量设置页面启用状态（管理端）
// @Tags Page
// @Accept json
// @Produce json
// @Param request body contract.BatchSetPageEnabledReq true "批量启用状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/pages/enabled [put]
// @Security JWTAuth
func (h *PageHandler) BatchSetPageEnabled(c *fiber.Ctx) error {
	var req contract.BatchSetPageEnabledReq
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

	if err := h.svc.BatchSetEnabled(c.Context(), page.BatchSetEnabledCmd{
		IDs:       req.IDs,
		IsEnabled: req.IsEnabled,
	}); err != nil {
		return err
	}

	if req.IsEnabled {
		return response.SuccessWithMessage[any](c, nil, "页面启用状态已批量更新为启用")
	}
	return response.SuccessWithMessage[any](c, nil, "页面启用状态已批量更新为禁用")
}

// GetPage godoc
// @Summary 获取页面详情
// @Tags Page
// @Produce json
// @Param id path int true "页面ID"
// @Security BearerAuth
// @Success 200 {object} contract.PageResp
// @Router /pages/{id} [get]
func (h *PageHandler) GetPage(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的页面ID")
	}

	pageItem, err := h.svc.GetPageByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, content.ErrPageNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
		}
		return err
	}
	if !pageItem.IsEnabled {
		return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
	}

	pageResponse, err := h.toPageResp(c.Context(), pageItem)
	if err != nil {
		return err
	}

	return response.Success(c, pageResponse)
}

// GetPageAdmin godoc
// @Summary 获取页面详情（管理员）
// @Tags Page
// @Produce json
// @Param id path int true "页面ID"
// @Security BearerAuth
// @Success 200 {object} contract.PageResp
// @Router /admin/pages/{id} [get]
// @Security JWTAuth
func (h *PageHandler) GetPageAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的页面ID")
	}
	pageItem, err := h.svc.GetPageByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, content.ErrPageNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
		}
		return err
	}
	pageResponse, err := h.toPageResp(c.Context(), pageItem)
	if err != nil {
		return err
	}
	return response.Success(c, pageResponse)
}

// GetPageByShortURL godoc
// @Summary 根据短链接获取页面
// @Tags Page
// @Produce json
// @Param shortUrl path string true "短链接"
// @Success 200 {object} contract.PageResp
// @Router /pages/short/{shortUrl} [get]
func (h *PageHandler) GetPageByShortURL(c *fiber.Ctx) error {
	shortURL := c.Params("shortUrl")
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "短链接不能为空")
	}

	pageItem, err := h.svc.GetPageByShortURL(c.Context(), shortURL)
	if err != nil {
		if errors.Is(err, content.ErrPageNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
		}
		return err
	}
	if !pageItem.IsEnabled {
		return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
	}

	pageResponse, err := h.toPageResp(c.Context(), pageItem)
	if err != nil {
		return err
	}

	return response.Success(c, pageResponse)
}

// ListPages godoc
// @Summary 获取页面列表
// @Tags Page
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param enabled query bool false "启用状态"
// @Param builtin query bool false "内置页面"
// @Param search query string false "搜索关键词"
// @Success 200 {object} contract.PageListResp
// @Router /pages [get]
func (h *PageHandler) ListPages(c *fiber.Ctx) error {
	query := contract.ListPagesReq{
		Page:     1,
		PageSize: 10,
	}

	if page, err := strconv.Atoi(c.Query("page", "1")); err == nil && page > 0 {
		query.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("pageSize", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		query.PageSize = pageSize
	}
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			query.Enabled = &enabled
		}
	}
	if builtinStr := c.Query("builtin"); builtinStr != "" {
		if builtin, err := strconv.ParseBool(builtinStr); err == nil {
			query.Builtin = &builtin
		}
	}
	if search := c.Query("search"); search != "" {
		query.Search = &search
	}

	_, hasAuth := middleware.GetClaims(c)
	if !hasAuth {
		enabled := true
		query.Enabled = &enabled
	}

	pages, total, err := h.svc.ListPages(c.Context(), content.PageListOptionsInternal(query))
	if err != nil {
		return err
	}

	pageResponses := make([]contract.PageListItemResp, len(pages))
	for i, item := range pages {
		resp, err := h.toPageListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		pageResponses[i] = *resp
	}

	listResponse := contract.PageListResp{
		Items: pageResponses,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}

	return response.Success(c, listResponse)
}

// CheckPageLatest godoc
// @Summary 校验页面是否最新
// @Tags Page
// @Accept json
// @Produce json
// @Param id path int true "页面ID"
// @Param request body contract.CheckPageLatestReq true "页面版本校验参数"
// @Success 200 {object} contract.CheckPageLatestResp
// @Router /pages/{id}/latest [post]
func (h *PageHandler) CheckPageLatest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的页面ID")
	}

	var req contract.CheckPageLatestReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	pageItem, err := h.svc.GetPageByID(c.Context(), id)
	if errors.Is(err, content.ErrPageNotFound) {
		return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
	} else if err != nil {
		return err
	}
	if !pageItem.IsEnabled {
		return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
	}

	if req.Hash == pageItem.ContentHash {
		return response.Success(c, contract.CheckPageLatestResp{
			Latest: true,
			PageContentPayload: contract.PageContentPayload{
				ContentHash: pageItem.ContentHash,
			},
		})
	}

	return response.Success(c, contract.CheckPageLatestResp{
		Latest: false,
		PageContentPayload: contract.PageContentPayload{
			ContentHash: pageItem.ContentHash,
			Title:       pageItem.Title,
			Description: pageItem.Description,
			TOC:         mapPageTOCNodes(pageItem.TOC),
			Content:     pageItem.Content,
		},
	})
}

// DeletePage godoc
// @Summary 删除页面
// @Tags Page
// @Produce json
// @Param id path int true "页面ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /pages/{id} [delete]
// @Security JWTAuth
func (h *PageHandler) DeletePage(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的页面ID")
	}

	if err := h.svc.DeletePage(c.Context(), id); err != nil {
		return err
	}

	Audit(c, "page.delete", map[string]any{
		"pageId": id,
		"userId": claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "页面删除成功")
}

// BatchDeletePages godoc
// @Summary 批量删除页面（管理端）
// @Tags Page
// @Accept json
// @Produce json
// @Param request body contract.BatchDeletePageReq true "批量删除参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/pages/batch-delete [post]
// @Security JWTAuth
func (h *PageHandler) BatchDeletePages(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.BatchDeletePageReq
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

	if err := h.svc.BatchDelete(c.Context(), page.BatchDeleteCmd{IDs: req.IDs}); err != nil {
		return err
	}

	Audit(c, "page.batch_delete", map[string]any{
		"pageIds": req.IDs,
		"userId":  claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "页面批量删除成功")
}

// GetPageMetrics godoc
// @Summary 获取页面指标
// @Tags Page
// @Produce json
// @Param id path int true "页面ID"
// @Success 200 {object} contract.MetricsResp
// @Router /pages/{id}/metrics [get]
func (h *PageHandler) GetPageMetrics(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的页面ID")
	}

	metrics, err := h.svc.GetPageMetrics(c.Context(), id)
	if err != nil {
		if errors.Is(err, content.ErrPageNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "页面不存在")
		}
		return err
	}

	resp := contract.MetricsResp{}
	if metrics != nil {
		resp.Views = metrics.Views
		resp.Likes = metrics.Likes
		resp.Comments = metrics.Comments
	}

	return response.Success(c, resp)
}

func (h *PageHandler) toPageResp(ctx context.Context, pageItem *content.Page) (*contract.PageResp, error) {
	metrics, err := h.svc.GetPageMetrics(ctx, pageItem.ID)
	if err != nil {
		return nil, err
	}

	resp := contract.PageResp{
		ID:               pageItem.ID,
		Title:            pageItem.Title,
		Description:      pageItem.Description,
		AISummary:        pageItem.AISummary,
		TOC:              mapPageTOCNodes(pageItem.TOC),
		Content:          pageItem.Content,
		ContentHash:      pageItem.ContentHash,
		CommentID:        pageItem.CommentID,
		ShortURL:         pageItem.ShortURL,
		IsEnabled:        pageItem.IsEnabled,
		IsBuiltin:        pageItem.IsBuiltin,
		IsHot:            false,
		AllowComment:     h.allowCommentByAreaID(ctx, pageItem.CommentID),
		ExtInfo:          jsonRawFromBytes(pageItem.ExtInfo),
		ContentUpdatedAt: pageItem.ContentUpdatedAt,
		Metrics: &contract.MetricsResp{
			Views:    0,
			Likes:    0,
			Comments: 0,
		},
		CreatedAt: pageItem.CreatedAt,
		UpdatedAt: pageItem.UpdatedAt,
	}

	if metrics != nil {
		resp.Metrics.Views = metrics.Views
		resp.Metrics.Likes = metrics.Likes
		resp.Metrics.Comments = metrics.Comments
	}

	return &resp, nil
}

func (h *PageHandler) toPageListItemResp(ctx context.Context, pageItem *content.Page) (*contract.PageListItemResp, error) {
	metrics, err := h.svc.GetPageMetrics(ctx, pageItem.ID)
	if err != nil {
		return nil, err
	}

	resp := contract.PageListItemResp{
		ID:               pageItem.ID,
		Title:            pageItem.Title,
		ShortURL:         pageItem.ShortURL,
		Description:      pageItem.Description,
		IsEnabled:        pageItem.IsEnabled,
		IsBuiltin:        pageItem.IsBuiltin,
		IsHot:            false,
		AllowComment:     h.allowCommentByAreaID(ctx, pageItem.CommentID),
		ContentUpdatedAt: pageItem.ContentUpdatedAt,
		CreatedAt:        pageItem.CreatedAt,
		UpdatedAt:        pageItem.UpdatedAt,
	}
	resp.CommentID = pageItem.CommentID

	if metrics != nil {
		resp.Views = metrics.Views
		resp.Likes = metrics.Likes
		resp.Comments = metrics.Comments
	}

	return &resp, nil
}

func mapPageTOCNodes(nodes []content.TOCNode) []contract.TOCNode {
	result := make([]contract.TOCNode, len(nodes))
	for i, node := range nodes {
		result[i] = contract.TOCNode{
			Name:     node.Name,
			Anchor:   node.Anchor,
			Children: mapPageTOCNodes(node.Children),
		}
	}
	return result
}

func (h *PageHandler) allowCommentByAreaID(ctx context.Context, areaID *int64) bool {
	if h.commentRepo == nil || areaID == nil || *areaID <= 0 {
		return true
	}
	area, err := h.commentRepo.GetAreaByID(ctx, *areaID)
	if err != nil || area == nil {
		return false
	}
	return !area.IsClosed
}
