package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	appai "github.com/grtsinry43/grtblog-v2/server/internal/app/ai"
	domainai "github.com/grtsinry43/grtblog-v2/server/internal/domain/ai"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
)

type AIHandler struct {
	svc *appai.Service
}

func NewAIHandler(svc *appai.Service) *AIHandler {
	return &AIHandler{svc: svc}
}

// ── Provider CRUD ──

func (h *AIHandler) ListProviders(c *fiber.Ctx) error {
	providers, err := h.svc.ListProviders(c.Context())
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取提供商列表失败", err)
	}
	resp := make([]contract.AIProviderResp, len(providers))
	for i, p := range providers {
		resp[i] = toProviderResp(p)
	}
	return response.Success(c, resp)
}

func (h *AIHandler) CreateProvider(c *fiber.Ctx) error {
	var req contract.CreateAIProviderReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "名称不能为空")
	}
	if !isValidProviderType(req.Type) {
		return response.NewBizErrorWithMsg(response.ParamsError, "提供商类型必须为 openai、openrouter 或 gemini")
	}

	p := &domainai.Provider{
		Name:     strings.TrimSpace(req.Name),
		Type:     req.Type,
		APIURL:   strings.TrimSpace(req.APIURL),
		APIKey:   strings.TrimSpace(req.APIKey),
		IsActive: true,
	}
	if req.IsActive != nil {
		p.IsActive = *req.IsActive
	}

	if err := h.svc.CreateProvider(c.Context(), p); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "创建提供商失败", err)
	}
	return response.SuccessWithMessage(c, toProviderResp(p), "提供商创建成功")
}

func (h *AIHandler) UpdateProvider(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的提供商 ID")
	}

	existing, err := h.svc.GetProviderByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domainai.ErrProviderNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "提供商不存在")
		}
		return response.NewBizErrorWithCause(response.ServerError, "获取提供商失败", err)
	}

	var req contract.UpdateAIProviderReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Type != nil {
		if !isValidProviderType(*req.Type) {
			return response.NewBizErrorWithMsg(response.ParamsError, "提供商类型必须为 openai、openrouter 或 gemini")
		}
		existing.Type = *req.Type
	}
	if req.APIURL != nil {
		existing.APIURL = strings.TrimSpace(*req.APIURL)
	}
	if req.APIKey != nil {
		existing.APIKey = strings.TrimSpace(*req.APIKey)
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateProvider(c.Context(), existing); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "更新提供商失败", err)
	}
	return response.SuccessWithMessage(c, toProviderResp(existing), "提供商更新成功")
}

func (h *AIHandler) DeleteProvider(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的提供商 ID")
	}
	if err := h.svc.DeleteProvider(c.Context(), id); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "删除提供商失败", err)
	}
	return response.SuccessWithMessage[any](c, nil, "提供商删除成功")
}

// ── Model CRUD ──

func (h *AIHandler) ListModels(c *fiber.Ctx) error {
	models, err := h.svc.ListModels(c.Context())
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取模型列表失败", err)
	}

	// 批量获取 provider 信息
	providers, _ := h.svc.ListProviders(c.Context())
	providerMap := make(map[int64]*domainai.Provider, len(providers))
	for _, p := range providers {
		providerMap[p.ID] = p
	}

	resp := make([]contract.AIModelResp, len(models))
	for i, m := range models {
		r := toModelResp(m)
		if p, ok := providerMap[m.ProviderID]; ok {
			r.ProviderName = p.Name
			r.ProviderType = p.Type
		}
		resp[i] = r
	}
	return response.Success(c, resp)
}

func (h *AIHandler) CreateModel(c *fiber.Ctx) error {
	var req contract.CreateAIModelReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Name) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "模型名称不能为空")
	}
	if strings.TrimSpace(req.ModelID) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "模型 ID 不能为空")
	}
	if req.ProviderID <= 0 {
		return response.NewBizErrorWithMsg(response.ParamsError, "请选择提供商")
	}

	m := &domainai.Model{
		ProviderID: req.ProviderID,
		Name:       strings.TrimSpace(req.Name),
		ModelID:    strings.TrimSpace(req.ModelID),
		IsActive:   true,
	}
	if req.IsActive != nil {
		m.IsActive = *req.IsActive
	}

	if err := h.svc.CreateModel(c.Context(), m); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "创建模型失败", err)
	}
	return response.SuccessWithMessage(c, toModelResp(m), "模型创建成功")
}

func (h *AIHandler) UpdateModel(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的模型 ID")
	}

	existing, err := h.svc.GetModelByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domainai.ErrModelNotFound) {
			return response.NewBizErrorWithMsg(response.NotFound, "模型不存在")
		}
		return response.NewBizErrorWithCause(response.ServerError, "获取模型失败", err)
	}

	var req contract.UpdateAIModelReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	if req.ProviderID != nil {
		existing.ProviderID = *req.ProviderID
	}
	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.ModelID != nil {
		existing.ModelID = strings.TrimSpace(*req.ModelID)
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.svc.UpdateModel(c.Context(), existing); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "更新模型失败", err)
	}
	return response.SuccessWithMessage(c, toModelResp(existing), "模型更新成功")
}

func (h *AIHandler) DeleteModel(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的模型 ID")
	}
	if err := h.svc.DeleteModel(c.Context(), id); err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "删除模型失败", err)
	}
	return response.SuccessWithMessage[any](c, nil, "模型删除成功")
}

// ── AI 功能 ──

func (h *AIHandler) ModerateComment(c *fiber.Ctx) error {
	var req contract.AIModerateCommentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "评论内容不能为空")
	}

	result, err := h.svc.ModerateComment(c.Context(), req.Content, "manual")
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, err.Error(), err)
	}
	return response.Success(c, contract.AIModerateCommentResp{
		Approved: result.Approved,
		Reason:   result.Reason,
		Score:    result.Score,
	})
}

func (h *AIHandler) GenerateTitle(c *fiber.Ctx) error {
	var req contract.AIGenerateTitleReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "文章内容不能为空")
	}

	result, err := h.svc.GenerateTitle(c.Context(), req.Content)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, err.Error(), err)
	}
	return response.Success(c, contract.AIGenerateTitleResp{
		Title:    result.Title,
		ShortURL: result.ShortURL,
	})
}

func (h *AIHandler) RewriteContent(c *fiber.Ctx) error {
	var req contract.AIRewriteContentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "内容不能为空")
	}

	result, err := h.svc.RewriteContent(c.Context(), req.Content, req.Instruction)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, err.Error(), err)
	}
	return response.Success(c, contract.AIRewriteContentResp{
		Content: result.Content,
	})
}

func (h *AIHandler) RewriteContentStream(c *fiber.Ctx) error {
	var req contract.AIRewriteContentReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "内容不能为空")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	// 创建独立 context：SetBodyStreamWriter 的回调在单独 goroutine 中运行，
	// fasthttp.RequestCtx 作为 context.Context 在该 goroutine 中不可用（Done() 会 panic）。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()
		err := h.svc.RewriteContentStream(ctx, req.Content, req.Instruction, func(chunk string) error {
			data, _ := json.Marshal(map[string]string{"content": chunk})
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
			return w.Flush()
		})
		if err != nil {
			data, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		w.Flush()
	})
	return nil
}

func (h *AIHandler) GenerateSummaryStream(c *fiber.Ctx) error {
	var req contract.AIGenerateSummaryReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "内容不能为空")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()
		err := h.svc.GenerateSummaryStream(ctx, req.Content, func(chunk string) error {
			data, _ := json.Marshal(map[string]string{"content": chunk})
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
			return w.Flush()
		})
		if err != nil {
			data, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		w.Flush()
	})
	return nil
}

// ── TaskLog ──

func (h *AIHandler) ListTaskLogs(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	opts := domainai.TaskLogListOptions{
		Page:     page,
		PageSize: pageSize,
	}
	if v := c.Query("taskType"); v != "" {
		opts.TaskType = &v
	}
	if v := c.Query("status"); v != "" {
		opts.Status = &v
	}
	if v := c.Query("search"); v != "" {
		opts.Search = &v
	}

	items, total, err := h.svc.ListTaskLogs(c.Context(), opts)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取任务日志列表失败", err)
	}

	respItems := make([]contract.AITaskLogResp, len(items))
	for i, l := range items {
		respItems[i] = toTaskLogResp(l)
	}
	return response.Success(c, contract.AITaskLogListResp{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

func (h *AIHandler) GetTaskLog(c *fiber.Ctx) error {
	id, err := parseInt64Param(c, "id")
	if err != nil {
		return response.NewBizErrorWithMsg(response.ParamsError, "无效的任务日志 ID")
	}

	l, err := h.svc.GetTaskLogByID(c.Context(), id)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取任务日志失败", err)
	}
	return response.Success(c, toTaskLogResp(l))
}

// ── Helpers ──

func isValidProviderType(t string) bool {
	return t == "openai" || t == "openrouter" || t == "gemini"
}

func toProviderResp(p *domainai.Provider) contract.AIProviderResp {
	return contract.AIProviderResp{
		ID:        p.ID,
		Name:      p.Name,
		Type:      p.Type,
		APIURL:    p.APIURL,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func toModelResp(m *domainai.Model) contract.AIModelResp {
	return contract.AIModelResp{
		ID:         m.ID,
		ProviderID: m.ProviderID,
		Name:       m.Name,
		ModelID:    m.ModelID,
		IsActive:   m.IsActive,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

func toTaskLogResp(l *domainai.TaskLog) contract.AITaskLogResp {
	return contract.AITaskLogResp{
		ID:            l.ID,
		TaskType:      l.TaskType,
		ModelName:     l.ModelName,
		ProviderName:  l.ProviderName,
		Status:        l.Status,
		InputText:     l.InputText,
		OutputText:    l.OutputText,
		ErrorMessage:  l.ErrorMessage,
		DurationMs:    l.DurationMs,
		TriggerSource: l.TriggerSource,
		CreatedAt:     l.CreatedAt,
		UpdatedAt:     l.UpdatedAt,
	}
}
