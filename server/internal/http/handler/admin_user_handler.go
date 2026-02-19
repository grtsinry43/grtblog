package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminuser"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AdminUserHandler struct {
	svc *adminuser.Service
}

func NewAdminUserHandler(svc *adminuser.Service) *AdminUserHandler {
	return &AdminUserHandler{svc: svc}
}

// ListUsers godoc
// @Summary 获取本站用户列表（管理端）
// @Tags AdminUser
// @Produce json
// @Param keyword query string false "关键词（用户名/昵称/邮箱）"
// @Param admin query bool false "是否管理员"
// @Param active query bool false "是否启用"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.AdminUserListResp
// @Security JWTAuth
// @Router /admin/users [get]
func (h *AdminUserHandler) ListUsers(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 20)
	keyword := strings.TrimSpace(c.Query("keyword"))

	var onlyAdmin *bool
	if raw := strings.TrimSpace(c.Query("admin")); raw != "" {
		val, err := strconv.ParseBool(raw)
		if err == nil {
			onlyAdmin = &val
		}
	}
	var onlyActive *bool
	if raw := strings.TrimSpace(c.Query("active")); raw != "" {
		val, err := strconv.ParseBool(raw)
		if err == nil {
			onlyActive = &val
		}
	}

	users, total, err := h.svc.ListUsers(c.Context(), adminuser.ListUsersCmd{
		Keyword:    keyword,
		OnlyAdmin:  onlyAdmin,
		OnlyActive: onlyActive,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return err
	}

	items := make([]contract.UserResp, len(users))
	for i := range users {
		users[i].Password = ""
		items[i] = contract.ToUserResp(users[i])
	}
	return response.Success(c, contract.AdminUserListResp{
		Items: items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// UpdateUser godoc
// @Summary 更新本站用户（管理端）
// @Tags AdminUser
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body contract.UpdateAdminUserReq true "更新用户参数"
// @Success 200 {object} contract.UserResp
// @Security JWTAuth
// @Router /admin/users/{id} [put]
func (h *AdminUserHandler) UpdateUser(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}
	userID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || userID <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的用户ID")
	}

	var req contract.UpdateAdminUserReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	updated, err := h.svc.UpdateUser(c.Context(), adminuser.UpdateUserCmd{
		OperatorID: claims.UserID,
		UserID:     userID,
		Nickname:   req.Nickname,
		Email:      req.Email,
		IsActive:   req.IsActive,
		IsAdmin:    req.IsAdmin,
	})
	if err != nil {
		return h.mapErr(err)
	}
	return response.SuccessWithMessage(c, contract.ToUserResp(*updated), "用户信息已更新")
}

func (h *AdminUserHandler) mapErr(err error) error {
	switch {
	case errors.Is(err, identity.ErrUserNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "用户不存在")
	case errors.Is(err, identity.ErrUserExists):
		return response.NewBizErrorWithMsg(response.ParamsError, "用户名或邮箱已存在")
	case errors.Is(err, adminuser.ErrLastAdminMutation):
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	case errors.Is(err, adminuser.ErrSelfAdminMutation):
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	default:
		return err
	}
}
