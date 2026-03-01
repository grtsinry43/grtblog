package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	domainai "github.com/grtsinry43/grtblog-v2/server/internal/domain/ai"
	domainconfig "github.com/grtsinry43/grtblog-v2/server/internal/domain/config"
	infraai "github.com/grtsinry43/grtblog-v2/server/internal/infra/ai"
)

// SysConfigReader 提供读取 sys_config 的能力，由 sysconfig.Service 实现。
type SysConfigReader interface {
	GetConfigValue(ctx context.Context, key string) (string, error)
}

type Service struct {
	repo   domainai.Repository
	cfgGet SysConfigReader
}

func NewService(repo domainai.Repository, cfgGet SysConfigReader) *Service {
	return &Service{repo: repo, cfgGet: cfgGet}
}

// ── AI 功能调用 ──

// ModerationResult 评论审核结果。
type ModerationResult struct {
	Approved bool    `json:"approved"`
	Reason   string  `json:"reason"`
	Score    float64 `json:"score"`
}

// TitleResult 标题生成结果。
type TitleResult struct {
	Title    string `json:"title"`
	ShortURL string `json:"shortUrl"`
}

// RewriteResult 内容改写结果。
type RewriteResult struct {
	Content string `json:"content"`
}

const (
	taskKeyCommentModeration  = "ai.task.commentModeration.modelId"
	taskKeyTitleGeneration    = "ai.task.titleGeneration.modelId"
	taskKeyContentRewrite     = "ai.task.contentRewrite.modelId"
	taskKeySummaryGeneration  = "ai.task.summaryGeneration.modelId"

	promptKeyCommentModeration  = "ai.prompt.commentModeration"
	promptKeyTitleGeneration    = "ai.prompt.titleGeneration"
	promptKeyContentRewrite     = "ai.prompt.contentRewrite"
	promptKeySummaryGeneration  = "ai.prompt.summaryGeneration"

	defaultModerationPrompt = `你是一个博客评论审核助手。请判断以下评论是否应该通过审核。
评判标准：
1. 拒绝垃圾广告、恶意链接、无意义灌水内容
2. 拒绝包含侮辱、歧视、仇恨言论的内容
3. 通过正常的讨论、提问、建议、赞赏等内容
请以 JSON 格式返回结果：{"approved": true/false, "reason": "原因说明", "score": 0.0-1.0}
其中 score 表示通过审核的置信度，1.0 表示完全确定应该通过。`

	defaultTitlePrompt = `你是一个博客标题生成助手。请根据以下文章内容生成一个合适的标题和 URL 短链接。
要求：
1. 标题应简洁、有吸引力，准确概括文章主题
2. 短链接应使用英文或拼音，用连字符分隔，全小写，不超过 50 个字符
请以 JSON 格式返回结果：{"title": "生成的标题", "shortUrl": "generated-short-url"}`

	defaultRewritePrompt = `你是一个专业的内容编辑助手。请根据用户的指令对以下内容进行改写或扩写。
要求：
1. 保持原文的核心观点和信息
2. 根据用户指令调整文风、篇幅或表达方式
3. 使用 Markdown 格式输出
请直接返回改写后的内容，不要包含额外的说明。`

	defaultSummaryPrompt = `你是一个博客摘要生成助手。请根据以下文章内容生成一段简洁的摘要，2-3句话概括文章核心内容。请直接返回摘要文本，不要包含额外说明。`
)

// ModerateComment 使用 AI 审核评论内容。
func (s *Service) ModerateComment(ctx context.Context, content string) (*ModerationResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, err := s.buildClient(ctx, taskKeyCommentModeration)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	prompt := s.readPrompt(ctx, promptKeyCommentModeration, defaultModerationPrompt)

	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	var result ModerationResult
	cleaned := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse moderation result: %w (raw: %s)", err, resp.Content)
	}
	return &result, nil
}

// GenerateTitle 使用 AI 根据内容生成标题和短链接。
func (s *Service) GenerateTitle(ctx context.Context, content string) (*TitleResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, err := s.buildClient(ctx, taskKeyTitleGeneration)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	prompt := s.readPrompt(ctx, promptKeyTitleGeneration, defaultTitlePrompt)

	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	var result TitleResult
	cleaned := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse title result: %w (raw: %s)", err, resp.Content)
	}
	return &result, nil
}

// RewriteContent 使用 AI 改写或扩写内容。
func (s *Service) RewriteContent(ctx context.Context, content, instruction string) (*RewriteResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, err := s.buildClient(ctx, taskKeyContentRewrite)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	prompt := s.readPrompt(ctx, promptKeyContentRewrite, defaultRewritePrompt)

	userMsg := content
	if instruction != "" {
		userMsg = fmt.Sprintf("用户指令：%s\n\n原文内容：\n%s", instruction, content)
	}

	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userMsg},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	return &RewriteResult{Content: resp.Content}, nil
}

// RewriteContentStream 使用 AI 流式改写内容，通过 onChunk 回调增量返回。
func (s *Service) RewriteContentStream(ctx context.Context, content, instruction string, onChunk func(string) error) error {
	if err := s.checkEnabled(ctx); err != nil {
		return err
	}

	client, modelID, err := s.buildClient(ctx, taskKeyContentRewrite)
	if err != nil {
		return fmt.Errorf("build ai client: %w", err)
	}

	prompt := s.readPrompt(ctx, promptKeyContentRewrite, defaultRewritePrompt)

	userMsg := content
	if instruction != "" {
		userMsg = fmt.Sprintf("用户指令：%s\n\n原文内容：\n%s", instruction, content)
	}

	return client.ChatStream(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userMsg},
		},
	}, onChunk)
}

// GenerateSummaryStream 使用 AI 流式生成文章摘要，通过 onChunk 回调增量返回。
func (s *Service) GenerateSummaryStream(ctx context.Context, content string, onChunk func(string) error) error {
	if err := s.checkEnabled(ctx); err != nil {
		return err
	}

	client, modelID, err := s.buildClient(ctx, taskKeySummaryGeneration)
	if err != nil {
		return fmt.Errorf("build ai client: %w", err)
	}

	prompt := s.readPrompt(ctx, promptKeySummaryGeneration, defaultSummaryPrompt)

	return client.ChatStream(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	}, onChunk)
}

// ── Provider / Model CRUD 代理 ──

func (s *Service) CreateProvider(ctx context.Context, p *domainai.Provider) error {
	return s.repo.CreateProvider(ctx, p)
}

func (s *Service) GetProviderByID(ctx context.Context, id int64) (*domainai.Provider, error) {
	return s.repo.GetProviderByID(ctx, id)
}

func (s *Service) ListProviders(ctx context.Context) ([]*domainai.Provider, error) {
	return s.repo.ListProviders(ctx)
}

func (s *Service) UpdateProvider(ctx context.Context, p *domainai.Provider) error {
	return s.repo.UpdateProvider(ctx, p)
}

func (s *Service) DeleteProvider(ctx context.Context, id int64) error {
	return s.repo.DeleteProvider(ctx, id)
}

func (s *Service) CreateModel(ctx context.Context, m *domainai.Model) error {
	if _, err := s.repo.GetProviderByID(ctx, m.ProviderID); err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}
	return s.repo.CreateModel(ctx, m)
}

func (s *Service) GetModelByID(ctx context.Context, id int64) (*domainai.Model, error) {
	return s.repo.GetModelByID(ctx, id)
}

func (s *Service) ListModels(ctx context.Context) ([]*domainai.Model, error) {
	return s.repo.ListModels(ctx)
}

func (s *Service) UpdateModel(ctx context.Context, m *domainai.Model) error {
	if _, err := s.repo.GetProviderByID(ctx, m.ProviderID); err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}
	return s.repo.UpdateModel(ctx, m)
}

func (s *Service) DeleteModel(ctx context.Context, id int64) error {
	return s.repo.DeleteModel(ctx, id)
}

// ── 内部辅助方法 ──

func (s *Service) checkEnabled(ctx context.Context) error {
	val, err := s.cfgGet.GetConfigValue(ctx, "ai.enabled")
	if err != nil {
		if err == domainconfig.ErrSysConfigNotFound {
			return fmt.Errorf("AI 功能未启用")
		}
		return err
	}
	enabled, _ := strconv.ParseBool(strings.TrimSpace(val))
	if !enabled {
		return fmt.Errorf("AI 功能未启用")
	}
	return nil
}

func (s *Service) buildClient(ctx context.Context, taskConfigKey string) (infraai.Client, string, error) {
	modelIDStr, err := s.cfgGet.GetConfigValue(ctx, taskConfigKey)
	if err != nil || strings.TrimSpace(modelIDStr) == "" {
		return nil, "", fmt.Errorf("该任务未配置 AI 模型 (key: %s)", taskConfigKey)
	}

	modelID, err := strconv.ParseInt(strings.TrimSpace(modelIDStr), 10, 64)
	if err != nil {
		return nil, "", fmt.Errorf("无效的模型 ID: %s", modelIDStr)
	}

	m, p, err := s.repo.GetModelWithProvider(ctx, modelID)
	if err != nil {
		return nil, "", fmt.Errorf("获取模型信息失败: %w", err)
	}

	if !p.IsActive {
		return nil, "", fmt.Errorf("AI 提供商 %q 已禁用", p.Name)
	}
	if !m.IsActive {
		return nil, "", fmt.Errorf("AI 模型 %q 已禁用", m.Name)
	}

	client, err := infraai.NewClient(p.Type, p.APIURL, p.APIKey)
	if err != nil {
		return nil, "", err
	}

	return client, m.ModelID, nil
}

func (s *Service) readPrompt(ctx context.Context, promptKey, defaultPrompt string) string {
	val, err := s.cfgGet.GetConfigValue(ctx, promptKey)
	if err != nil || strings.TrimSpace(val) == "" {
		return defaultPrompt
	}
	return val
}

// extractJSON 从 AI 回复中提取 JSON 内容（处理 markdown code block 包裹的情况）。
func extractJSON(raw string) string {
	trimmed := strings.TrimSpace(raw)
	// 移除可能的 markdown code block 包裹
	if strings.HasPrefix(trimmed, "```json") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		if idx := strings.LastIndex(trimmed, "```"); idx >= 0 {
			trimmed = trimmed[:idx]
		}
	} else if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```")
		if idx := strings.LastIndex(trimmed, "```"); idx >= 0 {
			trimmed = trimmed[:idx]
		}
	}
	return strings.TrimSpace(trimmed)
}
