package event

import (
	"context"
	"encoding/json"
	"strings"
)

// WebsiteInfoProvider abstracts the subset of sysconfig.Service used for
// building global template variables. Defining this interface here avoids
// an import cycle between event and sysconfig.
type WebsiteInfoProvider interface {
	WebsiteInfo(ctx context.Context) (map[string]string, error)
}

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

func BuildGlobalTemplateVariables(ctx context.Context, provider WebsiteInfoProvider) map[string]any {
	vars := map[string]any{
		"api_url":           "",
		"description":       "",
		"keywords":          "",
		"favicon":           "",
		"theme_extend_info": map[string]any{},
		"website_name":      "",
		"public_url":        "",
	}
	if provider == nil {
		return vars
	}
	info, err := provider.WebsiteInfo(ctx)
	if err != nil {
		return vars
	}
	for key, value := range info {
		switch key {
		case "api_url", "description", "keywords", "favicon", "website_name", "public_url":
			vars[key] = value
		case "theme_extend_info":
			parsed := parseThemeExtendInfoString(value)
			vars[key] = parsed
		}
	}
	return vars
}

func parseThemeExtendInfoString(val string) map[string]any {
	result := map[string]any{}
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return result
	}
	if err := json.Unmarshal([]byte(trimmed), &result); err == nil {
		return result
	}
	return result
}
