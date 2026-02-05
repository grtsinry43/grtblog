package handler

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/friendtimeline"
	domainfed "github.com/grtsinry43/grtblog-v2/server/internal/domain/federation"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type FriendTimelineHandler struct {
	svc *friendtimeline.Service
}

func NewFriendTimelineHandler(svc *friendtimeline.Service) *FriendTimelineHandler {
	return &FriendTimelineHandler{svc: svc}
}

// ListPublic 返回聚合后的朋友时间线（RSS + 联邦）。
// @Summary 朋友聚合时间线
// @Tags FriendTimeline
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} contract.FriendTimelineListResp
// @Router /public/friend-timeline [get]
func (h *FriendTimelineHandler) ListPublic(c *fiber.Ctx) error {
	if h.svc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "时间线服务未初始化")
	}
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 20)
	result, err := h.svc.List(c.Context(), page, pageSize)
	if err != nil {
		return err
	}
	items := make([]contract.FederationPostResp, len(result.Items))
	for i := range result.Items {
		items[i] = mapCachePostToFederationPost(result.Items[i])
	}
	return response.Success(c, contract.FriendTimelineListResp{
		Items: items,
		Total: result.Total,
		Page:  result.Page,
		Size:  result.Size,
	})
}

func mapCachePostToFederationPost(item domainfed.FederatedPostCache) contract.FederationPostResp {
	author := contract.FederationPostAuthorResp{Name: ""}
	var payload struct {
		Name   string  `json:"name"`
		URL    *string `json:"url,omitempty"`
		Avatar *string `json:"avatar,omitempty"`
	}
	if err := json.Unmarshal(item.Author, &payload); err == nil {
		author.Name = payload.Name
		author.URL = payload.URL
		author.Avatar = payload.Avatar
	}
	id := strings.TrimSpace(item.URL)
	if item.RemotePostID != nil && strings.TrimSpace(*item.RemotePostID) != "" {
		id = strings.TrimSpace(*item.RemotePostID)
	}
	return contract.FederationPostResp{
		ID:             id,
		URL:            item.URL,
		Title:          item.Title,
		Summary:        item.Summary,
		ContentPreview: item.ContentPreview,
		Author:         author,
		PublishedAt:    item.PublishedAt,
		UpdatedAt:      item.UpdatedAt,
		CoverImage:     item.CoverImage,
		Language:       item.Language,
		AllowCitation:  item.AllowCitation,
		AllowComment:   item.AllowComment,
	}
}
