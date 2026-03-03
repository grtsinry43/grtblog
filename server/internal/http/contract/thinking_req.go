package contract

import (
	"encoding/json"
	"strings"
	"time"
)

type CreateThinkingReq struct {
	Content      string     `json:"content" validate:"required"`
	AllowComment *bool      `json:"allowComment,omitempty"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"`
}

type createThinkingReqJSON struct {
	Content      string  `json:"content"`
	AllowComment *bool   `json:"allowComment"`
	CreatedAt    *string `json:"createdAt"`
}

func (r *CreateThinkingReq) UnmarshalJSON(data []byte) error {
	var aux createThinkingReqJSON
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Content = aux.Content
	r.AllowComment = aux.AllowComment

	if aux.CreatedAt == nil {
		r.CreatedAt = nil
		return nil
	}
	if strings.TrimSpace(*aux.CreatedAt) == "" {
		now := time.Now()
		r.CreatedAt = &now
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *aux.CreatedAt)
	if err != nil {
		return err
	}
	r.CreatedAt = &parsed
	return nil
}

type UpdateThinkingReq struct {
	Content      string `json:"content" validate:"required"`
	AllowComment *bool  `json:"allowComment,omitempty"`
}

// BatchDeleteThinkingReq 批量删除思考请求。
type BatchDeleteThinkingReq struct {
	IDs []int64 `json:"ids"`
}
