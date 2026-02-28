package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/email"
	domainemail "github.com/grtsinry43/grtblog-v2/server/internal/domain/email"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type EmailTemplateHandler struct {
	svc *email.Service
}

func NewEmailTemplateHandler(svc *email.Service) *EmailTemplateHandler {
	return &EmailTemplateHandler{svc: svc}
}

// ListEmailEvents godoc
// @Summary 获取邮件事件列表
// @Tags Email
// @Produce json
// @Success 200 {object} contract.EmailEventListResp
// @Security BearerAuth
// @Router /admin/email/events [get]
// @Security JWTAuth
func (h *EmailTemplateHandler) ListEmailEvents(c *fiber.Ctx) error {
	return response.Success(c, contract.EmailEventListResp{Events: h.svc.ListEvents()})
}

// ListPublicEmailEvents godoc
// @Summary 获取可订阅邮件事件列表
// @Tags EmailPublic
// @Produce json
// @Success 200 {object} contract.EmailEventListResp
// @Router /public/email/events [get]
func (h *EmailTemplateHandler) ListPublicEmailEvents(c *fiber.Ctx) error {
	return response.Success(c, contract.EmailEventListResp{Events: h.svc.ListPublicEvents()})
}

// ListEmailEventCatalog godoc
// @Summary 获取邮件事件参数目录
// @Tags Email
// @Produce json
// @Success 200 {object} contract.EmailEventCatalogResp
// @Security BearerAuth
// @Router /admin/email/events/catalog [get]
// @Security JWTAuth
func (h *EmailTemplateHandler) ListEmailEventCatalog(c *fiber.Ctx) error {
	items := h.svc.ListEventCatalog()
	respItems := make([]contract.EmailEventDescriptorResp, len(items))
	for i, item := range items {
		fields := make([]contract.EmailEventFieldResp, len(item.Fields))
		for j, field := range item.Fields {
			fields[j] = contract.EmailEventFieldResp{
				Name:        field.Name,
				Type:        field.Type,
				Required:    field.Required,
				Description: field.Description,
			}
		}
		respItems[i] = contract.EmailEventDescriptorResp{
			Name:        item.Name,
			Title:       item.Title,
			Category:    item.Category,
			Public:      item.PublicEmail,
			Description: item.Description,
			Fields:      fields,
		}
	}
	return response.Success(c, contract.EmailEventCatalogResp{Items: respItems})
}

// ListEmailTemplates godoc
// @Summary 获取邮件模板列表
// @Tags Email
// @Produce json
// @Success 200 {object} []contract.EmailTemplateResp
// @Security BearerAuth
// @Router /admin/email/templates [get]
// @Security JWTAuth
func (h *EmailTemplateHandler) ListEmailTemplates(c *fiber.Ctx) error {
	items, err := h.svc.ListTemplates(c.Context())
	if err != nil {
		return err
	}
	resp := make([]contract.EmailTemplateResp, len(items))
	for i, item := range items {
		resp[i] = mapEmailTemplateResp(item)
	}
	return response.Success(c, resp)
}

// CreateEmailTemplate godoc
// @Summary 创建邮件模板
// @Tags Email
// @Accept json
// @Produce json
// @Param request body contract.CreateEmailTemplateReq true "创建邮件模板参数"
// @Success 200 {object} contract.EmailTemplateResp
// @Security BearerAuth
// @Router /admin/email/templates [post]
// @Security JWTAuth
func (h *EmailTemplateHandler) CreateEmailTemplate(c *fiber.Ctx) error {
	var req contract.CreateEmailTemplateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	tpl := &domainemail.Template{
		Code:            req.Code,
		Name:            req.Name,
		EventName:       req.EventName,
		SubjectTemplate: req.SubjectTemplate,
		HTMLTemplate:    req.HTMLTemplate,
		TextTemplate:    req.TextTemplate,
		ToEmails:        req.ToEmails,
		IsEnabled:       req.IsEnabled,
	}
	if err := h.svc.CreateTemplate(c.Context(), tpl); err != nil {
		Audit(c, "email.template.create_failed", map[string]any{
			"code":      tpl.Code,
			"eventName": tpl.EventName,
			"error":     err.Error(),
		})
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.SuccessWithMessage(c, mapEmailTemplateResp(tpl), "邮件模板创建成功")
}

// UpdateEmailTemplate godoc
// @Summary 更新邮件模板
// @Tags Email
// @Accept json
// @Produce json
// @Param code path string true "模板编码"
// @Param request body contract.UpdateEmailTemplateReq true "更新邮件模板参数"
// @Success 200 {object} contract.EmailTemplateResp
// @Security BearerAuth
// @Router /admin/email/templates/{code} [put]
// @Security JWTAuth
func (h *EmailTemplateHandler) UpdateEmailTemplate(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的模板编码")
	}
	var req contract.UpdateEmailTemplateReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	tpl := &domainemail.Template{
		Name:            req.Name,
		EventName:       req.EventName,
		SubjectTemplate: req.SubjectTemplate,
		HTMLTemplate:    req.HTMLTemplate,
		TextTemplate:    req.TextTemplate,
		ToEmails:        req.ToEmails,
		IsEnabled:       req.IsEnabled,
	}
	if err := h.svc.UpdateTemplate(c.Context(), code, tpl); err != nil {
		Audit(c, "email.template.update_failed", map[string]any{
			"code":      code,
			"eventName": tpl.EventName,
			"error":     err.Error(),
		})
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	updated, err := h.svc.GetTemplateByCode(c.Context(), code)
	if err != nil {
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.SuccessWithMessage(c, mapEmailTemplateResp(updated), "邮件模板更新成功")
}

// DeleteEmailTemplate godoc
// @Summary 删除邮件模板
// @Tags Email
// @Produce json
// @Param code path string true "模板编码"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /admin/email/templates/{code} [delete]
// @Security JWTAuth
func (h *EmailTemplateHandler) DeleteEmailTemplate(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的模板编码")
	}
	if err := h.svc.DeleteTemplate(c.Context(), code); err != nil {
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.SuccessWithMessage[any](c, nil, "邮件模板删除成功")
}

// PreviewEmailTemplate godoc
// @Summary 预览邮件模板渲染结果
// @Tags Email
// @Accept json
// @Produce json
// @Param code path string true "模板编码"
// @Param request body contract.EmailTemplatePreviewReq false "预览参数"
// @Success 200 {object} contract.EmailTemplatePreviewResp
// @Security BearerAuth
// @Router /admin/email/templates/{code}/preview [post]
// @Security JWTAuth
func (h *EmailTemplateHandler) PreviewEmailTemplate(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的模板编码")
	}
	var req contract.EmailTemplatePreviewReq
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
		}
	}
	variables, err := parseEmailVariables(req.Variables)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "variables 无效", err)
	}
	rendered, err := h.svc.PreviewTemplate(c.Context(), code, variables)
	if err != nil {
		Audit(c, "email.template.preview_failed", map[string]any{
			"code":  code,
			"keys":  mapKeys(variables),
			"error": err.Error(),
		})
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.Success(c, contract.EmailTemplatePreviewResp{
		Subject:  rendered.Subject,
		HTMLBody: rendered.HTMLBody,
		TextBody: rendered.TextBody,
	})
}

// TestEmailTemplate godoc
// @Summary 测试发送邮件模板
// @Tags Email
// @Accept json
// @Produce json
// @Param code path string true "模板编码"
// @Param request body contract.EmailTemplateTestReq false "测试发送参数"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /admin/email/templates/{code}/test [post]
// @Security JWTAuth
func (h *EmailTemplateHandler) TestEmailTemplate(c *fiber.Ctx) error {
	code := strings.TrimSpace(c.Params("code"))
	if code == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的模板编码")
	}
	var req contract.EmailTemplateTestReq
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
		}
	}
	variables, err := parseEmailVariables(req.Variables)
	if err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "variables 无效", err)
	}
	if err := h.svc.TestSend(c.Context(), code, req.ToEmails, variables); err != nil {
		Audit(c, "email.template.test_failed", map[string]any{
			"code":    code,
			"toCount": len(req.ToEmails),
			"keys":    mapKeys(variables),
			"error":   err.Error(),
		})
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.SuccessWithMessage[any](c, nil, "测试邮件发送成功")
}

// SubscribeEmail godoc
// @Summary 订阅邮件事件
// @Tags EmailPublic
// @Accept json
// @Produce json
// @Param request body contract.EmailSubscribeReq true "订阅参数(使用 eventNames)"
// @Success 200 {object} contract.EmailSubscribeBatchResp
// @Router /public/email/subscriptions [post]
func (h *EmailTemplateHandler) SubscribeEmail(c *fiber.Ctx) error {
	var req contract.EmailSubscribeReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	items, err := h.svc.SubscribeBatch(c.Context(), req.Email, req.EventNames, c.IP())
	if err != nil {
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	respItems := make([]contract.EmailPublicSubscriptionResp, len(items))
	for i, item := range items {
		respItems[i] = mapEmailPublicSubscriptionResp(item)
	}
	return response.SuccessWithMessage(c, contract.EmailSubscribeBatchResp{Items: respItems}, "订阅成功")
}

// UnsubscribeEmail godoc
// @Summary 退订邮件事件
// @Tags EmailPublic
// @Accept json
// @Produce json
// @Param request body contract.EmailUnsubscribeReq true "退订参数"
// @Success 200 {object} any
// @Router /public/email/subscriptions/unsubscribe [post]
func (h *EmailTemplateHandler) UnsubscribeEmail(c *fiber.Ctx) error {
	var req contract.EmailUnsubscribeReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.svc.Unsubscribe(c.Context(), req.Token, req.Email, req.EventName); err != nil {
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.SuccessWithMessage[any](c, nil, "退订成功")
}

// ListEmailSubscriptions godoc
// @Summary 获取邮件订阅列表（管理端）
// @Tags Email
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param eventName query string false "事件名"
// @Param status query string false "状态"
// @Param search query string false "邮箱关键字"
// @Success 200 {object} contract.EmailSubscriptionListResp
// @Security BearerAuth
// @Router /admin/email/subscriptions [get]
// @Security JWTAuth
func (h *EmailTemplateHandler) ListEmailSubscriptions(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 20)
	eventName := strings.TrimSpace(c.Query("eventName"))
	status := strings.TrimSpace(c.Query("status"))
	search := strings.TrimSpace(c.Query("search"))

	var eventNamePtr *string
	var statusPtr *string
	var searchPtr *string
	if eventName != "" {
		eventNamePtr = &eventName
	}
	if status != "" {
		statusPtr = &status
	}
	if search != "" {
		searchPtr = &search
	}
	items, total, err := h.svc.ListSubscriptions(c.Context(), domainemail.SubscriptionListOptions{
		Page:      page,
		PageSize:  pageSize,
		EventName: eventNamePtr,
		Status:    statusPtr,
		Search:    searchPtr,
	})
	if err != nil {
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	respItems := make([]contract.EmailSubscriptionResp, len(items))
	for i, item := range items {
		respItems[i] = mapEmailSubscriptionResp(item, false)
	}
	return response.Success(c, contract.EmailSubscriptionListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// BatchUpdateEmailSubscriptionStatus godoc
// @Summary 批量更新邮件订阅状态（管理端）
// @Tags Email
// @Accept json
// @Produce json
// @Param request body contract.BatchUpdateEmailSubscriptionStatusReq true "批量更新状态参数"
// @Success 200 {object} any
// @Security BearerAuth
// @Router /admin/email/subscriptions/status [put]
// @Security JWTAuth
func (h *EmailTemplateHandler) BatchUpdateEmailSubscriptionStatus(c *fiber.Ctx) error {
	var req contract.BatchUpdateEmailSubscriptionStatusReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if err := h.svc.BatchUpdateSubscriptionStatus(c.Context(), req.IDs, req.Status); err != nil {
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.SuccessWithMessage[any](c, nil, "批量更新订阅状态成功")
}

// ListEmailOutbox godoc
// @Summary 获取邮件出站队列
// @Tags Email
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param status query string false "状态"
// @Param eventName query string false "事件名"
// @Param search query string false "主题关键字"
// @Success 200 {object} contract.EmailOutboxListResp
// @Security BearerAuth
// @Router /admin/email/outbox [get]
// @Security JWTAuth
func (h *EmailTemplateHandler) ListEmailOutbox(c *fiber.Ctx) error {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "pageSize", 20)
	status := strings.TrimSpace(c.Query("status"))
	eventName := strings.TrimSpace(c.Query("eventName"))
	search := strings.TrimSpace(c.Query("search"))

	var statusPtr, eventNamePtr, searchPtr *string
	if status != "" {
		statusPtr = &status
	}
	if eventName != "" {
		eventNamePtr = &eventName
	}
	if search != "" {
		searchPtr = &search
	}
	items, total, err := h.svc.ListOutbox(c.Context(), domainemail.OutboxListOptions{
		Page:      page,
		PageSize:  pageSize,
		Status:    statusPtr,
		EventName: eventNamePtr,
		Search:    searchPtr,
	})
	if err != nil {
		return err
	}
	respItems := make([]contract.EmailOutboxResp, len(items))
	for i, item := range items {
		respItems[i] = mapEmailOutboxResp(item, false)
	}
	return response.Success(c, contract.EmailOutboxListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

// GetEmailOutbox godoc
// @Summary 获取邮件出站详情
// @Tags Email
// @Produce json
// @Param id path int true "出站记录 ID"
// @Success 200 {object} contract.EmailOutboxResp
// @Security BearerAuth
// @Router /admin/email/outbox/{id} [get]
// @Security JWTAuth
func (h *EmailTemplateHandler) GetEmailOutbox(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的出站记录 ID")
	}
	item, err := h.svc.GetOutboxByID(c.Context(), int64(id))
	if err != nil {
		if mapped := mapEmailDomainError(err); mapped != nil {
			return mapped
		}
		return err
	}
	return response.Success(c, mapEmailOutboxResp(item, true))
}

func mapEmailOutboxResp(item *domainemail.Outbox, withBody bool) contract.EmailOutboxResp {
	toEmails := item.ToEmails
	if toEmails == nil {
		toEmails = []string{}
	}
	resp := contract.EmailOutboxResp{
		ID:           item.ID,
		TemplateCode: item.TemplateCode,
		EventName:    item.EventName,
		ToEmails:     toEmails,
		Subject:      item.Subject,
		Status:       item.Status,
		RetryCount:   item.RetryCount,
		NextRetryAt:  item.NextRetryAt,
		LastError:    item.LastError,
		SentAt:       item.SentAt,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
	if withBody {
		resp.HTMLBody = item.HTMLBody
		resp.TextBody = item.TextBody
	}
	return resp
}

func mapEmailTemplateResp(item *domainemail.Template) contract.EmailTemplateResp {
	toEmails := item.ToEmails
	if toEmails == nil {
		toEmails = []string{}
	}
	return contract.EmailTemplateResp{
		ID:              item.ID,
		Code:            item.Code,
		Name:            item.Name,
		EventName:       item.EventName,
		SubjectTemplate: item.SubjectTemplate,
		HTMLTemplate:    item.HTMLTemplate,
		TextTemplate:    item.TextTemplate,
		ToEmails:        toEmails,
		IsEnabled:       item.IsEnabled,
		IsInternal:      item.IsInternal,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func mapEmailSubscriptionResp(item *domainemail.Subscription, withToken bool) contract.EmailSubscriptionResp {
	resp := contract.EmailSubscriptionResp{
		ID:             item.ID,
		Email:          item.Email,
		EventName:      item.EventName,
		Status:         item.Status,
		SourceIP:       item.SourceIP,
		UnsubscribedAt: item.UnsubscribedAt,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
	if withToken {
		resp.Token = item.Token
	}
	return resp
}

func mapEmailPublicSubscriptionResp(item *domainemail.Subscription) contract.EmailPublicSubscriptionResp {
	return contract.EmailPublicSubscriptionResp{
		ID:        item.ID,
		Email:     item.Email,
		EventName: item.EventName,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func parseEmailVariables(raw json.RawMessage) (map[string]any, error) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" || strings.TrimSpace(string(raw)) == "null" {
		return map[string]any{}, nil
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	result := map[string]any{}
	if err := decoder.Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func mapKeys(input map[string]any) []string {
	if len(input) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	return keys
}

func mapEmailDomainError(err error) error {
	switch {
	case errors.Is(err, domainemail.ErrEmailTemplateNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "邮件模板不存在")
	case errors.Is(err, domainemail.ErrEmailTemplateCodeExists):
		return response.NewBizErrorWithMsg(response.ParamsError, "模板编码已存在")
	case errors.Is(err, domainemail.ErrEmailTemplateEventInvalid):
		return response.NewBizErrorWithMsg(response.ParamsError, "事件名称无效")
	case errors.Is(err, domainemail.ErrEmailTemplateRenderFailed):
		return response.NewBizErrorWithMsg(response.ParamsError, "模板内容无效或渲染失败")
	case errors.Is(err, domainemail.ErrEmailTemplateInternalLocked):
		return response.NewBizErrorWithMsg(response.ParamsError, "内置模板不允许删除")
	case errors.Is(err, domainemail.ErrEmailNoRecipient):
		return response.NewBizErrorWithMsg(response.ParamsError, "收件人为空，请配置模板收件人或 email.defaultTo")
	case errors.Is(err, domainemail.ErrEmailDisabled):
		return response.NewBizErrorWithMsg(response.ParamsError, "邮件功能未启用")
	case errors.Is(err, domainemail.ErrEmailConfigInvalid):
		return response.NewBizErrorWithMsg(response.ParamsError, "SMTP 配置不完整")
	case errors.Is(err, domainemail.ErrEmailSendFailed):
		return response.NewBizErrorWithMsg(response.ServerError, "邮件发送失败")
	case errors.Is(err, domainemail.ErrEmailSubscriptionInvalid):
		return response.NewBizErrorWithMsg(response.ParamsError, "订阅参数无效")
	case errors.Is(err, domainemail.ErrEmailSubscriptionEventInvalid):
		return response.NewBizErrorWithMsg(response.ParamsError, "订阅事件无效")
	case errors.Is(err, domainemail.ErrEmailSubscriptionNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "订阅记录不存在")
	case errors.Is(err, domainemail.ErrEmailSubscriptionStatusInvalid):
		return response.NewBizErrorWithMsg(response.ParamsError, "订阅状态无效")
	case errors.Is(err, domainemail.ErrEmailOutboxNotFound):
		return response.NewBizErrorWithMsg(response.NotFound, "邮件出站记录不存在")
	default:
		return nil
	}
}
