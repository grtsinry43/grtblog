package handler

import (
	"context"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/thinking"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	domainthinking "github.com/grtsinry43/grtblog-v2/server/internal/domain/thinking"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type ThinkingHandler struct {
	svc         *thinking.Service
	commentRepo domaincomment.CommentRepository
	userRepo    identity.Repository
}

func NewThinkingHandler(svc *thinking.Service, commentRepo domaincomment.CommentRepository, userRepo identity.Repository) *ThinkingHandler {
	return &ThinkingHandler{
		svc:         svc,
		commentRepo: commentRepo,
		userRepo:    userRepo,
	}
}

// CreateThinking godoc
// @Summary 发布思考
// @Tags Thinking
// @Accept json
// @Produce json
// @Param request body contract.CreateThinkingReq true "创建参数"
// @Success 200 {object} contract.ThinkingResp
// @Security JWTAuth
// @Router /thinkings [post]
func (h *ThinkingHandler) CreateThinking(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.CreateThinkingReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	created, err := h.svc.Create(c.Context(), thinking.CreateThinkingCmd{
		Content:      req.Content,
		AuthorID:     claims.UserID,
		AllowComment: req.AllowComment,
	})
	if err != nil {
		return h.mapError(c, err)
	}

	Audit(c, "thinking.create", map[string]any{
		"thinkingId":     created.ID,
		"thinkingAuthor": created.AuthorID,
		"userId":         claims.UserID,
	})

	resp, err := h.toThinkingResp(c.Context(), created)
	if err != nil {
		return err
	}
	return response.Success(c, resp)
}

// UpdateThinking godoc
// @Summary 更新思考
// @Tags Thinking
// @Accept json
// @Produce json
// @Param id path int true "思考ID"
// @Param request body contract.UpdateThinkingReq true "更新参数"
// @Success 200
// @Security JWTAuth
// @Router /thinkings/{id} [put]
func (h *ThinkingHandler) UpdateThinking(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的思考ID")
	}

	var req contract.UpdateThinkingReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "无效的请求体", err)
	}

	_, err = h.svc.Update(c.Context(), thinking.UpdateThinkingCmd{
		ID:           id,
		Content:      req.Content,
		AllowComment: req.AllowComment,
	})
	if err != nil {
		return h.mapError(c, err)
	}

	Audit(c, "thinking.update", map[string]any{
		"thinkingId": id,
		"userId":     claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "更新思考成功")
}

// ListThinkings godoc
// @Summary 获取思考列表
// @Tags Thinking
// @Param page query int false "页码"
// @Param pageSize query int false "页大小"
// @Success 200 {object} contract.ListThinkingResp
// @Router /thinkings [get]
func (h *ThinkingHandler) ListThinkings(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	items, total, err := h.svc.List(c.Context(), pageSize, offset)
	if err != nil {
		return h.mapError(c, err)
	}

	resItems := make([]contract.ThinkingListItemResp, len(items))
	for i, item := range items {
		resp, err := h.toThinkingListItemResp(c.Context(), item)
		if err != nil {
			return err
		}
		resItems[i] = *resp
	}

	return response.Success(c, contract.ListThinkingResp{
		Items: resItems,
		Total: total,
	})
}

// GetThinking godoc
// @Summary 获取单个思考
// @Tags Thinking
// @Param id path int true "思考ID"
// @Success 200 {object} contract.ThinkingResp
// @Router /thinkings/{id} [get]
func (h *ThinkingHandler) GetThinking(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的ID")
	}

	item, err := h.svc.FindByID(c.Context(), id)
	if err != nil {
		return h.mapError(c, err)
	}

	resp, err := h.toThinkingResp(c.Context(), item)
	if err != nil {
		return err
	}
	return response.Success(c, resp)
}

// DeleteThinking godoc
// @Summary 删除思考
// @Tags Thinking
// @Param id path int true "思考ID"
// @Success 200
// @Security JWTAuth
// @Router /thinkings/{id} [delete]
func (h *ThinkingHandler) DeleteThinking(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的ID")
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return h.mapError(c, err)
	}

	Audit(c, "thinking.delete", map[string]any{
		"thinkingId": id,
		"userId":     claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "思考删除成功")
}

// BatchDeleteThinkings godoc
// @Summary 批量删除思考（管理端）
// @Tags Thinking
// @Accept json
// @Produce json
// @Param request body contract.BatchDeleteThinkingReq true "批量删除参数"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security JWTAuth
// @Router /admin/thinkings/batch-delete [post]
func (h *ThinkingHandler) BatchDeleteThinkings(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req contract.BatchDeleteThinkingReq
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

	if err := h.svc.BatchDelete(c.Context(), thinking.BatchDeleteCmd{IDs: req.IDs}); err != nil {
		return h.mapError(c, err)
	}

	Audit(c, "thinking.batch_delete", map[string]any{
		"thinkingIds": req.IDs,
		"userId":      claims.UserID,
	})

	return response.SuccessWithMessage[any](c, nil, "思考批量删除成功")
}

func (h *ThinkingHandler) mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domainthinking.ErrThinkingNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "思考不存在")
	case errors.Is(err, domainthinking.ErrThinkingContentEmpty):
		return response.NewBizErrorWithMsg(response.ParamsError, "内容不能为空")
	default:
		return err
	}
}

func (h *ThinkingHandler) toThinkingResp(ctx context.Context, t *domainthinking.Thinking) (*contract.ThinkingResp, error) {
	resp := &contract.ThinkingResp{
		ID:           t.ID,
		CommentID:    t.CommentID,
		Content:      t.Content,
		AuthorID:     t.AuthorID,
		IsHot:        false,
		AllowComment: h.allowCommentByAreaID(ctx, t.CommentID),
		Metrics: contract.ThinkingMetrics{
			Views:    t.Metrics.Views,
			Likes:    t.Metrics.Likes,
			Comments: t.Metrics.Comments,
		},
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	if h.userRepo != nil {
		user, err := h.userRepo.FindByID(ctx, t.AuthorID)
		if err != nil {
			if !errors.Is(err, identity.ErrUserNotFound) {
				return nil, err
			}
		} else if user != nil {
			resp.AuthorName = user.Nickname
			resp.Avatar = user.Avatar
		}
	}
	return resp, nil
}

func (h *ThinkingHandler) toThinkingListItemResp(ctx context.Context, t *domainthinking.Thinking) (*contract.ThinkingListItemResp, error) {
	resp := &contract.ThinkingListItemResp{
		ID:           t.ID,
		CommentID:    t.CommentID,
		Content:      t.Content,
		AuthorID:     t.AuthorID,
		IsHot:        false,
		AllowComment: h.allowCommentByAreaID(ctx, t.CommentID),
		Views:        t.Metrics.Views,
		Likes:        t.Metrics.Likes,
		Comments:     t.Metrics.Comments,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
	if h.userRepo != nil {
		user, err := h.userRepo.FindByID(ctx, t.AuthorID)
		if err != nil {
			if !errors.Is(err, identity.ErrUserNotFound) {
				return nil, err
			}
		} else if user != nil {
			resp.AuthorName = user.Nickname
			resp.Avatar = user.Avatar
		}
	}
	return resp, nil
}

func (h *ThinkingHandler) allowCommentByAreaID(ctx context.Context, areaID int64) bool {
	if h.commentRepo == nil || areaID <= 0 {
		return true
	}
	area, err := h.commentRepo.GetAreaByID(ctx, areaID)
	if err != nil || area == nil {
		return false
	}
	return !area.IsClosed
}
