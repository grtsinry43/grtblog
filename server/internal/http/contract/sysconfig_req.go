package contract

import "time"

// SysConfigExportResp 配置导出响应。
type SysConfigExportResp struct {
	Version    int                   `json:"version"`
	ExportedAt time.Time             `json:"exportedAt"`
	Configs    []SysConfigExportItem `json:"configs"`
}

// SysConfigExportItem 导出的单条配置。
type SysConfigExportItem struct {
	Key   string `json:"key"`
	Value any    `json:"value,omitempty"`
}

// SysConfigImportReq 配置导入请求。
type SysConfigImportReq struct {
	Version int                   `json:"version"`
	Configs []SysConfigExportItem `json:"configs" validate:"required"`
}

// SysConfigBatchUpdateReq 系统配置批量更新请求。
type SysConfigBatchUpdateReq struct {
	Items []SysConfigUpdateItem `json:"items" validate:"required"`
}

// SysConfigUpdateItem 单条配置更新。
type SysConfigUpdateItem struct {
	Key          string   `json:"key" validate:"required,max=45"`
	Value        *JSONRaw `json:"value,omitempty" swaggertype:"object"`
	IsSensitive  *bool    `json:"isSensitive,omitempty"`
	GroupPath    *string  `json:"groupPath,omitempty"`
	Label        *string  `json:"label,omitempty"`
	Description  *string  `json:"description,omitempty"`
	ValueType    *string  `json:"valueType,omitempty"`
	EnumOptions  *JSONRaw `json:"enumOptions,omitempty" swaggertype:"object"`
	DefaultValue *JSONRaw `json:"defaultValue,omitempty" swaggertype:"object"`
	VisibleWhen  *JSONRaw `json:"visibleWhen,omitempty" swaggertype:"object"`
	Sort         *int     `json:"sort,omitempty"`
	Meta         *JSONRaw `json:"meta,omitempty" swaggertype:"object"`
}
