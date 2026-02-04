package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	appEvent "github.com/grtsinry43/grtblog-v2/server/internal/app/event"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type EventHandler struct{}

func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

// ListEvents godoc
// @Summary 获取事件分组列表
// @Tags Event
// @Produce json
// @Param channel query string false "通道过滤: email|webhook"
// @Success 200 {object} contract.AdminEventListResp
// @Security BearerAuth
// @Router /admin/events [get]
// @Security JWTAuth
func (h *EventHandler) ListEvents(c *fiber.Ctx) error {
	channel := normalizeChannel(c.Query("channel"))
	groups := appEvent.GroupsByChannel(channel)
	respGroups := make([]contract.AdminEventGroupResp, len(groups))
	for i, group := range groups {
		respGroups[i] = contract.AdminEventGroupResp{
			Category: group.Category,
			Events:   group.Events,
		}
	}
	return response.Success(c, contract.AdminEventListResp{Groups: respGroups})
}

// ListEventCatalog godoc
// @Summary 获取事件目录（含参数）
// @Tags Event
// @Produce json
// @Param channel query string false "通道过滤: email|webhook"
// @Success 200 {object} contract.AdminEventCatalogResp
// @Security BearerAuth
// @Router /admin/events/catalog [get]
// @Security JWTAuth
func (h *EventHandler) ListEventCatalog(c *fiber.Ctx) error {
	channel := normalizeChannel(c.Query("channel"))
	items := appEvent.Catalog()
	respItems := make([]contract.AdminEventDescriptorResp, 0, len(items))
	for _, item := range items {
		if channel != "" && !containsChannel(item.Channels, channel) {
			continue
		}
		respItems = append(respItems, mapEventDescriptor(item))
	}
	return response.Success(c, contract.AdminEventCatalogResp{Items: respItems})
}

// GetEventCatalogItem godoc
// @Summary 获取单个事件目录详情
// @Tags Event
// @Produce json
// @Param name path string true "事件名"
// @Success 200 {object} contract.AdminEventDescriptorResp
// @Security BearerAuth
// @Router /admin/events/catalog/{name} [get]
// @Security JWTAuth
func (h *EventHandler) GetEventCatalogItem(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.Params("name"))
	if name == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "事件名不能为空")
	}
	item, ok := appEvent.CatalogByName(name)
	if !ok {
		return response.NewBizErrorWithMsg(response.NotFound, "事件不存在")
	}
	return response.Success(c, mapEventDescriptor(item))
}

func normalizeChannel(channel string) string {
	switch strings.TrimSpace(strings.ToLower(channel)) {
	case appEvent.ChannelEmail:
		return appEvent.ChannelEmail
	case appEvent.ChannelWebhook:
		return appEvent.ChannelWebhook
	default:
		return ""
	}
}

func containsChannel(channels []string, target string) bool {
	for _, item := range channels {
		if item == target {
			return true
		}
	}
	return false
}

func mapEventDescriptor(item appEvent.EventDescriptor) contract.AdminEventDescriptorResp {
	allFields := append([]appEvent.EventField(nil), item.Fields...)
	allFields = append(allFields, appEvent.GlobalTemplateFields()...)
	fields := make([]contract.AdminEventFieldResp, len(allFields))
	for i, field := range allFields {
		fields[i] = contract.AdminEventFieldResp{
			Name:        field.Name,
			Type:        field.Type,
			Required:    field.Required,
			Description: field.Description,
		}
	}
	channels := append([]string(nil), item.Channels...)
	return contract.AdminEventDescriptorResp{
		Name:        item.Name,
		Title:       item.Title,
		Category:    item.Category,
		Description: item.Description,
		PublicEmail: item.PublicEmail,
		Channels:    channels,
		Fields:      fields,
	}
}
