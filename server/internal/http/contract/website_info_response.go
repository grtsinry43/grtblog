package contract

import (
	"encoding/json"

	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type WebsiteInfoResp struct {
	Key      string           `json:"key"`
	Name     *string          `json:"name,omitempty"`
	Value    *string          `json:"value,omitempty"`
	InfoJSON *json.RawMessage `json:"infoJson,omitempty" swaggertype:"object"`
}

// 用于 swagger 展示。
type WebsiteInfoListRespEnvelope struct {
	Code   int               `json:"code"`
	BizErr string            `json:"bizErr"`
	Msg    string            `json:"msg"`
	Data   []WebsiteInfoResp `json:"data"`
	Meta   response.Meta     `json:"meta"`
}

type WebsiteInfoDetailRespEnvelope struct {
	Code   int             `json:"code"`
	BizErr string          `json:"bizErr"`
	Msg    string          `json:"msg"`
	Data   WebsiteInfoResp `json:"data"`
	Meta   response.Meta   `json:"meta"`
}

type GenericMessageEnvelope struct {
	Code   int           `json:"code"`
	BizErr string        `json:"bizErr"`
	Msg    string        `json:"msg"`
	Data   interface{}   `json:"data"`
	Meta   response.Meta `json:"meta"`
}
