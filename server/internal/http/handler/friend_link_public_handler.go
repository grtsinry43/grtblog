package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendlink"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type FriendLinkPublicHandler struct {
	svc    *friendlink.LinkService
	sysCfg *sysconfig.Service
}

func NewFriendLinkPublicHandler(svc *friendlink.LinkService, sysCfg *sysconfig.Service) *FriendLinkPublicHandler {
	return &FriendLinkPublicHandler{svc: svc, sysCfg: sysCfg}
}

// ListPublic godoc
// @Summary 公开获取友链列表
// @Tags FriendLink
// @Produce json
// @Param type query string false "友链类型 federation/rss/norss"
// @Param keyword query string false "关键词"
// @Success 200 {object} []contract.FriendLinkPublicResp
// @Router /public/friend-links [get]
func (h *FriendLinkPublicHandler) ListPublic(c *fiber.Ctx) error {
	active := true
	items, _, err := h.svc.List(c.Context(), friendlink.FriendLinkListOptions{
		IsActive: &active,
		Type:     strings.TrimSpace(c.Query("type")),
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Page:     1,
		PageSize: 0,
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
			Type:        item.Type,
		}
	}
	return response.Success(c, respItems)
}

// GetApplyConfig godoc
// @Summary 获取友链申请配置
// @Tags FriendLink
// @Produce json
// @Success 200 {object} contract.FriendLinkApplyConfigResp
// @Router /public/friend-links/apply-config [get]
func (h *FriendLinkPublicHandler) GetApplyConfig(c *fiber.Ctx) error {
	resp := contract.FriendLinkApplyConfigResp{
		Enabled:      true,
		Requirements: "",
	}
	if h.sysCfg != nil {
		if v, err := h.sysCfg.GetConfigValue(c.Context(), "friendlink.apply.enabled"); err == nil && v == "false" {
			resp.Enabled = false
		}
		if v, err := h.sysCfg.GetConfigValue(c.Context(), "friendlink.apply.requirements"); err == nil {
			resp.Requirements = v
		}
	}
	return response.Success(c, resp)
}
