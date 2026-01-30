package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/websiteinfo"
	"github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type WebsiteInfoHandler struct {
	svc *websiteinfo.Service
}

const themeExtendInfoKey = "theme_extend_info"

func NewWebsiteInfoHandler(svc *websiteinfo.Service) *WebsiteInfoHandler {
	return &WebsiteInfoHandler{svc: svc}
}

func (h *WebsiteInfoHandler) listAll(c *fiber.Ctx) error {
	items, err := h.svc.List(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, contract.ToWebsiteInfoListResp(items))
}

func (h *WebsiteInfoHandler) listPublic(c *fiber.Ctx) error {
	items, err := h.svc.List(c.Context())
	if err != nil {
		return err
	}
	payload := make(map[string]any, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			continue
		}
		if key == themeExtendInfoKey {
			payload[key] = parseJSONOrEmpty(item.InfoJSON)
			continue
		}
		if item.Value == nil {
			payload[key] = ""
			continue
		}
		payload[key] = *item.Value
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
	return h.listAll(c)
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
	cmd := websiteinfo.UpdateCmd{
		Key:      key,
		Name:     req.Name,
		Value:    req.Value,
		InfoJSON: contract.RawMessagePtr(req.InfoJSON),
	}
	info, err := h.svc.Update(c.Context(), cmd)
	if err != nil {
		if errors.Is(err, config.ErrWebsiteInfoNotFound) {
			return response.NewBizError(response.NotFound)
		}
		return err
	}
	Audit(c, "website_info.update", map[string]any{"key": info.Key})
	return response.SuccessWithMessage(c, contract.ToWebsiteInfoResp(*info), "updated")
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

func parseJSONOrEmpty(value json.RawMessage) any {
	if len(value) == 0 {
		return map[string]any{}
	}
	var result any
	if err := json.Unmarshal(value, &result); err != nil {
		return map[string]any{}
	}
	return result
}
