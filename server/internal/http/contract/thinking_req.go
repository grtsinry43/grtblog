package contract

type CreateThinkingReq struct {
	Content      string `json:"content" validate:"required"`
	AllowComment *bool  `json:"allowComment,omitempty"`
}

type UpdateThinkingReq struct {
	Content      string `json:"content" validate:"required"`
	AllowComment *bool  `json:"allowComment,omitempty"`
}

// BatchDeleteThinkingReq 批量删除思考请求。
type BatchDeleteThinkingReq struct {
	IDs []int64 `json:"ids"`
}
