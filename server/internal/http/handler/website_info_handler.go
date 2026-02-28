package handler

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type WebsiteInfoHandler struct {
	sysCfg *sysconfig.Service
}

const themeExtendInfoKey = "theme_extend_info"

func NewWebsiteInfoHandler(sysCfg *sysconfig.Service) *WebsiteInfoHandler {
	return &WebsiteInfoHandler{sysCfg: sysCfg}
}

func (h *WebsiteInfoHandler) listPublic(c *fiber.Ctx) error {
	info, err := h.sysCfg.WebsiteInfo(c.Context())
	if err != nil {
		return err
	}
	payload := make(map[string]any, len(info))
	for key, value := range info {
		if key == themeExtendInfoKey {
			payload[key] = parseJSONStringOrEmpty(value)
			continue
		}
		payload[key] = value
	}
	return response.Success(c, payload)
}

// PublicList godoc
// @Summary 公开获取网站信息
// @Tags WebsiteInfo
// @Produce json
// @Success 200 {object} contract.GenericMessageEnvelope
// @Router /public/website-info [get]
func (h *WebsiteInfoHandler) PublicList(c *fiber.Ctx) error {
	return h.listPublic(c)
}

// List godoc
// @Summary 获取全部网站信息（需要 config:read 权限）
// @Tags WebsiteInfo
// @Produce json
// @Success 200 {object} contract.WebsiteInfoListRespEnvelope
// @Security BearerAuth
// @Router /website-info [get]
func (h *WebsiteInfoHandler) List(c *fiber.Ctx) error {
	info, err := h.sysCfg.WebsiteInfo(c.Context())
	if err != nil {
		return err
	}
	items := make([]contract.WebsiteInfoResp, 0, len(info))
	for key, value := range info {
		item := contract.WebsiteInfoResp{
			Key:   key,
			Value: &value,
		}
		if key == themeExtendInfoKey {
			raw := json.RawMessage(value)
			item.InfoJSON = &raw
			item.Value = nil
		}
		items = append(items, item)
	}
	return response.Success(c, items)
}

// Update godoc
// @Summary 更新网站信息
// @Tags WebsiteInfo
// @Accept json
// @Produce json
// @Param key path string true "配置键"
// @Param request body contract.WebsiteInfoReq true "网站配置"
// @Success 200 {object} contract.WebsiteInfoDetailRespEnvelope
// @Security BearerAuth
// @Router /website-info/{key} [put]
func (h *WebsiteInfoHandler) Update(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "key 不能为空")
	}
	var req contract.WebsiteInfoReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := validateWebsiteInfoReq(key, req); err != nil {
		return err
	}

	if strings.TrimSpace(key) == themeExtendInfoKey {
		if err := h.sysCfg.UpdateWebsiteInfoJSON(c.Context(), key, json.RawMessage(*req.InfoJSON)); err != nil {
			return err
		}
	} else {
		if err := h.sysCfg.UpdateWebsiteInfoByKey(c.Context(), key, *req.Value); err != nil {
			return err
		}
	}

	Audit(c, "website_info.update", map[string]any{"key": key})
	resp := contract.WebsiteInfoResp{Key: key, Value: req.Value}
	if req.InfoJSON != nil {
		raw := json.RawMessage(*req.InfoJSON)
		resp.InfoJSON = &raw
	}
	return response.SuccessWithMessage(c, resp, "updated")
}

func validateWebsiteInfoReq(key string, req contract.WebsiteInfoReq) error {
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "key 不能为空")
	}
	if trimmedKey == themeExtendInfoKey {
		if req.Value != nil {
			return response.NewBizErrorWithMsg(response.ParamsError, "theme_extend_info 仅支持 infoJson")
		}
		if req.InfoJSON == nil {
			return response.NewBizErrorWithMsg(response.ParamsError, "theme_extend_info 的 infoJson 不能为空")
		}
		if !json.Valid(json.RawMessage(*req.InfoJSON)) {
			return response.NewBizErrorWithMsg(response.ParamsError, "infoJson 不是合法 JSON")
		}
		return nil
	}
	if req.InfoJSON != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "普通配置不支持 infoJson")
	}
	if req.Value == nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "value 不能为空")
	}
	return nil
}

func parseJSONStringOrEmpty(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return map[string]any{}
	}
	var result any
	if err := json.Unmarshal([]byte(trimmed), &result); err != nil {
		return map[string]any{}
	}
	return result
}
