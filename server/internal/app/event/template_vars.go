package event

import (
	"context"
	"encoding/json"
	"strings"

	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
)

var globalTemplateFields = []EventField{
	{Name: "api_url", Type: "string", Required: false, Description: "站点 API 地址"},
	{Name: "description", Type: "string", Required: false, Description: "站点描述"},
	{Name: "keywords", Type: "string", Required: false, Description: "站点关键词"},
	{Name: "favicon", Type: "string", Required: false, Description: "站点 favicon"},
	{Name: "theme_extend_info", Type: "object", Required: false, Description: "主题扩展信息(JSON对象)"},
	{Name: "website_name", Type: "string", Required: false, Description: "网站名称"},
	{Name: "public_url", Type: "string", Required: false, Description: "站点公开地址"},
}

func GlobalTemplateFields() []EventField {
	return append([]EventField(nil), globalTemplateFields...)
}

func BuildGlobalTemplateVariables(ctx context.Context, repo domainconfig.WebsiteInfoRepository) map[string]any {
	vars := map[string]any{
		"api_url":           "",
		"description":       "",
		"keywords":          "",
		"favicon":           "",
		"theme_extend_info": map[string]any{},
		"website_name":      "",
		"public_url":        "",
	}
	if repo == nil {
		return vars
	}
	items, err := repo.List(ctx)
	if err != nil {
		return vars
	}
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		switch key {
		case "api_url", "description", "keywords", "favicon", "website_name", "public_url":
			if item.Value != nil {
				vars[key] = *item.Value
			}
		case "theme_extend_info":
			parsed := parseThemeExtendInfo(item)
			vars[key] = parsed
		}
	}
	return vars
}

func parseThemeExtendInfo(item domainconfig.WebsiteInfo) map[string]any {
	result := map[string]any{}
	if len(item.InfoJSON) > 0 {
		if err := json.Unmarshal(item.InfoJSON, &result); err == nil {
			return result
		}
	}
	if item.Value != nil {
		trimmed := strings.TrimSpace(*item.Value)
		if trimmed != "" {
			if err := json.Unmarshal([]byte(trimmed), &result); err == nil {
				return result
			}
		}
	}
	return result
}
