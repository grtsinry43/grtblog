package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/social"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type FriendLinkAdminHandler struct {
	svc *friendlink.AdminService
}

func NewFriendLinkAdminHandler(svc *friendlink.AdminService) *FriendLinkAdminHandler {
	return &FriendLinkAdminHandler{svc: svc}
}

// ListApplications godoc
// @Summary 获取友链申请列表
// @Tags FriendLinkAdmin
// @Produce json
// @Param status query string false "状态 pending/approved/rejected/blocked"
// @Param channel query string false "申请渠道 user/federation"
// @Param keyword query string false "关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} contract.FriendLinkApplicationListResp
// @Security BearerAuth
// @Router /admin/friend-links/applications [get]
func (h *FriendLinkAdminHandler) ListApplications(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	items, total, err := h.svc.ListApplications(c.Context(), friendlink.ApplicationListOptions{
		Status:       c.Query("status"),
		ApplyChannel: c.Query("channel"),
		Keyword:      c.Query("keyword"),
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		return err
	}
	respItems := make([]contract.FriendLinkApplicationResp, len(items))
	for i, item := range items {
		respItems[i] = contract.ToFriendLinkApplicationResp(item)
	}
	return response.Success(c, contract.FriendLinkApplicationListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// ApproveApplication godoc
// @Summary 审核通过友链申请
// @Tags FriendLinkAdmin
// @Produce json
// @Param id path int64 true "申请ID"
// @Success 200 {object} contract.FriendLinkApplicationResp
// @Security BearerAuth
// @Router /admin/friend-links/applications/{id}/approve [put]
func (h *FriendLinkAdminHandler) ApproveApplication(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的申请ID")
	}
	app, err := h.svc.ApproveApplication(c.Context(), id)
	if err != nil {
		return h.mapApplicationError(err)
	}
	return response.SuccessWithMessage(c, contract.ToFriendLinkApplicationResp(*app), "友链申请已通过")
}

// RejectApplication godoc
// @Summary 拒绝友链申请
// @Tags FriendLinkAdmin
// @Produce json
// @Param id path int64 true "申请ID"
// @Success 200 {object} contract.FriendLinkApplicationResp
// @Security BearerAuth
// @Router /admin/friend-links/applications/{id}/reject [put]
func (h *FriendLinkAdminHandler) RejectApplication(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的申请ID")
	}
	app, err := h.svc.RejectApplication(c.Context(), id)
	if err != nil {
		return h.mapApplicationError(err)
	}
	return response.SuccessWithMessage(c, contract.ToFriendLinkApplicationResp(*app), "友链申请已拒绝")
}

// BlockApplication godoc
// @Summary 封禁友链申请
// @Tags FriendLinkAdmin
// @Produce json
// @Param id path int64 true "申请ID"
// @Success 200 {object} contract.FriendLinkApplicationResp
// @Security BearerAuth
// @Router /admin/friend-links/applications/{id}/block [put]
func (h *FriendLinkAdminHandler) BlockApplication(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的申请ID")
	}
	app, err := h.svc.BlockApplication(c.Context(), id)
	if err != nil {
		return h.mapApplicationError(err)
	}
	return response.SuccessWithMessage(c, contract.ToFriendLinkApplicationResp(*app), "友链申请已封禁")
}

// UpdateApplicationStatus godoc
// @Summary 修改友链申请状态
// @Tags FriendLinkAdmin
// @Accept json
// @Produce json
// @Param id path int64 true "申请ID"
// @Param request body contract.FriendLinkApplicationStatusReq true "状态参数"
// @Success 200 {object} contract.FriendLinkApplicationResp
// @Security BearerAuth
// @Router /admin/friend-links/applications/{id}/status [put]
func (h *FriendLinkAdminHandler) UpdateApplicationStatus(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的申请ID")
	}
	var req contract.FriendLinkApplicationStatusReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Status) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "状态不能为空")
	}
	app, err := h.svc.UpdateApplicationStatus(c.Context(), id, req.Status)
	if err != nil {
		return h.mapApplicationError(err)
	}
	return response.SuccessWithMessage(c, contract.ToFriendLinkApplicationResp(*app), "友链申请状态已更新")
}

// ListFriendLinks godoc
// @Summary 获取友链列表（管理端）
// @Tags FriendLinkAdmin
// @Produce json
// @Param active query bool false "是否启用"
// @Param kind query string false "友链类型 manual/federation"
// @Param syncMode query string false "同步模式 none/rss/federation"
// @Param keyword query string false "关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} contract.FriendLinkListResp
// @Security BearerAuth
// @Router /admin/friend-links [get]
func (h *FriendLinkAdminHandler) ListFriendLinks(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 10)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var activePtr *bool
	if raw := strings.TrimSpace(c.Query("active")); raw != "" {
		if val, err := strconv.ParseBool(raw); err == nil {
			activePtr = &val
		}
	}

	items, total, err := h.svc.ListFriendLinks(c.Context(), friendlink.FriendLinkListOptions{
		IsActive: activePtr,
		Kind:     c.Query("kind"),
		SyncMode: c.Query("syncMode"),
		Keyword:  c.Query("keyword"),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return err
	}
	respItems := make([]contract.FriendLinkResp, len(items))
	for i, item := range items {
		respItems[i] = toFriendLinkResp(item)
	}
	return response.Success(c, contract.FriendLinkListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// CreateFriendLink godoc
// @Summary 创建友链
// @Tags FriendLinkAdmin
// @Accept json
// @Produce json
// @Param request body contract.FriendLinkCreateReq true "友链信息"
// @Success 200 {object} contract.FriendLinkResp
// @Security BearerAuth
// @Router /admin/friend-links [post]
func (h *FriendLinkAdminHandler) CreateFriendLink(c *fiber.Ctx) error {
	var req contract.FriendLinkCreateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	created, err := h.svc.CreateFriendLink(c.Context(), friendlink.CreateFriendLinkCmd{
		Name:         req.Name,
		URL:          req.URL,
		Logo:         req.Logo,
		Description:  req.Description,
		RSSURL:       req.RSSURL,
		Kind:         req.Kind,
		SyncMode:     req.SyncMode,
		InstanceID:   req.InstanceID,
		SyncInterval: req.SyncInterval,
		IsActive:     req.IsActive,
		UserID:       req.UserID,
	})
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}
	return response.SuccessWithMessage(c, toFriendLinkResp(*created), "友链创建成功")
}

// UpdateFriendLink godoc
// @Summary 更新友链
// @Tags FriendLinkAdmin
// @Accept json
// @Produce json
// @Param id path int64 true "友链ID"
// @Param request body contract.FriendLinkUpdateReq true "友链信息"
// @Success 200 {object} contract.FriendLinkResp
// @Security BearerAuth
// @Router /admin/friend-links/{id} [put]
func (h *FriendLinkAdminHandler) UpdateFriendLink(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的友链ID")
	}
	var req contract.FriendLinkUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	updated, err := h.svc.UpdateFriendLink(c.Context(), friendlink.UpdateFriendLinkCmd{
		ID:           id,
		Name:         req.Name,
		URL:          req.URL,
		Logo:         req.Logo,
		Description:  req.Description,
		RSSURL:       req.RSSURL,
		Kind:         req.Kind,
		SyncMode:     req.SyncMode,
		InstanceID:   req.InstanceID,
		SyncInterval: req.SyncInterval,
		IsActive:     req.IsActive,
		UserID:       req.UserID,
	})
	if err != nil {
		if errors.Is(err, social.ErrFriendLinkNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}
	return response.SuccessWithMessage(c, toFriendLinkResp(*updated), "友链更新成功")
}

// DeleteFriendLink godoc
// @Summary 删除友链
// @Tags FriendLinkAdmin
// @Produce json
// @Param id path int64 true "友链ID"
// @Success 200 {object} contract.EmptyRespEnvelope
// @Security BearerAuth
// @Router /admin/friend-links/{id} [delete]
func (h *FriendLinkAdminHandler) DeleteFriendLink(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的友链ID")
	}
	if err := h.svc.DeleteFriendLink(c.Context(), id); err != nil {
		return err
	}
	return response.SuccessWithMessage[any](c, nil, "友链已删除")
}

// BlockFriendLink godoc
// @Summary 封禁友链
// @Tags FriendLinkAdmin
// @Produce json
// @Param id path int64 true "友链ID"
// @Success 200 {object} contract.FriendLinkResp
// @Security BearerAuth
// @Router /admin/friend-links/{id}/block [put]
func (h *FriendLinkAdminHandler) BlockFriendLink(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的友链ID")
	}
	link, err := h.svc.BlockFriendLink(c.Context(), id)
	if err != nil {
		if errors.Is(err, social.ErrFriendLinkNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}
	return response.SuccessWithMessage(c, toFriendLinkResp(*link), "友链已封禁")
}

func (h *FriendLinkAdminHandler) mapApplicationError(err error) error {
	if errors.Is(err, social.ErrFriendLinkApplicationNotFound) {
		return response.NewBizError(response.NotFound)
	}
	if errors.Is(err, social.ErrFriendLinkApplicationBlocked) {
		return response.NewBizErrorWithMsg(response.Unauthorized, "已被封禁")
	}
	return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
}

func toFriendLinkResp(item social.FriendLink) contract.FriendLinkResp {
	return contract.FriendLinkResp{
		ID:               item.ID,
		Name:             item.Name,
		URL:              item.URL,
		Logo:             item.Logo,
		Description:      item.Description,
		RSSURL:           item.RSSURL,
		Kind:             item.Kind,
		SyncMode:         item.SyncMode,
		InstanceID:       item.InstanceID,
		LastSyncAt:       item.LastSyncAt,
		LastSyncStatus:   item.LastSyncStatus,
		SyncInterval:     item.SyncInterval,
		TotalPostsCached: item.TotalPostsCached,
		UserID:           item.UserID,
		IsActive:         item.IsActive,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}
