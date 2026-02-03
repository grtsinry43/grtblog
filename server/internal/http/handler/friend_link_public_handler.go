package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type FriendLinkPublicHandler struct {
	svc *friendlink.LinkService
}

func NewFriendLinkPublicHandler(svc *friendlink.LinkService) *FriendLinkPublicHandler {
	return &FriendLinkPublicHandler{svc: svc}
}

// ListPublic godoc
// @Summary 公开获取友链列表
// @Tags FriendLink
// @Produce json
// @Param kind query string false "友链类型 manual/federation"
// @Param syncMode query string false "同步模式 none/rss/federation"
// @Param keyword query string false "关键词"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(50)
// @Success 200 {object} contract.FriendLinkPublicListResp
// @Router /public/friend-links [get]
func (h *FriendLinkPublicHandler) ListPublic(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 50)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	active := true
	items, total, err := h.svc.List(c.Context(), friendlink.FriendLinkListOptions{
		IsActive: &active,
		Kind:     strings.TrimSpace(c.Query("kind")),
		SyncMode: strings.TrimSpace(c.Query("syncMode")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return err
	}
	respItems := make([]contract.FriendLinkPublicResp, len(items))
	for i, item := range items {
		respItems[i] = contract.FriendLinkPublicResp{
			Name:        item.Name,
			URL:         item.URL,
			Logo:        item.Logo,
			Description: item.Description,
			RSSURL:      item.RSSURL,
			Kind:        item.Kind,
			SyncMode:    item.SyncMode,
		}
	}
	return response.Success(c, contract.FriendLinkPublicListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}
