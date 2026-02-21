package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/activitypubconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type ActivityPubConfigHandler struct {
	svc *activitypubconfig.Service
}

func NewActivityPubConfigHandler(svc *activitypubconfig.Service) *ActivityPubConfigHandler {
	return &ActivityPubConfigHandler{svc: svc}
}

// ListActivityPubConfig lists activitypub config items.
// @Summary ActivityPub 配置列表
// @Tags ActivityPubAdmin
// @Accept json
// @Produce json
// @Param keys query string false "指定配置 key（逗号分隔）"
// @Success 200 {object} contract.SysConfigTreeResp
// @Security BearerAuth
// @Router /admin/activitypub/config [get]
// @Security JWTAuth
func (h *ActivityPubConfigHandler) ListActivityPubConfig(c *fiber.Ctx) error {
	keys, err := parseAndValidateActivityPubConfigKeys(c.Query("keys"), "activitypub.")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}
	items, err := h.svc.ListConfigs(c.Context(), keys)
	if err != nil {
		return err
	}
	items = filterActivityPubConfigsByPrefix(items, "activitypub.")
	tree, err := buildSysConfigTree(items)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "配置解析失败", err)
	}
	return response.Success(c, tree)
}

// UpdateActivityPubConfig updates activitypub config items.
// @Summary 更新 ActivityPub 配置
// @Tags ActivityPubAdmin
// @Accept json
// @Produce json
// @Param request body contract.SysConfigBatchUpdateReq true "配置更新参数"
// @Success 200 {object} contract.SysConfigTreeResp
// @Security BearerAuth
// @Router /admin/activitypub/config [put]
// @Security JWTAuth
func (h *ActivityPubConfigHandler) UpdateActivityPubConfig(c *fiber.Ctx) error {
	var req contract.SysConfigBatchUpdateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.Items) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "items 不能为空")
	}

	updates := make([]sysconfig.UpdateItem, 0, len(req.Items))
	for _, item := range req.Items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return response.NewBizErrorWithMsg(response.ParamsError, "key 不能为空")
		}
		if !strings.HasPrefix(key, "activitypub.") {
			return response.NewBizErrorWithMsg(response.ParamsError, "仅允许更新 activitypub.* 配置")
		}
		if key == "activitypub.instanceURL" && item.Value != nil {
			var instanceURL string
			if err := json.Unmarshal(json.RawMessage(*item.Value), &instanceURL); err != nil {
				return response.NewBizErrorWithMsg(response.ParamsError, "instanceURL 必须为字符串")
			}
			trimmed := strings.TrimSpace(instanceURL)
			if trimmed != "" && !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
				return response.NewBizErrorWithMsg(response.ParamsError, "instanceURL 必须以 http:// 或 https:// 开头")
			}
		}
		updates = append(updates, sysconfig.UpdateItem{
			Key:          key,
			Value:        contract.RawMessagePtr(item.Value),
			IsSensitive:  item.IsSensitive,
			GroupPath:    item.GroupPath,
			Label:        item.Label,
			Description:  item.Description,
			ValueType:    item.ValueType,
			EnumOptions:  contract.RawMessagePtr(item.EnumOptions),
			DefaultValue: contract.RawMessagePtr(item.DefaultValue),
			VisibleWhen:  contract.RawMessagePtr(item.VisibleWhen),
			Sort:         item.Sort,
			Meta:         contract.RawMessagePtr(item.Meta),
		})
	}

	updated, err := h.svc.UpdateConfigs(c.Context(), updates)
	if err != nil {
		var validationErr *sysconfig.UpdateValidationError
		if errors.As(err, &validationErr) {
			return response.NewBizErrorWithMsg(response.ParamsError, validationErr.Error())
		}
		return err
	}
	updated = filterActivityPubConfigsByPrefix(updated, "activitypub.")
	tree, err := buildSysConfigTree(updated)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "配置解析失败", err)
	}
	return response.SuccessWithMessage(c, tree, "更新成功")
}

func parseAndValidateActivityPubConfigKeys(raw string, prefix string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	keys := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, item := range parts {
		key := strings.TrimSpace(item)
		if key == "" {
			continue
		}
		if !strings.HasPrefix(key, prefix) {
			return nil, errors.New("仅允许查询 " + prefix + " 配置")
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys, nil
}

func filterActivityPubConfigsByPrefix(items []domainconfig.SysConfig, prefix string) []domainconfig.SysConfig {
	out := make([]domainconfig.SysConfig, 0, len(items))
	for _, item := range items {
		if strings.HasPrefix(strings.TrimSpace(item.Key), prefix) {
			out = append(out, item)
		}
	}
	return out
}
