package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/comment"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/security/jwt"
)

type CommentHandler struct {
	svc        *comment.Service
	jwtManager *jwt.Manager
}

func NewCommentHandler(svc *comment.Service, jwtManager *jwt.Manager) *CommentHandler {
	return &CommentHandler{svc: svc, jwtManager: jwtManager}
}

// CreateCommentLogin godoc
// @Summary 创建评论（登录用户）
// @Tags Comment
// @Accept json
// @Produce json
// @Param areaId path int true "评论区ID"
// @Param request body contract.CreateCommentLoginReq true "创建评论参数"
// @Success 200 {object} contract.CreateCommentResp
// @Security JWTAuth
// @Router /comments/areas/{areaId} [post]
func (h *CommentHandler) CreateCommentLogin(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	areaID, err := parseInt64Param(c, "areaId")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论区ID")
	}

	var req contract.CreateCommentLoginReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	var cmd comment.CreateCommentLoginCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}
	cmd.AreaID = areaID
	cmd.VisitorID = strings.TrimSpace(req.VisitorID)

	meta := comment.RequestMeta{
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent", ""),
	}
	created, err := h.svc.CreateCommentLogin(c.Context(), claims.UserID, cmd, meta)
	if err != nil {
		return h.mapCommentError(c, err)
	}
	resp := toCreateCommentResp(created)
	return response.SuccessWithMessage(c, resp, "评论创建成功")
}

// CreateCommentVisitor godoc
// @Summary 创建评论（访客）
// @Tags Comment
// @Accept json
// @Produce json
// @Param areaId path int true "评论区ID"
// @Param request body contract.CreateCommentVisitorReq true "创建评论参数"
// @Success 200 {object} contract.CreateCommentResp
// @Router /comments/areas/{areaId}/visitor [post]
func (h *CommentHandler) CreateCommentVisitor(c *fiber.Ctx) error {
	areaID, err := parseInt64Param(c, "areaId")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论区ID")
	}

	var req contract.CreateCommentVisitorReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	cmd := comment.CreateCommentVisitorCmd{
		AreaID:   areaID,
		Content:  req.Content,
		ParentID: req.ParentID,
	}
	if req.NickName != nil {
		cmd.NickName = *req.NickName
	}
	if req.Email != nil {
		cmd.Email = *req.Email
	}
	cmd.Website = req.Website
	cmd.VisitorID = strings.TrimSpace(req.VisitorID)

	meta := comment.RequestMeta{
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent", ""),
	}
	created, err := h.svc.CreateCommentVisitor(c.Context(), cmd, meta)
	if err != nil {
		return h.mapCommentError(c, err)
	}
	resp := toCreateCommentResp(created)
	return response.SuccessWithMessage(c, resp, "评论创建成功")
}

// ListCommentTree godoc
// @Summary 获取评论树（公开，根评论分页）
// @Tags Comment
// @Produce json
// @Param areaId path int true "评论区ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param size query int false "每页数量（兼容参数）"
// @Param visitorId query string false "访客ID（使用前端埋点 visitorId）"
// @Success 200 {object} contract.PublicCommentListResp
// @Router /comments/areas/{areaId} [get]
func (h *CommentHandler) ListCommentTree(c *fiber.Ctx) error {
	areaID, err := parseInt64Param(c, "areaId")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论区ID")
	}

	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 0)
	if pageSize <= 0 {
		pageSize = parseIntQuery(c, "size", 10)
	}

	viewerVisitorID := strings.TrimSpace(c.Query("visitorId"))
	viewerAuthorID := h.resolveViewerAuthorID(c)

	result, err := h.svc.ListPublicComments(c.Context(), comment.ListPublicCommentsCmd{
		AreaID:          areaID,
		Page:            page,
		PageSize:        pageSize,
		ViewerAuthorID:  viewerAuthorID,
		ViewerVisitorID: viewerVisitorID,
	})
	if err != nil {
		return h.mapCommentError(c, err)
	}

	respItems := make([]contract.CommentNodeResp, len(result.Items))
	for i, node := range result.Items {
		respItems[i] = toCommentNodeResp(node)
	}
	return response.Success(c, contract.PublicCommentListResp{
		Items:             respItems,
		Total:             result.Total,
		Page:              result.Page,
		Size:              result.Size,
		IsClosed:          result.IsClosed,
		RequireModeration: result.RequireModeration,
	})
}

// ListAdminComments godoc
// @Summary 获取评论列表（管理端）
// @Tags CommentAdmin
// @Produce json
// @Param areaId query int false "评论区ID"
// @Param status query string false "状态 pending/approved/rejected/blocked"
// @Param onlyUnviewed query bool false "仅返回未查看" default(true)
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.AdminCommentListResp
// @Security JWTAuth
// @Router /admin/comments [get]
func (h *CommentHandler) ListAdminComments(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 20)
	status := strings.TrimSpace(c.Query("status"))
	if status != "" && !isValidCommentStatus(status) {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论状态")
	}
	onlyUnviewed := parseBoolQuery(c, "onlyUnviewed", true)

	var areaID *int64
	if raw := strings.TrimSpace(c.Query("areaId")); raw != "" {
		val, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论区ID")
		}
		areaID = &val
	}

	items, total, err := h.svc.ListAdminComments(c.Context(), comment.ListAdminCommentsCmd{
		AreaID:       areaID,
		Status:       status,
		OnlyUnviewed: onlyUnviewed,
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		return h.mapCommentError(c, err)
	}

	respItems := make([]contract.AdminCommentResp, len(items))
	for i := range items {
		respItems[i] = toAdminCommentResp(items[i])
	}
	return response.Success(c, contract.AdminCommentListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// ListAdminVisitors godoc
// @Summary 获取访客画像列表（管理端）
// @Tags CommentAdmin
// @Produce json
// @Param keyword query string false "关键词（visitorId/昵称/邮箱/IP/地区/设备）"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.AdminVisitorListResp
// @Security JWTAuth
// @Router /admin/visitors [get]
func (h *CommentHandler) ListAdminVisitors(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 20)
	keyword := strings.TrimSpace(c.Query("keyword"))

	items, total, err := h.svc.ListAdminVisitors(c.Context(), comment.ListAdminVisitorsCmd{
		Keyword:  keyword,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return h.mapCommentError(c, err)
	}

	respItems := make([]contract.AdminVisitorResp, len(items))
	for i := range items {
		respItems[i] = toAdminVisitorResp(items[i])
	}
	return response.Success(c, contract.AdminVisitorListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// GetAdminVisitorProfile godoc
// @Summary 获取访客画像详情（管理端）
// @Tags CommentAdmin
// @Produce json
// @Param visitorId path string true "访客ID"
// @Param recentLimit query int false "最近评论数量" default(20)
// @Success 200 {object} contract.AdminVisitorProfileResp
// @Security JWTAuth
// @Router /admin/visitors/{visitorId} [get]
func (h *CommentHandler) GetAdminVisitorProfile(c *fiber.Ctx) error {
	visitorID := strings.TrimSpace(c.Params("visitorId"))
	recentLimit := parseIntQuery(c, "recentLimit", 20)

	profile, recentComments, err := h.svc.GetVisitorProfile(c.Context(), comment.GetVisitorProfileCmd{
		VisitorID:   visitorID,
		RecentLimit: recentLimit,
	})
	if err != nil {
		return h.mapCommentError(c, err)
	}

	recentItems := make([]contract.AdminVisitorRecentCommentResp, 0, len(recentComments))
	for _, item := range recentComments {
		recentItems = append(recentItems, contract.AdminVisitorRecentCommentResp{
			ID:        strconv.FormatInt(item.ID, 10),
			AreaID:    item.AreaID,
			Content:   item.Content,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
			IsDeleted: item.IsDeleted,
		})
	}

	return response.Success(c, contract.AdminVisitorProfileResp{
		Profile:        toAdminVisitorResp(*profile),
		RecentComments: recentItems,
	})
}

// GetAdminVisitorInsights godoc
// @Summary 获取访客画像统计图数据（管理端）
// @Tags CommentAdmin
// @Produce json
// @Param days query int false "统计天数（默认30，最大180）"
// @Success 200 {object} contract.AdminVisitorInsightsResp
// @Security JWTAuth
// @Router /admin/visitors/insights [get]
func (h *CommentHandler) GetAdminVisitorInsights(c *fiber.Ctx) error {
	days := parseIntQuery(c, "days", 30)
	insights, err := h.svc.GetVisitorInsights(c.Context(), comment.GetVisitorInsightsCmd{Days: days})
	if err != nil {
		return h.mapCommentError(c, err)
	}
	return response.Success(c, toAdminVisitorInsightsResp(insights))
}

// MarkCommentsViewed godoc
// @Summary 批量标记评论已读/未读（管理端）
// @Tags CommentAdmin
// @Accept json
// @Produce json
// @Param request body contract.MarkCommentsViewedReq true "批量已读参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security JWTAuth
// @Router /admin/comments/viewed [put]
func (h *CommentHandler) MarkCommentsViewed(c *fiber.Ctx) error {
	var req contract.MarkCommentsViewedReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.IDs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "ids 不能为空")
	}
	ids := make([]int64, 0, len(req.IDs))
	for _, raw := range req.IDs {
		id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		if err != nil || id <= 0 {
			return response.NewBizErrorWithMsg(response.ParamsError, "ids 包含无效值")
		}
		ids = append(ids, id)
	}
	isViewed := true
	if req.IsViewed != nil {
		isViewed = *req.IsViewed
	}
	if err := h.svc.MarkCommentsViewed(c.Context(), comment.MarkCommentsViewedCmd{
		IDs:      ids,
		IsViewed: isViewed,
	}); err != nil {
		return h.mapCommentError(c, err)
	}
	if isViewed {
		return response.SuccessWithMessage[any](c, nil, "已标记已读")
	}
	return response.SuccessWithMessage[any](c, nil, "已标记未读")
}

// ImportComment godoc
// @Summary 导入评论（管理端，全字段）
// @Tags CommentAdmin
// @Accept json
// @Produce json
// @Param request body contract.ImportCommentReq true "评论导入参数"
// @Success 200 {object} contract.CreateCommentResp
// @Security JWTAuth
// @Router /admin/comments/import [post]
func (h *CommentHandler) ImportComment(c *fiber.Ctx) error {
	var req contract.ImportCommentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	var cmd comment.ImportCommentCmd
	if err := copier.Copy(&cmd, req); err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "请求体映射失败")
	}

	created, err := h.svc.ImportComment(c.Context(), cmd)
	if err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage(c, toCreateCommentResp(created), "评论导入成功")
}

// ReplyComment godoc
// @Summary 快捷回复评论（管理端）
// @Tags CommentAdmin
// @Accept json
// @Produce json
// @Param id path int true "父评论ID"
// @Param request body contract.ReplyCommentReq true "回复参数"
// @Success 200 {object} contract.CreateCommentResp
// @Security JWTAuth
// @Router /admin/comments/{id}/reply [post]
func (h *CommentHandler) ReplyComment(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	parentID, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论ID")
	}
	var req contract.ReplyCommentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	reply, err := h.svc.ReplyComment(c.Context(), comment.ReplyCommentCmd{
		ParentID: parentID,
		Content:  req.Content,
		AdminID:  claims.UserID,
	})
	if err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage(c, toCreateCommentResp(reply), "回复成功")
}

// UpdateCommentStatus godoc
// @Summary 更新评论状态（管理端）
// @Tags CommentAdmin
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param request body contract.UpdateCommentStatusReq true "状态参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security JWTAuth
// @Router /admin/comments/{id}/status [put]
func (h *CommentHandler) UpdateCommentStatus(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论ID")
	}
	var req contract.UpdateCommentStatusReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.svc.UpdateCommentStatus(c.Context(), comment.UpdateCommentStatusCmd{
		ID:     id,
		Status: req.Status,
	}); err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage[any](c, nil, "评论状态已更新")
}

// SetCommentAuthor godoc
// @Summary 标记评论为作者评论（管理端）
// @Tags CommentAdmin
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param request body contract.SetCommentAuthorReq true "作者标记参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security JWTAuth
// @Router /admin/comments/{id}/author [put]
func (h *CommentHandler) SetCommentAuthor(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论ID")
	}
	var req contract.SetCommentAuthorReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.svc.SetCommentAuthor(c.Context(), comment.SetCommentAuthorCmd{
		ID:       id,
		IsAuthor: req.IsAuthor,
	}); err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage[any](c, nil, "评论作者标记已更新")
}

// SetCommentTop godoc
// @Summary 置顶/取消置顶评论（管理端）
// @Tags CommentAdmin
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param request body contract.SetCommentTopReq true "置顶参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security JWTAuth
// @Router /admin/comments/{id}/top [put]
func (h *CommentHandler) SetCommentTop(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论ID")
	}
	var req contract.SetCommentTopReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.svc.SetCommentTop(c.Context(), comment.SetCommentTopCmd{
		ID:    id,
		IsTop: req.IsTop,
	}); err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage[any](c, nil, "评论置顶状态已更新")
}

// DeleteComment godoc
// @Summary 删除评论（软删除，管理端）
// @Tags CommentAdmin
// @Produce json
// @Param id path int true "评论ID"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security JWTAuth
// @Router /admin/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论ID")
	}
	if err := h.svc.DeleteComment(c.Context(), id); err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage[any](c, nil, "评论已删除")
}

// EditOwnComment godoc
// @Summary 编辑自己的评论
// @Tags Comment
// @Accept json
// @Produce json
// @Param id path int true "评论ID"
// @Param request body contract.UpdateCommentReq true "编辑评论参数"
// @Success 200 {object} contract.CreateCommentResp
// @Router /comments/{id} [put]
func (h *CommentHandler) EditOwnComment(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论ID")
	}
	var req contract.UpdateCommentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	viewerAuthorID := h.resolveViewerAuthorID(c)
	updated, err := h.svc.EditComment(c.Context(), comment.EditCommentCmd{
		ID:              id,
		Content:         req.Content,
		ViewerAuthorID:  viewerAuthorID,
		ViewerVisitorID: strings.TrimSpace(req.VisitorID),
	})
	if err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage(c, toCreateCommentResp(updated), "评论已更新")
}

// DeleteOwnComment godoc
// @Summary 删除自己的评论
// @Tags Comment
// @Produce json
// @Param id path int true "评论ID"
// @Router /comments/{id} [delete]
func (h *CommentHandler) DeleteOwnComment(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论ID")
	}
	var req contract.DeleteOwnCommentReq
	// Body is optional for DELETE — visitors send visitorId in body
	_ = c.BodyParser(&req)
	viewerAuthorID := h.resolveViewerAuthorID(c)
	if err := h.svc.DeleteOwnComment(c.Context(), comment.DeleteOwnCommentCmd{
		ID:              id,
		ViewerAuthorID:  viewerAuthorID,
		ViewerVisitorID: strings.TrimSpace(req.VisitorID),
	}); err != nil {
		return h.mapCommentError(c, err)
	}
	return response.SuccessWithMessage[any](c, nil, "评论已删除")
}

// SetCommentAreaClose godoc
// @Summary 关闭/开启评论区（管理端）
// @Tags CommentAdmin
// @Accept json
// @Produce json
// @Param areaId path int true "评论区ID"
// @Param request body contract.SetCommentAreaCloseReq true "评论区开关参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security JWTAuth
// @Router /admin/comments/areas/{areaId}/close [put]
func (h *CommentHandler) SetCommentAreaClose(c *fiber.Ctx) error {
	areaID, err := parseInt64Param(c, "areaId")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论区ID")
	}
	var req contract.SetCommentAreaCloseReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.svc.SetAreaClosed(c.Context(), areaID, req.IsClosed); err != nil {
		return h.mapCommentError(c, err)
	}
	if req.IsClosed {
		return response.SuccessWithMessage[any](c, nil, "评论区已关闭")
	}
	return response.SuccessWithMessage[any](c, nil, "评论区已开启")
}

func (h *CommentHandler) mapCommentError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domaincomment.ErrCommentAreaNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "评论区不存在")
	case errors.Is(err, domaincomment.ErrCommentNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "评论不存在")
	case errors.Is(err, domaincomment.ErrCommentParentNotFound):
		return response.NewBizErrorWithMsg(response.ParamsError, "父评论不存在")
	case errors.Is(err, domaincomment.ErrCommentTooDeep):
		return response.NewBizErrorWithMsg(response.ParamsError, "评论层级过深")
	case errors.Is(err, domaincomment.ErrCommentContentEmpty):
		return response.NewBizErrorWithMsg(response.ParamsError, "评论内容不能为空")
	case errors.Is(err, domaincomment.ErrCommentContentTooLong):
		return response.NewBizErrorWithMsg(response.ParamsError, "评论内容不能超过500字")
	case errors.Is(err, domaincomment.ErrCommentAreaClosed):
		return response.NewBizErrorWithMsg(response.ParamsError, "评论区已关闭")
	case errors.Is(err, domaincomment.ErrCommentDisabled):
		return response.NewBizErrorWithMsg(response.ParamsError, "全站已关闭评论")
	case errors.Is(err, domaincomment.ErrCommentBlocked):
		return response.NewBizErrorWithMsg(response.Unauthorized, "你已被禁止评论")
	case errors.Is(err, domaincomment.ErrCommentStatusInvalid):
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的评论状态")
	case errors.Is(err, domaincomment.ErrVisitorNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "访客不存在")
	case errors.Is(err, domaincomment.ErrCommentReplyDisabled):
		return response.NewBizErrorWithMsg(response.ParamsError, "该评论仅支持联邦回复")
	case errors.Is(err, domaincomment.ErrCommentNotOwner):
		return response.NewBizErrorWithMsg(response.Unauthorized, "无权操作此评论")
	case errors.Is(err, domaincomment.ErrCommentAlreadyDeleted):
		return response.NewBizErrorWithMsg(response.ParamsError, "评论已被删除")
	default:
		return err
	}
}

func parseInt64Param(c *fiber.Ctx, name string) (int64, error) {
	return strconv.ParseInt(c.Params(name), 10, 64)
}

func parseBoolQuery(c *fiber.Ctx, key string, defaultVal bool) bool {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultVal
	}
	return val
}

func (h *CommentHandler) resolveViewerAuthorID(c *fiber.Ctx) *int64 {
	if claims, ok := middleware.GetClaims(c); ok && claims != nil && claims.UserID > 0 {
		id := claims.UserID
		return &id
	}

	if h.jwtManager == nil {
		return nil
	}

	token := extractBearerToken(c.Get("Authorization"))
	if token == "" || strings.HasPrefix(token, "gt_") {
		return nil
	}

	claims, err := h.jwtManager.Parse(token)
	if err != nil || claims == nil || claims.UserID <= 0 {
		return nil
	}
	id := claims.UserID
	return &id
}

func extractBearerToken(header string) string {
	raw := strings.TrimSpace(header)
	if raw == "" {
		return ""
	}
	parts := strings.SplitN(raw, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func toCreateCommentResp(entity *domaincomment.Comment) contract.CreateCommentResp {
	return contract.CreateCommentResp{
		ID:                entity.ID,
		AreaID:            entity.AreaID,
		Content:           entity.Content,
		NickName:          entity.NickName,
		Avatar:            entity.Avatar,
		Location:          entity.Location,
		Platform:          entity.Platform,
		Browser:           entity.Browser,
		Website:           entity.Website,
		IsOwner:           entity.IsOwner,
		IsFriend:          entity.IsFriend,
		IsAuthor:          entity.IsAuthor,
		IsViewed:          entity.IsViewed,
		IsTop:             entity.IsTop,
		IsMy:              entity.IsMy,
		IsFederated:       entity.IsFederated,
		FederatedProtocol: entity.FederatedProtocol,
		FederatedActor:    entity.FederatedActor,
		CanReply:          entity.CanReply,
		Status:            entity.Status,
		IsEdited:          entity.IsEdited,
		ParentID:          entity.ParentID,
		CreatedAt:         entity.CreatedAt,
		UpdatedAt:         entity.UpdatedAt,
		DeletedAt:         entity.DeletedAt,
		IsDeleted:         entity.DeletedAt != nil,
	}
}

func toCommentNodeResp(node *comment.CommentNode) contract.CommentNodeResp {
	resp := contract.CommentNodeResp{
		ID:                node.Comment.ID,
		AreaID:            node.Comment.AreaID,
		Floor:             node.Floor,
		Content:           publicContent(node.Comment),
		NickName:          node.Comment.NickName,
		Avatar:            node.Comment.Avatar,
		Location:          node.Comment.Location,
		Platform:          node.Comment.Platform,
		Browser:           node.Comment.Browser,
		Website:           node.Comment.Website,
		IsOwner:           node.Comment.IsOwner,
		IsFriend:          node.Comment.IsFriend,
		IsAuthor:          node.Comment.IsAuthor,
		IsViewed:          node.Comment.IsViewed,
		IsTop:             node.Comment.IsTop,
		IsMy:              node.Comment.IsMy,
		IsFederated:       node.Comment.IsFederated,
		FederatedProtocol: node.Comment.FederatedProtocol,
		FederatedActor:    node.Comment.FederatedActor,
		CanReply:          node.Comment.CanReply,
		Status:            node.Comment.Status,
		IsEdited:          node.Comment.IsEdited,
		ParentID:          node.Comment.ParentID,
		CreatedAt:         node.Comment.CreatedAt,
		UpdatedAt:         node.Comment.UpdatedAt,
		DeletedAt:         node.Comment.DeletedAt,
		IsDeleted:         node.Comment.DeletedAt != nil,
	}
	if len(node.Children) > 0 {
		resp.Children = make([]contract.CommentNodeResp, len(node.Children))
		for i, child := range node.Children {
			resp.Children[i] = toCommentNodeResp(child)
		}
	}
	return resp
}

func toAdminCommentResp(item *domaincomment.Comment) contract.AdminCommentResp {
	var content *string
	if item.DeletedAt == nil {
		content = toStringPtr(item.Content)
	}
	return contract.AdminCommentResp{
		ID:                strconv.FormatInt(item.ID, 10),
		AreaID:            item.AreaID,
		AreaType:          item.AreaType,
		AreaRefID:         item.AreaRefID,
		AreaName:          item.AreaName,
		AreaTitle:         item.AreaTitle,
		AreaClosed:        item.AreaClosed,
		Content:           content,
		AuthorID:          item.AuthorID,
		NickName:          item.NickName,
		Avatar:            item.Avatar,
		Email:             item.Email,
		IP:                item.IP,
		Location:          item.Location,
		Platform:          item.Platform,
		Browser:           item.Browser,
		Website:           item.Website,
		IsOwner:           item.IsOwner,
		IsFriend:          item.IsFriend,
		IsAuthor:          item.IsAuthor,
		IsViewed:          item.IsViewed,
		IsTop:             item.IsTop,
		IsFederated:       item.IsFederated,
		FederatedProtocol: item.FederatedProtocol,
		FederatedActor:    item.FederatedActor,
		CanReply:          item.CanReply,
		Status:            item.Status,
		IsEdited:          item.IsEdited,
		ParentID:          int64PtrToString(item.ParentID),
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
		DeletedAt:         item.DeletedAt,
		IsDeleted:         item.DeletedAt != nil,
	}
}

func toAdminVisitorResp(item domaincomment.VisitorProfile) contract.AdminVisitorResp {
	return contract.AdminVisitorResp{
		VisitorID:        item.VisitorID,
		NickName:         item.NickName,
		Email:            item.Email,
		Website:          item.Website,
		IP:               item.IP,
		Location:         item.Location,
		Platform:         item.Platform,
		Browser:          item.Browser,
		TotalComments:    item.TotalComments,
		ApprovedComments: item.ApprovedComments,
		PendingComments:  item.PendingComments,
		RejectedComments: item.RejectedComments,
		BlockedComments:  item.BlockedComments,
		DeletedComments:  item.DeletedComments,
		TopComments:      item.TopComments,
		ActiveDays:       item.ActiveDays,
		TotalLikes:       item.TotalLikes,
		UniqueLikedItems: item.UniqueLikedItems,
		TotalViews:       item.TotalViews,
		UniqueViewItems:  item.UniqueViewItems,
		FirstSeenAt:      item.FirstSeenAt,
		LastSeenAt:       item.LastSeenAt,
		LastLikedAt:      item.LastLikedAt,
		LastViewedAt:     item.LastViewedAt,
	}
}

func toAdminVisitorInsightsResp(item *domaincomment.VisitorInsights) contract.AdminVisitorInsightsResp {
	if item == nil {
		return contract.AdminVisitorInsightsResp{}
	}
	platformTop := make([]contract.AdminVisitorDistributionResp, 0, len(item.PlatformTop))
	for _, row := range item.PlatformTop {
		platformTop = append(platformTop, contract.AdminVisitorDistributionResp{
			Name:  row.Name,
			Count: row.Count,
		})
	}
	browserTop := make([]contract.AdminVisitorDistributionResp, 0, len(item.BrowserTop))
	for _, row := range item.BrowserTop {
		browserTop = append(browserTop, contract.AdminVisitorDistributionResp{
			Name:  row.Name,
			Count: row.Count,
		})
	}
	locationTop := make([]contract.AdminVisitorDistributionResp, 0, len(item.LocationTop))
	for _, row := range item.LocationTop {
		locationTop = append(locationTop, contract.AdminVisitorDistributionResp{
			Name:  row.Name,
			Count: row.Count,
		})
	}
	trend := make([]contract.AdminVisitorTrendResp, 0, len(item.Trend))
	for _, row := range item.Trend {
		trend = append(trend, contract.AdminVisitorTrendResp{
			Date:              row.Date,
			ActiveVisitors:    row.ActiveVisitors,
			NewVisitors:       row.NewVisitors,
			ReturningVisitors: row.ReturningVisitors,
			Views:             row.Views,
			Likes:             row.Likes,
			Comments:          row.Comments,
		})
	}

	return contract.AdminVisitorInsightsResp{
		Days:        item.Days,
		GeneratedAt: item.GeneratedAt,
		DataSource:  item.DataSource,
		PlatformTop: platformTop,
		BrowserTop:  browserTop,
		LocationTop: locationTop,
		Trend:       trend,
		Funnel: contract.AdminVisitorFunnelResp{
			ViewVisitors:      item.Funnel.ViewVisitors,
			LikeVisitors:      item.Funnel.LikeVisitors,
			CommentVisitors:   item.Funnel.CommentVisitors,
			LikeRate:          item.Funnel.LikeRate,
			CommentRateByView: item.Funnel.CommentRateByView,
			CommentRateByLike: item.Funnel.CommentRateByLike,
		},
		Segments: contract.AdminVisitorSegmentsResp{
			Active1D:      item.Segments.Active1D,
			Active3D:      item.Segments.Active3D,
			Active7D:      item.Segments.Active7D,
			Active30D:     item.Segments.Active30D,
			HighlyEngaged: item.Segments.HighlyEngaged,
		},
	}
}

func publicContent(item *domaincomment.Comment) *string {
	if item.DeletedAt != nil {
		return nil
	}
	return toStringPtr(item.Content)
}

func toStringPtr(v string) *string {
	val := strings.TrimSpace(v)
	if val == "" {
		return nil
	}
	return &val
}

func int64PtrToString(v *int64) *string {
	if v == nil {
		return nil
	}
	value := strconv.FormatInt(*v, 10)
	return &value
}

func isValidCommentStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case domaincomment.CommentStatusPending,
		domaincomment.CommentStatusApproved,
		domaincomment.CommentStatusRejected,
		domaincomment.CommentStatusBlocked:
		return true
	default:
		return false
	}
}
