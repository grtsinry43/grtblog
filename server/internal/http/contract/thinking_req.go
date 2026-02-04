package contract

type CreateThinkingReq struct {
	Content      string `json:"content" validate:"required"`
	AllowComment *bool  `json:"allowComment,omitempty"`
}

type UpdateThinkingReq struct {
	Content      string `json:"content" validate:"required"`
	AllowComment *bool  `json:"allowComment,omitempty"`
}
