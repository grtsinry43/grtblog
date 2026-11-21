package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/adminstats"
	appai "github.com/grtsinry43/grtblog-v2/server/internal/app/ai"
	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domainai "github.com/grtsinry43/grtblog-v2/server/internal/domain/ai"
	domaincomment "github.com/grtsinry43/grtblog-v2/server/internal/domain/comment"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/contract"
	"github.com/grtsinry43/grtblog-v2/server/internal/http/response"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

const (
	dashboardBriefCacheTTL = 6 * time.Hour
	opsReportCacheTTL      = 12 * time.Hour
)

type cachedOpsReportEntry struct {
	result *appai.OpsReportResult
	exp    time.Time
}

type visitorInsightsReader interface {
	GetVisitorInsights(ctx context.Context, days int) (*domaincomment.VisitorInsights, error)
}

type AIHandler struct {
	svc             *appai.Service
	sysCfgSvc       *sysconfig.Service
	statsSvc        *adminstats.Service
	db              *gorm.DB
	commentRepo     visitorInsightsReader
	redis           *redis.Client
	redisPrefix     string
	briefMu         sync.RWMutex
	briefMemCache   *appai.DashboardBriefResult
	briefMemExp     time.Time
	reportMu        sync.RWMutex
	reportMemCaches map[int]cachedOpsReportEntry
}

func NewAIHandler(svc *appai.Service) *AIHandler {
	return &AIHandler{
		svc:             svc,
		reportMemCaches: make(map[int]cachedOpsReportEntry),
	}
}

func (h *AIHandler) SetSysConfigService(svc *sysconfig.Service) {
	h.sysCfgSvc = svc
}

func (h *AIHandler) SetDashboardBriefDeps(statsSvc *adminstats.Service, redisClient *redis.Client, redisPrefix string) {
	h.statsSvc = statsSvc
	h.redis = redisClient
	h.redisPrefix = redisPrefix
}

func (h *AIHandler) SetOpsReportDeps(db *gorm.DB, commentRepo visitorInsightsReader) {
	h.db = db
	h.commentRepo = commentRepo
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
	c.Set("X-Accel-Buffering", "no")

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
	c.Set("X-Accel-Buffering", "no")

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

// CompleteStream 实时幽灵续写流式接口（不写入任务日志）。
func (h *AIHandler) CompleteStream(c *fiber.Ctx) error {
	var req contract.AICompleteReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Prefix) == "" {
		return response.NewBizErrorWithMsg(response.ParamsError, "上文不能为空")
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()
		err := h.svc.CompleteTextStream(ctx, req.Prefix, req.Suffix, req.Title, func(chunk string) error {
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

// CheckWriting 实时行文纠错检查接口（不写入任务日志）。
func (h *AIHandler) CheckWriting(c *fiber.Ctx) error {
	var req contract.AICheckWritingReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Content) == "" {
		return response.Success(c, contract.AICheckWritingResp{Issues: []contract.AIWritingIssue{}})
	}

	issues, err := h.svc.CheckWriting(c.UserContext(), req.Content)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "行文检查失败", err)
	}

	respIssues := make([]contract.AIWritingIssue, 0, len(issues))
	for _, item := range issues {
		respIssues = append(respIssues, contract.AIWritingIssue{
			Original:   item.Original,
			Suggestion: item.Suggestion,
			Reason:     item.Reason,
			Severity:   item.Severity,
		})
	}
	return response.Success(c, contract.AICheckWritingResp{Issues: respIssues})
}

// AnnotateParagraph 段落纸边眉批接口（不写入任务日志）。
func (h *AIHandler) AnnotateParagraph(c *fiber.Ctx) error {
	var req contract.AIAnnotateParagraphReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}
	if strings.TrimSpace(req.Paragraph) == "" {
		return response.Success(c, contract.AIAnnotateParagraphResp{Annotations: []contract.AIMarginAnnotation{}})
	}

	annotations, err := h.svc.AnnotateParagraph(c.UserContext(), req.Paragraph, req.PrevContext, req.Title)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "段落批注失败", err)
	}

	respItems := make([]contract.AIMarginAnnotation, 0, len(annotations))
	for _, item := range annotations {
		respItems = append(respItems, contract.AIMarginAnnotation{
			Kind:       item.Kind,
			Quote:      item.Quote,
			Comment:    item.Comment,
			Suggestion: item.Suggestion,
		})
	}
	return response.Success(c, contract.AIAnnotateParagraphResp{Annotations: respItems})
}

// GetDashboardBrief 获取仪表盘 AI 运营简报（优先读 6 小时缓存，AI 未启用时静默返回 enabled=false）。
func (h *AIHandler) GetDashboardBrief(c *fiber.Ctx) error {
	return h.handleDashboardBrief(c, false)
}

// RefreshDashboardBrief 强制刷新仪表盘 AI 运营简报并更新缓存。
func (h *AIHandler) RefreshDashboardBrief(c *fiber.Ctx) error {
	return h.handleDashboardBrief(c, true)
}

// DashboardBriefStream 流式返回仪表盘 AI 运营简报（缓存命中则直接下发，否则流式推送进度与 headline）。
func (h *AIHandler) DashboardBriefStream(c *fiber.Ctx) error {
	var req struct {
		Force bool `json:"force"`
	}
	_ = c.BodyParser(&req)

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()
		writeEvent := func(payload any) error {
			data, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
			return w.Flush()
		}
		finish := func() {
			fmt.Fprintf(w, "data: [DONE]\n\n")
			w.Flush()
		}

		if h.statsSvc == nil || !h.svc.IsDashboardBriefAvailable(ctx) {
			_ = writeEvent(map[string]any{"type": "disabled"})
			finish()
			return
		}

		if !req.Force {
			if cached, ok := h.getCachedDashboardBrief(ctx); ok {
				_ = writeEvent(map[string]any{
					"type": "result",
					"data": toDashboardBriefResp(cached, true),
				})
				finish()
				return
			}
		}

		_ = writeEvent(map[string]any{
			"type":    "stage",
			"stage":   "collecting",
			"message": "正在汇总近 7 天流量与内容数据…",
		})

		stats, err := h.statsSvc.GetDashboardStats(ctx)
		if err != nil {
			_ = writeEvent(map[string]any{"type": "error", "error": "获取站点数据失败"})
			finish()
			return
		}

		_ = writeEvent(map[string]any{
			"type":    "stage",
			"stage":   "analyzing",
			"message": "正在结合近期趋势生成站点简报…",
		})

		snapshot := buildDashboardSnapshotForAI(stats)
		result, err := h.svc.GenerateDashboardBriefStream(ctx, snapshot, func(headline string) error {
			return writeEvent(map[string]any{
				"type": "headline",
				"text": headline,
			})
		})
		if err != nil {
			_ = writeEvent(map[string]any{"type": "error", "error": err.Error()})
			finish()
			return
		}

		h.setCachedDashboardBrief(ctx, result)
		_ = writeEvent(map[string]any{
			"type": "result",
			"data": toDashboardBriefResp(result, false),
		})
		finish()
	})

	return nil
}

// OpsReportStream 流式生成完整版深度站点运营复盘报告（含主题梳理、访客画像、评论区声音与整体运营建议）。
func (h *AIHandler) OpsReportStream(c *fiber.Ctx) error {
	var req struct {
		Days  int  `json:"days"`
		Force bool `json:"force"`
	}
	_ = c.BodyParser(&req)
	days := 30
	if req.Days == 7 {
		days = 7
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cancel()
		writeEvent := func(payload any) error {
			data, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return err
			}
			return w.Flush()
		}
		finish := func() {
			fmt.Fprintf(w, "data: [DONE]\n\n")
			w.Flush()
		}

		if h.statsSvc == nil || !h.svc.IsOpsReportAvailable(ctx) {
			_ = writeEvent(map[string]any{"type": "disabled"})
			finish()
			return
		}

		if !req.Force {
			if cached, ok := h.getCachedOpsReport(ctx, days); ok {
				_ = writeEvent(map[string]any{
					"type": "result",
					"data": toOpsReportResp(cached, true),
				})
				finish()
				return
			}
		}

		_ = writeEvent(map[string]any{
			"type":    "stage",
			"stage":   "collecting",
			"message": fmt.Sprintf("正在汇总近 %d 天站点趋势与内容结构…", days),
		})

		stats, err := h.statsSvc.GetDashboardStats(ctx)
		if err != nil {
			_ = writeEvent(map[string]any{"type": "error", "error": "获取站点数据失败"})
			finish()
			return
		}

		_ = writeEvent(map[string]any{
			"type":    "stage",
			"stage":   "reading",
			"message": "正在读取访客画像与近期评论区声音…",
		})

		var visitorInsights *domaincomment.VisitorInsights
		if h.commentRepo != nil {
			visitorInsights, _ = h.commentRepo.GetVisitorInsights(ctx, days)
		}
		recentComments := h.loadRecentReaderComments(ctx, 20)

		_ = writeEvent(map[string]any{
			"type":    "stage",
			"stage":   "analyzing",
			"message": "正在整理主题脉络与运营复盘报告…",
		})

		snapshot := buildOpsReportSnapshotForAI(stats, visitorInsights, recentComments, days)
		result, err := h.svc.GenerateOpsReportStream(ctx, snapshot, days, func(summary string) error {
			return writeEvent(map[string]any{
				"type": "summary",
				"text": summary,
			})
		})
		if err != nil {
			_ = writeEvent(map[string]any{"type": "error", "error": err.Error()})
			finish()
			return
		}

		h.setCachedOpsReport(ctx, days, result)
		_ = writeEvent(map[string]any{
			"type": "result",
			"data": toOpsReportResp(result, false),
		})
		finish()
	})

	return nil
}

// GetCachedOpsReportIdeas 仅读取已缓存的运营复盘选题建议，绝不触发新的 AI 调用。
func (h *AIHandler) GetCachedOpsReportIdeas(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if !h.svc.IsOpsReportAvailable(ctx) {
		return response.Success(c, contract.AIOpsReportIdeasResp{
			Ideas: []contract.AIOpsReportTopicIdea{},
		})
	}
	cached, ok := h.getCachedOpsReport(ctx, 30)
	if !ok || len(cached.Topics.Ideas) == 0 {
		cached, ok = h.getCachedOpsReport(ctx, 7)
	}
	if !ok || cached == nil || len(cached.Topics.Ideas) == 0 {
		return response.Success(c, contract.AIOpsReportIdeasResp{
			Ideas: []contract.AIOpsReportTopicIdea{},
		})
	}

	ideas := make([]contract.AIOpsReportTopicIdea, 0, len(cached.Topics.Ideas))
	for _, idea := range cached.Topics.Ideas {
		ideas = append(ideas, contract.AIOpsReportTopicIdea{
			Title: idea.Title,
			Angle: idea.Angle,
			Why:   idea.Why,
		})
	}
	genAt := cached.GeneratedAt
	return response.Success(c, contract.AIOpsReportIdeasResp{
		Ideas:       ideas,
		GeneratedAt: &genAt,
	})
}

// GetFeatureFlags 获取全部 AI 子功能持久化开关状态。
func (h *AIHandler) GetFeatureFlags(c *fiber.Ctx) error {
	flags := h.svc.GetFeatureFlags(c.UserContext())
	return response.Success(c, contract.AIFeatureFlagsResp{
		Enabled:         flags.Enabled,
		RealtimeAssist:  flags.RealtimeAssist,
		GhostCompletion: flags.GhostCompletion,
		MarginNotes:     flags.MarginNotes,
		InlineLint:      flags.InlineLint,
		DashboardBrief:  flags.DashboardBrief,
		OpsReport:       flags.OpsReport,
		OnboardingSeen:  flags.OnboardingSeen,
	})
}

// UpdateFeatureFlags 更新 AI 子功能持久化开关状态（写入 sys_config）。
func (h *AIHandler) UpdateFeatureFlags(c *fiber.Ctx) error {
	if h.sysCfgSvc == nil {
		return response.NewBizErrorWithMsg(response.ServerError, "系统配置服务不可用")
	}
	var req contract.UpdateAIFeatureFlagsReq
	if err := c.BodyParser(&req); err != nil {
		return response.NewBizErrorWithCause(response.ParamsError, "请求体解析失败", err)
	}

	items := make([]sysconfig.UpdateItem, 0, 7)
	addBool := func(key string, val *bool) {
		if val == nil {
			return
		}
		raw := json.RawMessage(strconv.FormatBool(*val))
		items = append(items, sysconfig.UpdateItem{
			Key:   key,
			Value: &raw,
		})
	}

	addBool(appai.FeatureKeyRealtimeAssist, req.RealtimeAssist)
	addBool(appai.FeatureKeyGhostCompletion, req.GhostCompletion)
	addBool(appai.FeatureKeyMarginNotes, req.MarginNotes)
	addBool(appai.FeatureKeyInlineLint, req.InlineLint)
	addBool(appai.FeatureKeyDashboardBrief, req.DashboardBrief)
	addBool(appai.FeatureKeyOpsReport, req.OpsReport)
	addBool(appai.FeatureKeyOnboardingSeen, req.OnboardingSeen)

	if len(items) > 0 {
		if _, err := h.sysCfgSvc.UpdateConfigs(c.UserContext(), items); err != nil {
			return response.NewBizErrorWithCause(response.ServerError, "保存 AI 功能配置失败", err)
		}
	}

	flags := h.svc.GetFeatureFlags(c.UserContext())
	return response.Success(c, contract.AIFeatureFlagsResp{
		Enabled:         flags.Enabled,
		RealtimeAssist:  flags.RealtimeAssist,
		GhostCompletion: flags.GhostCompletion,
		MarginNotes:     flags.MarginNotes,
		InlineLint:      flags.InlineLint,
		DashboardBrief:  flags.DashboardBrief,
		OpsReport:       flags.OpsReport,
		OnboardingSeen:  flags.OnboardingSeen,
	})
}

func (h *AIHandler) opsReportCacheKey(days int) string {
	return fmt.Sprintf("%sadmin:ai:ops_report:%dd:v1", h.redisPrefix, days)
}

func (h *AIHandler) getCachedOpsReport(ctx context.Context, days int) (*appai.OpsReportResult, bool) {
	if h.redis != nil {
		payload, err := h.redis.Get(ctx, h.opsReportCacheKey(days)).Bytes()
		if err == nil {
			var res appai.OpsReportResult
			if json.Unmarshal(payload, &res) == nil && res.Summary != "" {
				return &res, true
			}
		}
	}

	h.reportMu.RLock()
	defer h.reportMu.RUnlock()
	if entry, ok := h.reportMemCaches[days]; ok && entry.result != nil && time.Now().Before(entry.exp) {
		copied := *entry.result
		return &copied, true
	}
	return nil, false
}

func (h *AIHandler) setCachedOpsReport(ctx context.Context, days int, res *appai.OpsReportResult) {
	if res == nil {
		return
	}
	h.reportMu.Lock()
	if h.reportMemCaches == nil {
		h.reportMemCaches = make(map[int]cachedOpsReportEntry)
	}
	copied := *res
	h.reportMemCaches[days] = cachedOpsReportEntry{
		result: &copied,
		exp:    time.Now().Add(opsReportCacheTTL),
	}
	h.reportMu.Unlock()

	if h.redis != nil {
		if payload, err := json.Marshal(res); err == nil {
			cacheCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			_ = h.redis.Set(cacheCtx, h.opsReportCacheKey(days), payload, opsReportCacheTTL).Err()
		}
	}
}

func (h *AIHandler) loadRecentReaderComments(ctx context.Context, limit int) []model.Comment {
	if h.db == nil || limit <= 0 {
		return nil
	}
	var rows []model.Comment
	_ = h.db.WithContext(ctx).
		Model(&model.Comment{}).
		Select("id, nick_name, content, location, created_at").
		Where("status = ? AND is_owner = ?", "approved", false).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows
}

func toOpsReportResp(res *appai.OpsReportResult, cached bool) contract.AIOpsReportResp {
	pillars := make([]contract.AIOpsReportTopicPillar, 0, len(res.Topics.Pillars))
	for _, p := range res.Topics.Pillars {
		pillars = append(pillars, contract.AIOpsReportTopicPillar{
			Name:        p.Name,
			Strength:    p.Strength,
			Observation: p.Observation,
		})
	}
	ideas := make([]contract.AIOpsReportTopicIdea, 0, len(res.Topics.Ideas))
	for _, idea := range res.Topics.Ideas {
		ideas = append(ideas, contract.AIOpsReportTopicIdea{
			Title: idea.Title,
			Angle: idea.Angle,
			Why:   idea.Why,
		})
	}
	actions := make([]contract.AIOpsReportActionItem, 0, len(res.Operations.Actions))
	for _, act := range res.Operations.Actions {
		actions = append(actions, contract.AIOpsReportActionItem{
			Priority: act.Priority,
			Title:    act.Title,
			Detail:   act.Detail,
			Route:    act.Route,
		})
	}
	genAt := res.GeneratedAt
	return contract.AIOpsReportResp{
		Enabled: true,
		Summary: res.Summary,
		Mood:    res.Mood,
		Days:    res.Days,
		Topics: &contract.AIOpsReportTopicsSection{
			Overview: res.Topics.Overview,
			Pillars:  pillars,
			Ideas:    ideas,
		},
		Audience: &contract.AIOpsReportAudienceSection{
			Overview:      res.Audience.Overview,
			Highlights:    res.Audience.Highlights,
			RetentionNote: res.Audience.RetentionNote,
		},
		Comments: &contract.AIOpsReportCommentsSection{
			Overview:   res.Comments.Overview,
			Themes:     res.Comments.Themes,
			Suggestion: res.Comments.Suggestion,
		},
		Operations: &contract.AIOpsReportOperationsSection{
			Overview: res.Operations.Overview,
			Actions:  actions,
		},
		GeneratedAt: &genAt,
		Cached:      cached,
	}
}

func buildOpsReportSnapshotForAI(
	s *adminstats.DashboardStats,
	vi *domaincomment.VisitorInsights,
	comments []model.Comment,
	days int,
) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("【分析时间窗口：近 %d 天】\n\n", days))
	sb.WriteString(buildDashboardSnapshotForAI(s))

	if len(s.TagTop) > 0 {
		sb.WriteString("\n【高频标签 Top 15】\n")
		tags := make([]string, 0, 15)
		for i, t := range s.TagTop {
			if i >= 15 {
				break
			}
			tags = append(tags, fmt.Sprintf("%s(%d)", t.Name, t.Count))
		}
		sb.WriteString(strings.Join(tags, "、") + "\n")
	}

	if len(s.Columns) > 0 {
		sb.WriteString("\n【手记专栏分布】\n")
		cols := make([]string, 0, 8)
		for i, c := range s.Columns {
			if i >= 8 {
				break
			}
			cols = append(cols, fmt.Sprintf("%s(%d)", c.Name, c.Count))
		}
		sb.WriteString(strings.Join(cols, "、") + "\n")
	}

	if len(s.TopMoments) > 0 {
		sb.WriteString("\n【热门手记 Top 5】\n")
		for i, m := range s.TopMoments {
			if i >= 5 {
				break
			}
			sb.WriteString(fmt.Sprintf("%d. 《%s》（浏览 %d / 点赞 %d / 评论 %d）\n", i+1, m.Title, m.Views, m.Likes, m.Comments))
		}
	}

	if vi != nil {
		sb.WriteString("\n【访客画像与互动漏斗】\n")
		sb.WriteString(fmt.Sprintf("- 互动漏斗：浏览访客 %d → 点赞访客 %d（%.1f%%）→ 评论访客 %d（占浏览 %.1f%%）\n",
			vi.Funnel.ViewVisitors, vi.Funnel.LikeVisitors, vi.Funnel.LikeRate*100, vi.Funnel.CommentVisitors, vi.Funnel.CommentRateByView*100))
		sb.WriteString(fmt.Sprintf("- 活跃分层：1日活跃 %d、7日活跃 %d、30日活跃 %d、高粘性访客 %d\n",
			vi.Segments.Active1D, vi.Segments.Active7D, vi.Segments.Active30D, vi.Segments.HighlyEngaged))
		if len(vi.PlatformTop) > 0 {
			sb.WriteString("- 主要系统：" + formatVisitorDist(vi.PlatformTop, 5) + "\n")
		}
		if len(vi.BrowserTop) > 0 {
			sb.WriteString("- 主要浏览器：" + formatVisitorDist(vi.BrowserTop, 5) + "\n")
		}
		if len(vi.LocationTop) > 0 {
			sb.WriteString("- 主要地区：" + formatVisitorDist(vi.LocationTop, 6) + "\n")
		}
	}

	sb.WriteString("\n【近期读者评论抽样（最多 20 条）】\n")
	if len(comments) == 0 {
		sb.WriteString("暂无近期读者评论。\n")
	} else {
		for i, cm := range comments {
			nick := strings.TrimSpace(cm.NickName)
			if nick == "" {
				nick = "匿名读者"
			}
			text := strings.ReplaceAll(strings.TrimSpace(cm.Content), "\n", " ")
			runes := []rune(text)
			if len(runes) > 90 {
				text = string(runes[:90]) + "…"
			}
			sb.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, nick, text))
		}
	}

	return sb.String()
}

func formatVisitorDist(items []domaincomment.VisitorDistributionItem, limit int) string {
	out := make([]string, 0, limit)
	for i, it := range items {
		if i >= limit {
			break
		}
		out = append(out, fmt.Sprintf("%s(%d)", it.Name, it.Count))
	}
	return strings.Join(out, "、")
}

func (h *AIHandler) handleDashboardBrief(c *fiber.Ctx, forceRefresh bool) error {
	if h.statsSvc == nil || !h.svc.IsDashboardBriefAvailable(c.UserContext()) {
		return response.Success(c, contract.AIDashboardBriefResp{Enabled: false})
	}

	if !forceRefresh {
		if cached, ok := h.getCachedDashboardBrief(c.UserContext()); ok {
			return response.Success(c, toDashboardBriefResp(cached, true))
		}
	}

	stats, err := h.statsSvc.GetDashboardStats(c.UserContext())
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "获取站点数据失败", err)
	}

	snapshot := buildDashboardSnapshotForAI(stats)
	result, err := h.svc.GenerateDashboardBrief(c.UserContext(), snapshot)
	if err != nil {
		return response.NewBizErrorWithCause(response.ServerError, "生成运营简报失败", err)
	}

	h.setCachedDashboardBrief(c.UserContext(), result)
	return response.Success(c, toDashboardBriefResp(result, false))
}

func (h *AIHandler) dashboardBriefCacheKey() string {
	return fmt.Sprintf("%sadmin:ai:dashboard_brief:v1", h.redisPrefix)
}

func (h *AIHandler) getCachedDashboardBrief(ctx context.Context) (*appai.DashboardBriefResult, bool) {
	if h.redis != nil {
		payload, err := h.redis.Get(ctx, h.dashboardBriefCacheKey()).Bytes()
		if err == nil {
			var res appai.DashboardBriefResult
			if json.Unmarshal(payload, &res) == nil && res.Headline != "" {
				return &res, true
			}
		}
	}

	h.briefMu.RLock()
	defer h.briefMu.RUnlock()
	if h.briefMemCache != nil && time.Now().Before(h.briefMemExp) {
		copied := *h.briefMemCache
		return &copied, true
	}
	return nil, false
}

func (h *AIHandler) setCachedDashboardBrief(ctx context.Context, res *appai.DashboardBriefResult) {
	if res == nil {
		return
	}
	h.briefMu.Lock()
	copied := *res
	h.briefMemCache = &copied
	h.briefMemExp = time.Now().Add(dashboardBriefCacheTTL)
	h.briefMu.Unlock()

	if h.redis != nil {
		if payload, err := json.Marshal(res); err == nil {
			cacheCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			_ = h.redis.Set(cacheCtx, h.dashboardBriefCacheKey(), payload, dashboardBriefCacheTTL).Err()
		}
	}
}

func toDashboardBriefResp(res *appai.DashboardBriefResult, cached bool) contract.AIDashboardBriefResp {
	highlights := make([]contract.AIDashboardBriefHighlight, 0, len(res.Highlights))
	for _, item := range res.Highlights {
		highlights = append(highlights, contract.AIDashboardBriefHighlight{
			Label:  item.Label,
			Detail: item.Detail,
		})
	}
	actions := make([]contract.AIDashboardBriefAction, 0, len(res.Actions))
	for _, item := range res.Actions {
		actions = append(actions, contract.AIDashboardBriefAction{
			Title:  item.Title,
			Reason: item.Reason,
			Route:  item.Route,
		})
	}
	genAt := res.GeneratedAt
	return contract.AIDashboardBriefResp{
		Enabled:     true,
		Headline:    res.Headline,
		Mood:        res.Mood,
		Highlights:  highlights,
		Actions:     actions,
		GeneratedAt: &genAt,
		Cached:      cached,
	}
}

func buildDashboardSnapshotForAI(s *adminstats.DashboardStats) string {
	var sb strings.Builder
	sb.WriteString("【站点核心概况】\n")
	sb.WriteString(fmt.Sprintf("- 注册用户：%d，全站总浏览：%d，总点赞：%d，总评论：%d，总字数：%d\n",
		s.Overview.Users, s.Interaction.ViewsTotal, s.Interaction.LikesTotal, s.Interaction.CommentsTotal, s.Words.Total))
	sb.WriteString(fmt.Sprintf("- 内容规模：已发布文章 %d 篇（草稿 %d 篇），已发布动态 %d 条（草稿 %d 条），思考 %d 条\n",
		s.Overview.ArticlesPublished, s.Overview.ArticlesDraft, s.Overview.MomentsPublished, s.Overview.MomentsDraft, s.Overview.ThinkingsTotal))
	sb.WriteString(fmt.Sprintf("- 今日最高在线：%d，当前在线：%d\n", s.TodayPeakOnline, s.CurrentOnline))
	sb.WriteString(fmt.Sprintf("- 当前真实待办：待审核评论 %d 条，未读读者评论 %d 条，待审核友链申请 %d 条\n",
		s.Pending.PendingComments, s.Pending.UnviewedComments, s.Pending.FriendLinkApplications))

	// 近 7 天 vs 前 7 天 PV / 评论对比
	recentViews, prevViews := sumWindowCounts(s.ViewTrend, 7)
	recentComments, prevComments := sumWindowCounts(s.CommentTrend, 7)
	sb.WriteString("\n【近 7 天趋势对比（vs 前 7 天）】\n")
	sb.WriteString(fmt.Sprintf("- 浏览量（PV）：近 7 天 %d，前 7 天 %d（%s）\n", recentViews, prevViews, formatDeltaPct(recentViews, prevViews)))
	sb.WriteString(fmt.Sprintf("- 新增评论：近 7 天 %d，前 7 天 %d（%s）\n", recentComments, prevComments, formatDeltaPct(recentComments, prevComments)))

	// 近 30 天发布节奏
	var pubArticles30, pubMoments30, pubThinkings30 int64
	lastPublishDaysAgo := -1
	for i, pt := range s.Trend {
		pubArticles30 += pt.Articles
		pubMoments30 += pt.Moments
		pubThinkings30 += pt.Thinkings
		if pt.Articles > 0 || pt.Moments > 0 || pt.Thinkings > 0 {
			lastPublishDaysAgo = len(s.Trend) - 1 - i
		}
	}
	sb.WriteString(fmt.Sprintf("- 近 30 天发布：文章 %d 篇、动态 %d 条、思考 %d 条", pubArticles30, pubMoments30, pubThinkings30))
	if lastPublishDaysAgo >= 0 {
		sb.WriteString(fmt.Sprintf("（最近一次更新距今 %d 天）\n", lastPublishDaysAgo))
	} else {
		sb.WriteString("（近 30 天暂无新发布）\n")
	}

	if len(s.TopArticles) > 0 {
		sb.WriteString("\n【热门文章 Top 3】\n")
		for i, a := range s.TopArticles {
			if i >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("%d. 《%s》（浏览 %d / 点赞 %d / 评论 %d）\n", i+1, a.Title, a.Views, a.Likes, a.Comments))
		}
	}

	if len(s.Categories) > 0 {
		sb.WriteString("\n【主要分类】\n")
		names := make([]string, 0, 4)
		for i, c := range s.Categories {
			if i >= 4 {
				break
			}
			names = append(names, fmt.Sprintf("%s(%d)", c.Name, c.Count))
		}
		sb.WriteString(strings.Join(names, "、") + "\n")
	}

	if len(s.LocationTop) > 0 {
		sb.WriteString("\n【近 7 天主要访客地区】\n")
		locs := make([]string, 0, 4)
		for i, l := range s.LocationTop {
			if i >= 4 {
				break
			}
			locs = append(locs, fmt.Sprintf("%s(%d)", l.Name, l.Count))
		}
		sb.WriteString(strings.Join(locs, "、") + "\n")
	}

	return sb.String()
}

func sumWindowCounts(series []adminstats.DayCountPoint, days int) (recent int64, prev int64) {
	n := len(series)
	if n == 0 || days <= 0 {
		return 0, 0
	}
	recentStart := n - days
	if recentStart < 0 {
		recentStart = 0
	}
	for i := recentStart; i < n; i++ {
		recent += series[i].Count
	}
	prevStart := recentStart - days
	if prevStart < 0 {
		prevStart = 0
	}
	for i := prevStart; i < recentStart; i++ {
		prev += series[i].Count
	}
	return recent, prev
}

func formatDeltaPct(recent, prev int64) string {
	if prev == 0 {
		if recent == 0 {
			return "持平"
		}
		return "较前 7 天新增"
	}
	pct := float64(recent-prev) / float64(prev) * 100
	if pct > 0.5 {
		return fmt.Sprintf("上升 %.0f%%", pct)
	}
	if pct < -0.5 {
		return fmt.Sprintf("下降 %.0f%%", -pct)
	}
	return "基本持平"
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
