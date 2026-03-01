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

	"github.com/grtsinry43/grtblog-v2/server/internal/app/moment"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type MomentHandler struct {
	svc         *moment.Service
	contentRepo content.Repository
	commentRepo domaincomment.CommentRepository
	userRepo    identity.Repository
}

func NewMomentHandler(svc *moment.Service, contentRepo content.Repository, commentRepo domaincomment.CommentRepository, userRepo identity.Repository) *MomentHandler {
	return &MomentHandler{
		svc:         svc,
		contentRepo: contentRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
	}
}

// CreateMoment godoc
// @Summary 创建手记
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.CreateMomentReq true "创建手记参数"
// @Success 200 {object} contract.MomentResp
// @Security BearerAuth
// @Router /moments [post]
// @Security JWTAuth
func (h *MomentHandler) CreateMoment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.CreateMomentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	extInfo, err := parseExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	cmd := moment.CreateMomentCmd{
		Title:        req.Title,
		Summary:      req.Summary,
		Content:      req.Content,
		Image:        joinImages(req.Image),
		ColumnID:     req.ColumnID,
		TopicIDs:     req.TopicIDs,
		ShortURL:     req.ShortURL,
		IsPublished:  req.IsPublished,
		IsTop:        req.IsTop,
		AllowComment: req.AllowComment,
		IsOriginal:   req.IsOriginal,
		ExtInfo:      extInfo,
		CreatedAt:    req.CreatedAt,
	}
	if cmd.AllowComment == nil {
		defaultAllow := true
		cmd.AllowComment = &defaultAllow
	}

	createdMoment, err := h.svc.CreateMoment(c.Context(), claims.UserID, cmd)
	if err != nil {
		if errors.Is(err, content.ErrMomentShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		if errors.Is(err, content.ErrColumnNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "分区不存在")
		}
		if errors.Is(err, content.ErrTagNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "话题不存在")
		}
		return err
	}

	momentResponse, err := h.toMomentResp(c.Context(), createdMoment)
	if err != nil {
		return err
	}

	Audit(c, "moment.create", map[string]any{
		"momentId": createdMoment.ID,
		"title":    createdMoment.Title,
		"userId":   claims.UserID,
	})

	return response.SuccessWithMessage(c, momentResponse, "手记创建成功")
}

// UpdateMoment godoc
// @Summary 更新手记
// @Tags Moment
// @Accept json
// @Produce json
// @Param id path int true "手记ID"
// @Param request body contract.UpdateMomentReq true "更新手记参数"
// @Success 200 {object} contract.MomentResp
// @Security BearerAuth
// @Router /moments/{id} [put]
// @Security JWTAuth
func (h *MomentHandler) UpdateMoment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	var req contract.UpdateMomentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	extInfo, err := parseExtInfo(req.ExtInfo)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "extInfo格式错误", err)
	}

	cmd := moment.UpdateMomentCmd{
		Title:        req.Title,
		Summary:      req.Summary,
		Content:      req.Content,
		Image:        joinImages(req.Image),
		ColumnID:     req.ColumnID,
		TopicIDs:     req.TopicIDs,
		ShortURL:     req.ShortURL,
		IsPublished:  req.IsPublished,
		IsTop:        req.IsTop,
		AllowComment: req.AllowComment,
		IsOriginal:   req.IsOriginal,
		ExtInfo:      extInfo,
	}
	cmd.ID = id

	updatedMoment, err := h.svc.UpdateMoment(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, content.ErrMomentShortURLExists) {
			return response.NewBizErrorWithMsg(response.ParamsError, "短链接已存在")
		}
		if errors.Is(err, content.ErrColumnNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "分区不存在")
		}
		if errors.Is(err, content.ErrTagNotFound) {
			return response.NewBizErrorWithMsg(response.ParamsError, "话题不存在")
		}
		return err
	}

	momentResponse, err := h.toMomentResp(c.Context(), updatedMoment)
	if err != nil {
		return err
	}

	Audit(c, "moment.update", map[string]any{
		"momentId": updatedMoment.ID,
		"title":    updatedMoment.Title,
		"userId":   claims.UserID,
	})

	return response.SuccessWithMessage(c, momentResponse, "手记更新成功")
}

// BatchSetMomentPublished godoc
// @Summary 批量设置手记发布状态（管理端）
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.BatchSetMomentPublishedReq true "批量发布状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/moments/published [put]
// @Security JWTAuth
func (h *MomentHandler) BatchSetMomentPublished(c *fiber.Ctx) error {
	var req contract.BatchSetMomentPublishedReq
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

	if err := h.svc.BatchSetPublished(c.Context(), moment.BatchSetPublishedCmd{
		IDs:         req.IDs,
		IsPublished: req.IsPublished,
	}); err != nil {
		return err
	}

	if req.IsPublished {
		return response.SuccessWithMessage[any](c, nil, "手记发布状态已批量更新为已发布")
	}
	return response.SuccessWithMessage[any](c, nil, "手记发布状态已批量更新为未发布")
}

// BatchSetMomentTop godoc
// @Summary 批量设置手记置顶状态（管理端）
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.BatchSetMomentTopReq true "批量置顶状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/moments/top [put]
// @Security JWTAuth
func (h *MomentHandler) BatchSetMomentTop(c *fiber.Ctx) error {
	var req contract.BatchSetMomentTopReq
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

	if err := h.svc.BatchSetTop(c.Context(), moment.BatchSetTopCmd{
		IDs:   req.IDs,
		IsTop: req.IsTop,
	}); err != nil {
		return err
	}

	if req.IsTop {
		return response.SuccessWithMessage[any](c, nil, "手记置顶状态已批量更新为置顶")
	}
	return response.SuccessWithMessage[any](c, nil, "手记置顶状态已批量更新为取消置顶")
}

// GetMoment godoc
// @Summary 获取手记详情
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Security BearerAuth
// @Success 200 {object} contract.MomentResp
// @Router /moments/{id} [get]
func (h *MomentHandler) GetMoment(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if err != nil {
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	momentResponse, err := h.toMomentResp(c.Context(), momentItem)
	if err != nil {
		return err
	}

	return response.Success(c, momentResponse)
}

// GetMomentAdmin godoc
// @Summary 获取手记详情（管理员）
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Security BearerAuth
// @Success 200 {object} contract.MomentResp
// @Router /admin/moments/{id} [get]
// @Security JWTAuth
func (h *MomentHandler) GetMomentAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}
	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if err != nil {
		return err
	}
	momentResponse, err := h.toMomentResp(c.Context(), momentItem)
	if err != nil {
		return err
	}
	return response.Success(c, momentResponse)
}

// ListSamePeriodArticles godoc
// @Summary 获取手记同一时期的文章（两周内）
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Success 200 {object} contract.SamePeriodArticleListResp
// @Router /moments/{id}/same-period-articles [get]
func (h *MomentHandler) ListSamePeriodArticles(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if err != nil {
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	const windowDays = 14
	const limit = 2
	start := momentItem.CreatedAt.AddDate(0, 0, -windowDays)
	end := momentItem.CreatedAt.AddDate(0, 0, windowDays)

	articles, err := h.contentRepo.ListPublishedArticlesByCreatedAtRange(c.Context(), start, end, limit)
	if err != nil {
		return err
	}

	items := make([]contract.SamePeriodArticleItemResp, 0, len(articles))
	for _, item := range articles {
		resp := contract.SamePeriodArticleItemResp{
			ID:        item.ID,
			Title:     item.Title,
			ShortURL:  item.ShortURL,
			Summary:   item.Summary,
			CreatedAt: item.CreatedAt,
		}
		if item.Cover != nil {
			resp.Cover = *item.Cover
		}
		items = append(items, resp)
	}

	return response.Success(c, contract.SamePeriodArticleListResp{
		Items: items,
	})
}

// GetMomentByShortURL godoc
// @Summary 根据短链接获取手记
// @Tags Moment
// @Produce json
// @Param shortUrl path string true "短链接"
// @Success 200 {object} contract.MomentResp
// @Router /moments/short/{shortUrl} [get]
func (h *MomentHandler) GetMomentByShortURL(c *fiber.Ctx) error {
	shortURL := c.Params("shortUrl")
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "短链接不能为空")
	}

	momentItem, err := h.svc.GetMomentByShortURL(c.Context(), shortURL)
	if err != nil {
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	momentResponse, err := h.toMomentResp(c.Context(), momentItem)
	if err != nil {
		return err
	}

	return response.Success(c, momentResponse)
}

// ListMoments godoc
// @Summary 获取手记列表
// @Tags Moment
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param columnId query int false "分区ID"
// @Param topicId query int false "话题ID"
// @Param search query string false "搜索关键词"
// @Success 200 {object} contract.MomentListResp
// @Router /moments [get]
func (h *MomentHandler) ListMoments(c *fiber.Ctx) error {
	query := buildMomentListQuery(c)

	// 公共接口只返回已发布
	published := true
	query.Published = &published

	return h.listMomentsWithQuery(c, query)
}

// ListMomentsByColumnShortURL godoc
// @Summary 根据专栏短链接获取手记列表
// @Tags Moment
// @Produce json
// @Param shortUrl path string true "专栏短链接"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} contract.MomentListResp
// @Router /columns/short/{shortUrl}/moments [get]
func (h *MomentHandler) ListMomentsByColumnShortURL(c *fiber.Ctx) error {
	shortURL := strings.TrimSpace(c.Params("shortUrl"))
	if shortURL == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "专栏短链接不能为空")
	}

	column, err := h.contentRepo.GetColumnByShortURL(c.Context(), shortURL)
	if err != nil {
		if errors.Is(err, content.ErrColumnNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "专栏不存在")
		}
		return err
	}

	query := buildMomentListQuery(c)
	query.ColumnID = &column.ID
	published := true
	query.Published = &published

	return h.listMomentsWithQuery(c, query)
}

// ListMomentsAdmin godoc
// @Summary 获取手记列表（管理员）
// @Tags Moment
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param columnId query int false "分区ID"
// @Param topicId query int false "话题ID"
// @Param authorId query int false "作者ID"
// @Param published query bool false "是否发布"
// @Param search query string false "搜索关键词"
// @Success 200 {object} contract.MomentListResp
// @Security BearerAuth
// @Router /admin/moments [get]
func (h *MomentHandler) ListMomentsAdmin(c *fiber.Ctx) error {
	query := buildMomentListQuery(c)
	return h.listMomentsWithQuery(c, query)
}

func (h *MomentHandler) listMomentsWithQuery(c *fiber.Ctx, query contract.ListMomentsReq) error {
	moments, total, err := h.svc.ListMoments(c.Context(), content.MomentListOptionsInternal(query))
	if err != nil {
		return err
	}

	momentResponses := make([]contract.MomentListItemResp, len(moments))
	for i, item := range moments {
		resp, err := h.toMomentListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		momentResponses[i] = *resp
	}

	listResponse := contract.MomentListResp{
		Items: momentResponses,
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
	}

	return response.Success(c, listResponse)
}

func buildMomentListQuery(c *fiber.Ctx) contract.ListMomentsReq {
	query := contract.ListMomentsReq{
		Page:     1,
		PageSize: 10,
	}

	if page, err := strconv.Atoi(c.Query("page", "1")); err == nil && page > 0 {
		query.Page = page
	}
	if pageSize, err := strconv.Atoi(c.Query("pageSize", "10")); err == nil && pageSize > 0 && pageSize <= 100 {
		query.PageSize = pageSize
	}
	if columnID, err := strconv.ParseInt(c.Query("columnId"), 10, 64); err == nil {
		query.ColumnID = &columnID
	}
	if topicID, err := strconv.ParseInt(c.Query("topicId"), 10, 64); err == nil {
		query.TopicID = &topicID
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

// ListRecentPublicMoments godoc
// @Summary 获取最近公开手记
// @Tags Public
// @Produce json
// @Success 200 {object} contract.MomentListResp
// @Router /public/moments/recent [get]
func (h *MomentHandler) ListRecentPublicMoments(c *fiber.Ctx) error {
	const page = 1
	const size = 5
	published := true

	moments, total, err := h.svc.ListMoments(c.Context(), content.MomentListOptionsInternal{
		Page:      page,
		PageSize:  size,
		Published: &published,
	})
	if err != nil {
		return err
	}

	momentResponses := make([]contract.MomentListItemResp, len(moments))
	for i, item := range moments {
		resp, err := h.toMomentListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		momentResponses[i] = *resp
	}

	return response.Success(c, contract.MomentListResp{
		Items: momentResponses,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// CheckMomentLatest godoc
// @Summary 校验手记是否最新
// @Tags Moment
// @Accept json
// @Produce json
// @Param id path int true "手记ID"
// @Param request body contract.CheckMomentLatestReq true "手记版本校验参数"
// @Success 200 {object} contract.CheckMomentLatestResp
// @Router /moments/{id}/latest [post]
func (h *MomentHandler) CheckMomentLatest(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	var req contract.CheckMomentLatestReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	momentItem, err := h.svc.GetMomentByID(c.Context(), id)
	if errors.Is(err, content.ErrMomentNotFound) {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	} else if err != nil {
		return err
	}
	if !momentItem.IsPublished {
		return response.NewBizErrorWithMsg(response.NotFound, "手记不存在")
	}

	if req.Hash == momentItem.ContentHash {
		return response.Success(c, contract.CheckMomentLatestResp{
			Latest: true,
			MomentContentPayload: contract.MomentContentPayload{
				ContentHash: momentItem.ContentHash,
			},
		})
	}

	return response.Success(c, contract.CheckMomentLatestResp{
		Latest: false,
		MomentContentPayload: contract.MomentContentPayload{
			ContentHash: momentItem.ContentHash,
			Title:       momentItem.Title,
			Summary:     momentItem.Summary,
			TOC:         mapMomentTOCNodes(momentItem.TOC),
			Content:     momentItem.Content,
		},
	})
}

// DeleteMoment godoc
// @Summary 删除手记
// @Tags Moment
// @Produce json
// @Param id path int true "手记ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /moments/{id} [delete]
// @Security JWTAuth
func (h *MomentHandler) DeleteMoment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的手记ID")
	}

	if err := h.svc.DeleteMoment(c.Context(), id); err != nil {
		return err
	}

	Audit(c, "moment.delete", map[string]any{
		"momentId": id,
		"userId":   claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "手记删除成功")
}

// BatchDeleteMoments godoc
// @Summary 批量删除手记（管理端）
// @Tags Moment
// @Accept json
// @Produce json
// @Param request body contract.BatchDeleteMomentReq true "批量删除参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/moments/batch-delete [post]
// @Security JWTAuth
func (h *MomentHandler) BatchDeleteMoments(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.BatchDeleteMomentReq
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

	if err := h.svc.BatchDelete(c.Context(), moment.BatchDeleteCmd{IDs: req.IDs}); err != nil {
		return err
	}

	Audit(c, "moment.batch_delete", map[string]any{
		"momentIds": req.IDs,
		"userId":    claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "手记批量删除成功")
}

func (h *MomentHandler) toMomentResp(ctx context.Context, momentItem *content.Moment) (*contract.MomentResp, error) {
	topics, err := h.svc.GetMomentTopics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	metrics, err := h.svc.GetMomentMetrics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	resp := contract.MomentResp{
		ID:           momentItem.ID,
		Title:        momentItem.Title,
		Summary:      momentItem.Summary,
		AISummary:    momentItem.AISummary,
		TOC:          mapMomentTOCNodes(momentItem.TOC),
		Content:      momentItem.Content,
		ContentHash:  momentItem.ContentHash,
		AuthorID:     momentItem.AuthorID,
		Image:        splitImages(momentItem.Image),
		ColumnID:     momentItem.ColumnID,
		CommentID:    momentItem.CommentID,
		ShortURL:     momentItem.ShortURL,
		IsPublished:  momentItem.IsPublished,
		IsTop:        momentItem.IsTop,
		IsHot:        momentItem.IsHot,
		AllowComment: h.allowCommentByAreaID(ctx, momentItem.CommentID),
		IsOriginal:   momentItem.IsOriginal,
		ExtInfo:      jsonRawFromBytes(momentItem.ExtInfo),
		CreatedAt:    momentItem.CreatedAt,
		UpdatedAt:    momentItem.UpdatedAt,
	}

	if momentItem.ColumnID != nil {
		column, colErr := h.contentRepo.GetColumnByID(ctx, *momentItem.ColumnID)
		if colErr == nil && column != nil {
			resp.ColumnName = column.Name
			if column.ShortURL != nil {
				resp.ColumnShortURL = *column.ShortURL
			}
		}
	}

	if len(topics) > 0 {
		resp.Topics = make([]contract.TagResp, len(topics))
		for i, topic := range topics {
			if err := copier.Copy(&resp.Topics[i], topic); err != nil {
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

func (h *MomentHandler) toMomentListItemResp(ctx context.Context, momentItem *content.Moment) (*contract.MomentListItemResp, error) {
	topics, err := h.svc.GetMomentTopics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	metrics, err := h.svc.GetMomentMetrics(ctx, momentItem.ID)
	if err != nil {
		return nil, err
	}

	resp := contract.MomentListItemResp{
		ID:           momentItem.ID,
		Title:        momentItem.Title,
		ShortURL:     momentItem.ShortURL,
		Summary:      momentItem.Summary,
		IsTop:        momentItem.IsTop,
		IsHot:        momentItem.IsHot,
		AllowComment: h.allowCommentByAreaID(ctx, momentItem.CommentID),
		IsOriginal:   momentItem.IsOriginal,
		IsPublished:  momentItem.IsPublished,
		CreatedAt:    momentItem.CreatedAt,
		UpdatedAt:    momentItem.UpdatedAt,
		Topics:       []string{},
		Image:        splitImages(momentItem.Image),
	}
	resp.CommentID = momentItem.CommentID

	if metrics != nil {
		resp.Views = metrics.Views
		resp.Likes = metrics.Likes
		resp.Comments = metrics.Comments
	}

	if len(topics) > 0 {
		topicNames := make([]string, len(topics))
		for i, topic := range topics {
			topicNames[i] = topic.Name
		}
		resp.Topics = topicNames
	}

	if momentItem.ColumnID != nil {
		column, err := h.contentRepo.GetColumnByID(ctx, *momentItem.ColumnID)
		if err != nil {
			if !errors.Is(err, content.ErrColumnNotFound) {
				return nil, err
			}
		} else if column != nil {
			resp.ColumnName = column.Name
			if column.ShortURL != nil {
				resp.ColumnShortURL = *column.ShortURL
			}
		}
	}

	if h.userRepo != nil {
		user, err := h.userRepo.FindByID(ctx, momentItem.AuthorID)
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

func mapMomentTOCNodes(nodes []content.TOCNode) []contract.TOCNode {
	result := make([]contract.TOCNode, len(nodes))
	for i, node := range nodes {
		result[i] = contract.TOCNode{
			Name:     node.Name,
			Anchor:   node.Anchor,
			Children: mapMomentTOCNodes(node.Children),
		}
	}
	return result
}

func splitImages(input *string) []string {
	if input == nil {
		return []string{}
	}
	trimmed := strings.TrimSpace(*input)
	if trimmed == "" {
		return []string{}
	}
	parts := strings.Split(trimmed, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func joinImages(images []string) *string {
	if len(images) == 0 {
		return nil
	}
	out := make([]string, 0, len(images))
	for _, img := range images {
		item := strings.TrimSpace(img)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	joined := strings.Join(out, ",")
	return &joined
}

func (h *MomentHandler) allowCommentByAreaID(ctx context.Context, areaID *int64) bool {
	if h.commentRepo == nil || areaID == nil || *areaID <= 0 {
		return true
	}
	area, err := h.commentRepo.GetAreaByID(ctx, *areaID)
	if err != nil || area == nil {
		return false
	}
	return !area.IsClosed
}
