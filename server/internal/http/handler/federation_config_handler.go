package handler

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

// FederationConfigHandler provides settings-center style APIs for federation_config.
type FederationConfigHandler struct {
	svc *sysconfig.Service
}

func NewFederationConfigHandler(svc *sysconfig.Service) *FederationConfigHandler {
	return &FederationConfigHandler{svc: svc}
}

// ListFederationConfig lists federation config items.
// @Summary 联合配置列表
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param keys query string false "指定配置 key（逗号分隔）"
// @Success 200 {object} contract.SysConfigTreeResp
// @Security BearerAuth
// @Router /admin/federation/config [get]
// @Security JWTAuth
func (h *FederationConfigHandler) ListFederationConfig(c *fiber.Ctx) error {
	keys, err := parseAndValidateFederationConfigKeys(c.Query("keys"), "federation.")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, err.Error())
	}

	items, err := h.svc.ListConfigs(c.Context(), keys)
	if err != nil {
		return err
	}
	items = filterFederationConfigsByPrefix(items, "federation.")
	tree, err := buildSysConfigTree(items)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "配置解析失败", err)
	}
	return response.Success(c, tree)
}

// UpdateFederationConfig updates federation config items.
// @Summary 更新联合配置
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param request body contract.SysConfigBatchUpdateReq true "配置更新参数"
// @Success 200 {object} contract.SysConfigTreeResp
// @Security BearerAuth
// @Router /admin/federation/config [put]
// @Security JWTAuth
func (h *FederationConfigHandler) UpdateFederationConfig(c *fiber.Ctx) error {
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
		if !strings.HasPrefix(key, "federation.") {
			return response.NewBizErrorWithMsg(response.ParamsError, "仅允许更新 federation.* 配置")
		}
		if key == "federation.instanceURL" && item.Value != nil {
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
	updated = filterFederationConfigsByPrefix(updated, "federation.")
	tree, err := buildSysConfigTree(updated)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "配置解析失败", err)
	}
	return response.SuccessWithMessage(c, tree, "更新成功")
}

// ExportFederationConfigs exports all federation.* and activitypub.* configs including sensitive values.
// @Summary 导出联合配置
// @Tags FederationAdmin
// @Produce json
// @Success 200 {object} contract.SysConfigExportResp
// @Security BearerAuth
// @Router /admin/federation/export [get]
// @Security JWTAuth
func (h *FederationConfigHandler) ExportFederationConfigs(c *fiber.Ctx) error {
	items, err := h.svc.ListConfigs(c.Context(), nil)
	if err != nil {
		return err
	}

	var fedItems []domainconfig.SysConfig
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if strings.HasPrefix(key, "federation.") || strings.HasPrefix(key, "activitypub.") {
			fedItems = append(fedItems, item)
		}
	}

	configs := make([]contract.SysConfigExportItem, 0, len(fedItems))
	for _, item := range fedItems {
		valueType, err := normalizeValueType(item.ValueType)
		if err != nil {
			return response.NewBizErrorWithCause(response.ServerError, "配置解析失败", err)
		}
		raw, err := valueToJSON(valueType, item.Value)
		if err != nil {
			return response.NewBizErrorWithCause(response.ServerError, "配置值序列化失败", err)
		}
		var value any
		if raw != nil {
			_ = json.Unmarshal(*raw, &value)
		}
		configs = append(configs, contract.SysConfigExportItem{
			Key:   item.Key,
			Value: value,
		})
	}

	return response.Success(c, contract.SysConfigExportResp{
		Version:    1,
		ExportedAt: time.Now().UTC(),
		Configs:    configs,
	})
}

// ImportFederationConfigs imports federation.* and activitypub.* configs from an export payload.
// @Summary 导入联合配置
// @Tags FederationAdmin
// @Accept json
// @Produce json
// @Param request body contract.SysConfigImportReq true "导入数据"
// @Success 200
// @Security BearerAuth
// @Router /admin/federation/import [post]
// @Security JWTAuth
func (h *FederationConfigHandler) ImportFederationConfigs(c *fiber.Ctx) error {
	var req contract.SysConfigImportReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if len(req.Configs) == 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "configs 不能为空")
	}

	updates := make([]sysconfig.UpdateItem, 0, len(req.Configs))
	for _, item := range req.Configs {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return response.NewBizErrorWithMsg(response.ParamsError, "key 不能为空")
		}
		if !strings.HasPrefix(key, "federation.") && !strings.HasPrefix(key, "activitypub.") {
			return response.NewBizErrorWithMsg(response.ParamsError, "仅允许导入 federation.* 或 activitypub.* 配置")
		}

		// Validate instanceURL format
		if key == "federation.instanceURL" || key == "activitypub.instanceURL" {
			if s, ok := item.Value.(string); ok {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" && !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
					return response.NewBizErrorWithMsg(response.ParamsError, key+" 必须以 http:// 或 https:// 开头")
				}
			}
		}

		encoded, err := json.Marshal(item.Value)
		if err != nil {
			return response.NewBizErrorWithMsg(response.ParamsError, "配置值序列化失败: "+key)
		}
		raw := json.RawMessage(encoded)
		updates = append(updates, sysconfig.UpdateItem{
			Key:   key,
			Value: &raw,
		})
	}

	if _, err := h.svc.UpdateConfigs(c.Context(), updates); err != nil {
		var validationErr *sysconfig.UpdateValidationError
		if errors.As(err, &validationErr) {
			return response.NewBizErrorWithMsg(response.ParamsError, validationErr.Error())
		}
		return err
	}

	return response.SuccessWithMessage(c, "导入成功", "导入成功")
}

func parseAndValidateFederationConfigKeys(raw string, prefix string) ([]string, error) {
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

func filterFederationConfigsByPrefix(items []domainconfig.SysConfig, prefix string) []domainconfig.SysConfig {
	out := make([]domainconfig.SysConfig, 0, len(items))
	for _, item := range items {
		if strings.HasPrefix(strings.TrimSpace(item.Key), prefix) {
			out = append(out, item)
		}
	}
	return out
}
