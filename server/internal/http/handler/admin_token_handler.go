package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/middleware"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence"
)

type AdminTokenHandler struct {
	tokenRepo *persistence.AdminTokenRepository
	userRepo  *persistence.IdentityRepository
}

func NewAdminTokenHandler(tokenRepo *persistence.AdminTokenRepository, userRepo *persistence.IdentityRepository) *AdminTokenHandler {
	return &AdminTokenHandler{tokenRepo: tokenRepo, userRepo: userRepo}
}

type AdminTokenListItem struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"userId"`
	Username     string `json:"username"`
	Description  string `json:"description"`
	TokenPreview string `json:"tokenPreview"`
	ExpireAt     string `json:"expireAt"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
	IsExpired    bool   `json:"isExpired"`
}

type AdminTokenListResp struct {
	Items []AdminTokenListItem `json:"items"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

type CreateAdminTokenReq struct {
	Description string `json:"description"`
	ExpireAt    string `json:"expireAt"`
}

type CreateAdminTokenResp struct {
	AdminTokenListItem
	Token string `json:"token"`
}

// List godoc
// @Summary 管理员令牌列表
// @Tags Admin-Token
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} AdminTokenListResp
// @Security BearerAuth
// @Router /admin/tokens [get]
func (h *AdminTokenHandler) List(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	size := parseIntQuery(c, "pageSize", 10)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	items, total, err := h.tokenRepo.List(c.Context(), page, size)
	if err != nil {
		return err
	}

	respItems := make([]AdminTokenListItem, 0, len(items))
	for _, item := range items {
		username := strconv.FormatInt(item.UserID, 10)
		if user, uErr := h.userRepo.FindByID(c.Context(), item.UserID); uErr == nil && user != nil {
			if strings.TrimSpace(user.Nickname) != "" {
				username = user.Nickname
			} else if strings.TrimSpace(user.Username) != "" {
				username = user.Username
			}
		}
		respItems = append(respItems, toAdminTokenListItem(item, username))
	}

	return response.Success(c, AdminTokenListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create godoc
// @Summary 创建管理员令牌
// @Tags Admin-Token
// @Accept json
// @Produce json
// @Param request body CreateAdminTokenReq true "创建参数"
// @Success 200 {object} CreateAdminTokenResp
// @Security BearerAuth
// @Router /admin/tokens [post]
func (h *AdminTokenHandler) Create(c *fiber.Ctx) error {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		return response.ErrorFromBiz[any](c, response.NotLogin)
	}

	var req CreateAdminTokenReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	desc := strings.TrimSpace(req.Description)
	expireAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ExpireAt))
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "expireAt 必须是 RFC3339 时间")
	}
	expireAt = expireAt.UTC()
	if !expireAt.After(time.Now().UTC()) {
		return response.NewBizErrorWithMsg(response.ParamsError, "expireAt 必须晚于当前时间")
	}

	tokenEntity := identity.AdminToken{
		UserID:      claims.UserID,
		Description: desc,
		ExpireAt:    expireAt,
	}

	rawToken, err := h.createTokenWithRetry(c.Context(), &tokenEntity)
	if err != nil {
		return err
	}

	username := strconv.FormatInt(tokenEntity.UserID, 10)
	if user, uErr := h.userRepo.FindByID(c.Context(), tokenEntity.UserID); uErr == nil && user != nil {
		if strings.TrimSpace(user.Nickname) != "" {
			username = user.Nickname
		} else if strings.TrimSpace(user.Username) != "" {
			username = user.Username
		}
	}

	Audit(c, "admin.token.create", map[string]any{"id": tokenEntity.ID})
	return response.Success(c, CreateAdminTokenResp{
		AdminTokenListItem: toAdminTokenListItem(tokenEntity, username),
		Token:              rawToken,
	})
}

// Delete godoc
// @Summary 删除管理员令牌
// @Tags Admin-Token
// @Produce json
// @Param id path int true "令牌ID"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /admin/tokens/{id} [delete]
func (h *AdminTokenHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的令牌 ID")
	}
	if err := h.tokenRepo.DeleteByID(c.Context(), id); err != nil {
		if errors.Is(err, identity.ErrAdminTokenNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return err
	}
	Audit(c, "admin.token.delete", map[string]any{"id": id})
	return response.SuccessWithMessage[any](c, nil, "deleted")
}

func (h *AdminTokenHandler) createTokenWithRetry(ctx context.Context, tokenEntity *identity.AdminToken) (string, error) {
	const maxRetry = 5
	for i := 0; i < maxRetry; i++ {
		rawToken, err := newGTToken()
		if err != nil {
			return "", response.NewBizErrorWithCause(response.ServerError, "生成令牌失败", err)
		}
		tokenEntity.Token = persistence.HashAdminToken(rawToken)
		err = h.tokenRepo.Create(ctx, tokenEntity)
		if err == nil {
			return rawToken, nil
		}
		if !h.tokenRepo.IsDuplicateTokenError(err) {
			return "", err
		}
	}
	return "", response.NewBizErrorWithMsg(response.ServerError, "生成令牌失败，请重试")
}

func toAdminTokenListItem(item identity.AdminToken, username string) AdminTokenListItem {
	return AdminTokenListItem{
		ID:           item.ID,
		UserID:       item.UserID,
		Username:     username,
		Description:  item.Description,
		TokenPreview: maskToken(item.Token),
		ExpireAt:     item.ExpireAt.UTC().Format(time.RFC3339),
		CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    item.UpdatedAt.UTC().Format(time.RFC3339),
		IsExpired:    time.Now().UTC().After(item.ExpireAt),
	}
}

func maskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if len(token) <= 10 {
		return token
	}
	return token[:7] + "..." + token[len(token)-4:]
}

func newGTToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "gt_" + hex.EncodeToString(buf), nil
}
