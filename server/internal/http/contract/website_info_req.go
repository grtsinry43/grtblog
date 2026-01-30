package contract

// WebsiteInfoReq 网站信息请求。
type WebsiteInfoReq struct {
	Key      string   `json:"key"`
	Name     *string  `json:"name"`
	Value    *string  `json:"value"`
	InfoJSON *JSONRaw `json:"infoJson" swaggertype:"object"`
}
