package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"

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
	taskKeyCommentModeration = "ai.task.commentModeration.modelId"
	taskKeyTitleGeneration   = "ai.task.titleGeneration.modelId"
	taskKeyContentRewrite    = "ai.task.contentRewrite.modelId"
	taskKeySummaryGeneration = "ai.task.summaryGeneration.modelId"

	promptKeyCommentModeration = "ai.prompt.commentModeration"
	promptKeyTitleGeneration   = "ai.prompt.titleGeneration"
	promptKeyContentRewrite    = "ai.prompt.contentRewrite"
	promptKeySummaryGeneration = "ai.prompt.summaryGeneration"

	FeatureKeyRealtimeAssist  = "ai.feature.realtimeAssist.enabled"
	FeatureKeyGhostCompletion = "ai.feature.realtimeAssist.ghostCompletion"
	FeatureKeyMarginNotes     = "ai.feature.realtimeAssist.marginNotes"
	FeatureKeyInlineLint      = "ai.feature.realtimeAssist.inlineLint"
	FeatureKeyDashboardBrief  = "ai.feature.dashboardBrief.enabled"
	FeatureKeyOpsReport       = "ai.feature.opsReport.enabled"
	FeatureKeyOnboardingSeen  = "ai.feature.onboardingSeen"

	defaultModerationPrompt = `你是一个博客评论审核助手。请判断 <comment> 标签内的评论是否应该通过审核。
评判标准：
1. 拒绝垃圾广告、恶意链接、无意义灌水内容
2. 拒绝包含侮辱、歧视、仇恨言论的内容
3. 通过正常的讨论、提问、建议、赞赏等内容
请以 JSON 格式返回结果：{"approved": true/false, "reason": "原因说明", "score": 0.0-1.0}
其中 score 表示你对本次 approved 决策的置信度，1.0 表示非常确定。`

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
func (s *Service) ModerateComment(ctx context.Context, content, triggerSource string) (*ModerationResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, modelName, providerName, err := s.buildClientExt(ctx, taskKeyCommentModeration)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	taskLog, startTime := s.recordTaskStart(ctx, domainai.TaskTypeCommentModeration, content, triggerSource, modelName, providerName)

	prompt := ensureModerationPromptSafety(s.readPrompt(ctx, promptKeyCommentModeration, defaultModerationPrompt))
	userContent := WrapCommentForModeration(content)

	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userContent},
		},
	})
	if err != nil {
		s.recordTaskEnd(ctx, taskLog, "", err, startTime)
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	var result ModerationResult
	cleaned := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		parseErr := fmt.Errorf("parse moderation result: %w (raw: %s)", err, resp.Content)
		s.recordTaskEnd(ctx, taskLog, resp.Content, parseErr, startTime)
		return nil, parseErr
	}
	normalizeModerationResult(&result)

	s.recordTaskEnd(ctx, taskLog, resp.Content, nil, startTime)
	return &result, nil
}

// GenerateTitle 使用 AI 根据内容生成标题和短链接。
func (s *Service) GenerateTitle(ctx context.Context, content string) (*TitleResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, modelName, providerName, err := s.buildClientExt(ctx, taskKeyTitleGeneration)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	taskLog, startTime := s.recordTaskStart(ctx, domainai.TaskTypeTitleGeneration, content, domainai.TriggerManual, modelName, providerName)

	prompt := s.readPrompt(ctx, promptKeyTitleGeneration, defaultTitlePrompt)

	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		s.recordTaskEnd(ctx, taskLog, "", err, startTime)
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	var result TitleResult
	cleaned := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		parseErr := fmt.Errorf("parse title result: %w (raw: %s)", err, resp.Content)
		s.recordTaskEnd(ctx, taskLog, resp.Content, parseErr, startTime)
		return nil, parseErr
	}

	s.recordTaskEnd(ctx, taskLog, resp.Content, nil, startTime)
	return &result, nil
}

// RewriteContent 使用 AI 改写或扩写内容。
func (s *Service) RewriteContent(ctx context.Context, content, instruction string) (*RewriteResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, modelName, providerName, err := s.buildClientExt(ctx, taskKeyContentRewrite)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	userMsg := content
	if instruction != "" {
		userMsg = fmt.Sprintf("用户指令：%s\n\n原文内容：\n%s", instruction, content)
	}

	taskLog, startTime := s.recordTaskStart(ctx, domainai.TaskTypeContentRewrite, userMsg, domainai.TriggerManual, modelName, providerName)

	prompt := s.readPrompt(ctx, promptKeyContentRewrite, defaultRewritePrompt)

	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userMsg},
		},
	})
	if err != nil {
		s.recordTaskEnd(ctx, taskLog, "", err, startTime)
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	s.recordTaskEnd(ctx, taskLog, resp.Content, nil, startTime)
	return &RewriteResult{Content: resp.Content}, nil
}

// RewriteContentStream 使用 AI 流式改写内容，通过 onChunk 回调增量返回。
func (s *Service) RewriteContentStream(ctx context.Context, content, instruction string, onChunk func(string) error) error {
	if err := s.checkEnabled(ctx); err != nil {
		return err
	}

	client, modelID, modelName, providerName, err := s.buildClientExt(ctx, taskKeyContentRewrite)
	if err != nil {
		return fmt.Errorf("build ai client: %w", err)
	}

	userMsg := content
	if instruction != "" {
		userMsg = fmt.Sprintf("用户指令：%s\n\n原文内容：\n%s", instruction, content)
	}

	taskLog, startTime := s.recordTaskStart(ctx, domainai.TaskTypeContentRewrite, userMsg, domainai.TriggerManual, modelName, providerName)

	prompt := s.readPrompt(ctx, promptKeyContentRewrite, defaultRewritePrompt)

	var collected strings.Builder
	streamErr := client.ChatStream(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: userMsg},
		},
	}, func(chunk string) error {
		collected.WriteString(chunk)
		return onChunk(chunk)
	})

	s.recordTaskEnd(ctx, taskLog, collected.String(), streamErr, startTime)
	return streamErr
}

// WritingIssue 表示实时行文检查发现的拼写或表达问题。
type WritingIssue struct {
	Original   string `json:"original"`
	Suggestion string `json:"suggestion"`
	Reason     string `json:"reason"`
	Severity   string `json:"severity"` // "typo" | "style"
}

// MarginAnnotation 表示写完段落后生成的非侵入式纸边眉批。
type MarginAnnotation struct {
	Kind       string `json:"kind"`                 // "refine" | "spark" | "structure"
	Quote      string `json:"quote,omitempty"`      // 段落中的简短摘录（可选）
	Comment    string `json:"comment"`              // 简短温和的编辑批注
	Suggestion string `json:"suggestion,omitempty"` // 可选的改写方向或启发句
}

// DashboardBriefHighlight 表示仪表盘 AI 运营简报中的数据观察条目。
type DashboardBriefHighlight struct {
	Label  string `json:"label"`
	Detail string `json:"detail"`
}

// DashboardBriefAction 表示仪表盘 AI 运营简报中的可执行建议。
type DashboardBriefAction struct {
	Title  string `json:"title"`
	Reason string `json:"reason"`
	Route  string `json:"route,omitempty"`
}

// DashboardBriefResult 表示仪表盘 AI 运营简报结构化结果。
type DashboardBriefResult struct {
	Headline    string                    `json:"headline"`
	Mood        string                    `json:"mood"` // "growing" | "steady" | "quiet" | "attention"
	Highlights  []DashboardBriefHighlight `json:"highlights"`
	Actions     []DashboardBriefAction    `json:"actions"`
	GeneratedAt time.Time                 `json:"generatedAt"`
	Cached      bool                      `json:"cached"`
}

// OpsReportTopicPillar 表示博客当前形成的内容主题支柱。
type OpsReportTopicPillar struct {
	Name        string `json:"name"`
	Strength    string `json:"strength"` // "core" | "growing" | "quiet"
	Observation string `json:"observation"`
}

// OpsReportTopicIdea 表示基于站点数据的后续选题方向。
type OpsReportTopicIdea struct {
	Title string `json:"title"`
	Angle string `json:"angle"`
	Why   string `json:"why"`
}

// OpsReportTopicsSection 主题与内容梳理板块。
type OpsReportTopicsSection struct {
	Overview string                 `json:"overview"`
	Pillars  []OpsReportTopicPillar `json:"pillars"`
	Ideas    []OpsReportTopicIdea   `json:"ideas"`
}

// OpsReportAudienceSection 访客画像与留存板块。
type OpsReportAudienceSection struct {
	Overview      string   `json:"overview"`
	Highlights    []string `json:"highlights"`
	RetentionNote string   `json:"retentionNote"`
}

// OpsReportCommentsSection 评论区声音与互动板块。
type OpsReportCommentsSection struct {
	Overview   string   `json:"overview"`
	Themes     []string `json:"themes"`
	Suggestion string   `json:"suggestion"`
}

// OpsReportActionItem 整体运营行动建议。
type OpsReportActionItem struct {
	Priority string `json:"priority"` // "high" | "medium" | "low"
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Route    string `json:"route,omitempty"`
}

// OpsReportOperationsSection 整体运营节奏板块。
type OpsReportOperationsSection struct {
	Overview string                `json:"overview"`
	Actions  []OpsReportActionItem `json:"actions"`
}

// OpsReportResult 表示完整版深度站点运营复盘报告。
type OpsReportResult struct {
	Summary     string                     `json:"summary"`
	Mood        string                     `json:"mood"`
	Days        int                        `json:"days"`
	Topics      OpsReportTopicsSection     `json:"topics"`
	Audience    OpsReportAudienceSection   `json:"audience"`
	Comments    OpsReportCommentsSection   `json:"comments"`
	Operations  OpsReportOperationsSection `json:"operations"`
	GeneratedAt time.Time                  `json:"generatedAt"`
	Cached      bool                       `json:"cached"`
}

const defaultCompletePrompt = `你是一个极度克制、懂中文写作心流的博客实时续写助手。
请根据用户光标前的上文（以及可选的标题和下文），判断此处是否真的需要递上一小段自然引子。
要求：
1. 只输出续写内容本身，绝不重复上文末尾的任何文字，不要输出解释、引号或 Markdown 代码块包裹。
2. 保持极短：只续写半句到一句话（约 12~32 个汉字或对应英文长度），帮作者跨过卡壳即可，绝不长篇大论抢占写作主权。
3. 严禁使用"总而言之"、"综上所述"、"值得注意的是"、"可以说"等套话和 AI 腔，严格贴合上文的语气、人称与节奏。
4. 学会保持沉默：如果上文已经是一个完整收束的句子/段落，或者上下文信息不足以给出不流于套话的续写，请只输出 <SKIP> 五个字符，不要勉强凑字。`

const defaultCheckWritingPrompt = `你是一个极度克制、精准的中文博客校对助手。
请仅检查用户提供的片段中是否存在确定性的硬伤：错别字、"的/地/得"明显误用、中英文标点混用、漏字多字或明显病句。
要求：
1. 宁缺毋滥：不要对个人文风、口语化表达或主观措辞吹毛求疵，没有明显硬伤时必须返回空数组 {"issues": []}。
2. 每次最多返回 3 条最值得修改的问题。
3. "original" 必须是原文中精确存在的连续子串，不可改写。
4. "suggestion" 是可直接替换 "original" 的文本。
5. "reason" 用不超过 12 个字简述原因（如"错别字"、"的/地误用"、"漏字"）。
6. "severity" 只能是 "typo"（错别字/标点/语病）或 "style"（极明显拗口表达）。
7. 严格输出合法 JSON：{"issues":[{"original":"...","suggestion":"...","reason":"...","severity":"typo"}]}`

const defaultMarginAnnotationPrompt = `你是一位安静、克制、有审美的中文专栏编辑，正在作者刚刚写完的段落旁做"纸边眉批"。
你的目标不是批改作业，也不是替作者重写，而是在真正有价值的地方轻声提一句。
请阅读段落（及可选标题、前文语境），判断是否值得留下一条眉批：
1. 宁缺毋滥：如果这一段写得自然通顺、没有明显可提升点，必须返回空数组 {"annotations": []}。不要为了回复而硬夸或硬挑刺。
2. 每次最多返回 1 条最有价值的眉批。
3. "kind" 只能是以下三类之一：
   - "refine"（表达推敲）：指出节奏偏散、用词重复、定语过长或主语突兀等问题，并可给出简洁改法。
   - "spark"（思路延伸）：不代笔，只抛出一个有启发性的视角、追问或可补充的细节（如"这里可以顺手补一句为什么放弃了方案 B"）。
   - "structure"（逻辑结构）：指出前后承接断层、承诺未兑现（如前文说两点只写了一点）或段落过长建议拆分。
4. "quote" 可选，摘录段落中最相关的 4~18 个连续原字（若针对整段可留空）。
5. "comment" 用温和、平实的中文写 1 句话（不超过 55 字），像朋友在纸边留下的便签，杜绝说教感和 AI 腔。
6. "suggestion" 可选：对于 "refine" 可给出一句参考改法或改写指令；对于 "spark"/"structure" 可留空。
7. 严格输出合法 JSON：{"annotations":[{"kind":"refine","quote":"...","comment":"...","suggestion":"..."}]}`

// FeatureFlags 表示所有 AI 子功能的持久化开关状态。
type FeatureFlags struct {
	Enabled         bool `json:"enabled"`
	RealtimeAssist  bool `json:"realtimeAssist"`
	GhostCompletion bool `json:"ghostCompletion"`
	MarginNotes     bool `json:"marginNotes"`
	InlineLint      bool `json:"inlineLint"`
	DashboardBrief  bool `json:"dashboardBrief"`
	OpsReport       bool `json:"opsReport"`
	OnboardingSeen  bool `json:"onboardingSeen"`
}

func (s *Service) readBoolConfig(ctx context.Context, key string, defaultVal bool) bool {
	val, err := s.cfgGet.GetConfigValue(ctx, key)
	if err != nil || strings.TrimSpace(val) == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(val))
	if err != nil {
		return defaultVal
	}
	return parsed
}

// GetFeatureFlags 读取所有 AI 功能开关配置。
func (s *Service) GetFeatureFlags(ctx context.Context) FeatureFlags {
	return FeatureFlags{
		Enabled:         s.readBoolConfig(ctx, "ai.enabled", false),
		RealtimeAssist:  s.readBoolConfig(ctx, FeatureKeyRealtimeAssist, true),
		GhostCompletion: s.readBoolConfig(ctx, FeatureKeyGhostCompletion, false),
		MarginNotes:     s.readBoolConfig(ctx, FeatureKeyMarginNotes, true),
		InlineLint:      s.readBoolConfig(ctx, FeatureKeyInlineLint, true),
		DashboardBrief:  s.readBoolConfig(ctx, FeatureKeyDashboardBrief, true),
		OpsReport:       s.readBoolConfig(ctx, FeatureKeyOpsReport, true),
		OnboardingSeen:  s.readBoolConfig(ctx, FeatureKeyOnboardingSeen, false),
	}
}

// CompleteTextStream 用于编辑器的实时幽灵续写（不记入 ai_task_logs）。
func (s *Service) CompleteTextStream(ctx context.Context, prefix, suffix, title string, onChunk func(string) error) error {
	if err := s.checkEnabled(ctx); err != nil {
		return err
	}
	if !s.readBoolConfig(ctx, FeatureKeyRealtimeAssist, true) || !s.readBoolConfig(ctx, FeatureKeyGhostCompletion, false) {
		return fmt.Errorf("幽灵续写功能未启用")
	}

	client, modelID, _, _, err := s.buildClientExt(ctx, taskKeyContentRewrite)
	if err != nil {
		return fmt.Errorf("build ai client: %w", err)
	}

	var sb strings.Builder
	if strings.TrimSpace(title) != "" {
		sb.WriteString("文章标题：")
		sb.WriteString(strings.TrimSpace(title))
		sb.WriteString("\n\n")
	}
	sb.WriteString("光标前上文：\n")
	sb.WriteString(prefix)
	if strings.TrimSpace(suffix) != "" {
		sb.WriteString("\n\n光标后下文（供衔接参考）：\n")
		sb.WriteString(suffix)
	}

	temp := 0.4
	maxTokens := 96
	return client.ChatStream(ctx, infraai.ChatRequest{
		Model:       modelID,
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: defaultCompletePrompt},
			{Role: "user", Content: sb.String()},
		},
	}, onChunk)
}

// AnnotateParagraph 为刚写完的段落生成非侵入式纸边眉批（不记入 ai_task_logs）。
func (s *Service) AnnotateParagraph(ctx context.Context, paragraph, prevContext, title string) ([]MarginAnnotation, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}
	if !s.readBoolConfig(ctx, FeatureKeyRealtimeAssist, true) || !s.readBoolConfig(ctx, FeatureKeyMarginNotes, true) {
		return []MarginAnnotation{}, nil
	}

	client, modelID, _, _, err := s.buildClientExt(ctx, taskKeyContentRewrite)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	var sb strings.Builder
	if strings.TrimSpace(title) != "" {
		sb.WriteString("文章标题：")
		sb.WriteString(strings.TrimSpace(title))
		sb.WriteString("\n\n")
	}
	if strings.TrimSpace(prevContext) != "" {
		sb.WriteString("前文语境（供连贯性参考，无需批注）：\n")
		sb.WriteString(strings.TrimSpace(prevContext))
		sb.WriteString("\n\n")
	}
	sb.WriteString("待审阅段落：\n")
	sb.WriteString(strings.TrimSpace(paragraph))

	temp := 0.35
	maxTokens := 320
	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model:       modelID,
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: defaultMarginAnnotationPrompt},
			{Role: "user", Content: sb.String()},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	raw := extractJSON(resp.Content)
	var parsed struct {
		Annotations []MarginAnnotation `json:"annotations"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		var list []MarginAnnotation
		if err2 := json.Unmarshal([]byte(raw), &list); err2 != nil {
			return []MarginAnnotation{}, nil
		}
		parsed.Annotations = list
	}

	valid := make([]MarginAnnotation, 0, 1)
	for _, item := range parsed.Annotations {
		comment := strings.TrimSpace(item.Comment)
		if comment == "" {
			continue
		}
		kind := strings.TrimSpace(item.Kind)
		if kind != "refine" && kind != "spark" && kind != "structure" {
			kind = "refine"
		}
		quote := strings.TrimSpace(item.Quote)
		if quote != "" && !strings.Contains(paragraph, quote) {
			quote = ""
		}
		valid = append(valid, MarginAnnotation{
			Kind:       kind,
			Quote:      quote,
			Comment:    comment,
			Suggestion: strings.TrimSpace(item.Suggestion),
		})
		// 每段最多保留 1 条最有价值的眉批，保持克制
		break
	}
	return valid, nil
}

// CheckWriting 用于编辑器的轻量行内纠错与润色检查（不记入 ai_task_logs）。
func (s *Service) CheckWriting(ctx context.Context, content string) ([]WritingIssue, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}
	if !s.readBoolConfig(ctx, FeatureKeyRealtimeAssist, true) || !s.readBoolConfig(ctx, FeatureKeyInlineLint, true) {
		return []WritingIssue{}, nil
	}

	client, modelID, _, _, err := s.buildClientExt(ctx, taskKeyContentRewrite)
	if err != nil {
		return nil, fmt.Errorf("build ai client: %w", err)
	}

	temp := 0.2
	maxTokens := 512
	resp, err := client.Chat(ctx, infraai.ChatRequest{
		Model:       modelID,
		Temperature: &temp,
		MaxTokens:   &maxTokens,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: defaultCheckWritingPrompt},
			{Role: "user", Content: content},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ai chat: %w", err)
	}

	raw := extractJSON(resp.Content)
	var parsed struct {
		Issues []WritingIssue `json:"issues"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		var list []WritingIssue
		if err2 := json.Unmarshal([]byte(raw), &list); err2 != nil {
			return []WritingIssue{}, nil
		}
		parsed.Issues = list
	}

	valid := make([]WritingIssue, 0, len(parsed.Issues))
	for _, issue := range parsed.Issues {
		if issue.Original == "" || issue.Suggestion == "" || issue.Original == issue.Suggestion {
			continue
		}
		if !strings.Contains(content, issue.Original) {
			continue
		}
		if issue.Severity != "typo" && issue.Severity != "style" {
			issue.Severity = "style"
		}
		valid = append(valid, issue)
	}
	return valid, nil
}

const defaultDashboardBriefPrompt = `你是一位安静、务实、懂独立博客运营的中文编辑助手，正在为站长撰写仪表盘顶部的"每日运营简报"。
请基于提供的真实站点数据快照，输出克制、有具体数字支撑、无空话套话的简报：
1. "headline"：1 句话总结近期站点最核心的动态（不超过 48 个字），温和自然，避免"总体来看/值得注意的是/表现亮眼"等官话。
2. "mood"：只能是 "growing"（流量或互动上升）、"steady"（平稳运行）、"quiet"（近期更新偏少或流量偏静）、"attention"（有待审核事项或草稿积压需要关注）。
3. "highlights"：2~3 条简短数据观察，每条包含 "label"（不超过 6 字，如"流量走势"、"创作节奏"、"互动概况"）和 "detail"（不超过 42 字，基于快照里的真实数字，不可编造）。
4. "actions"：1~3 条具体可执行的轻建议，每条包含 "title"（不超过 10 字，如"收尾草稿"、"处理待审评论"、"顺势写续篇"）、"reason"（不超过 30 字）和 "route"。
   "route" 只能从以下白名单中选择最贴切的一个，若无对应页面则留空字符串 ""：
   - "/articles/list"（查看文章或草稿）
   - "/articles/edit/new"（新建文章）
   - "/notes/list"（查看动态或草稿）
   - "/notes/edit/new"（发布新动态）
   - "/comments"（处理待审或最新评论）
   - "/friend-links/applications"（审核友链申请）
   - "/audience/visitors"（查看访客画像）
5. 严格输出合法 JSON：{"headline":"...","mood":"steady","highlights":[{"label":"...","detail":"..."}],"actions":[{"title":"...","reason":"...","route":"/articles/list"}]}`

// IsDashboardBriefAvailable 检查仪表盘站点简报功能及可用模型是否已配置。
func (s *Service) IsDashboardBriefAvailable(ctx context.Context) bool {
	if err := s.checkEnabled(ctx); err != nil {
		return false
	}
	if !s.readBoolConfig(ctx, FeatureKeyDashboardBrief, true) {
		return false
	}
	if _, _, _, _, err := s.buildClientExt(ctx, taskKeySummaryGeneration); err == nil {
		return true
	}
	if _, _, _, _, err := s.buildClientExt(ctx, taskKeyContentRewrite); err == nil {
		return true
	}
	return false
}

// IsOpsReportAvailable 检查深度运营复盘与选题灵感功能及可用模型是否已配置。
func (s *Service) IsOpsReportAvailable(ctx context.Context) bool {
	if err := s.checkEnabled(ctx); err != nil {
		return false
	}
	if !s.readBoolConfig(ctx, FeatureKeyOpsReport, true) {
		return false
	}
	if _, _, _, _, err := s.buildClientExt(ctx, taskKeySummaryGeneration); err == nil {
		return true
	}
	if _, _, _, _, err := s.buildClientExt(ctx, taskKeyContentRewrite); err == nil {
		return true
	}
	return false
}

// GenerateDashboardBrief 基于站点数据快照生成结构化 AI 运营简报（不记入 ai_task_logs）。
func (s *Service) GenerateDashboardBrief(ctx context.Context, snapshot string) (*DashboardBriefResult, error) {
	return s.GenerateDashboardBriefStream(ctx, snapshot, nil)
}

// GenerateDashboardBriefStream 流式生成仪表盘 AI 运营简报，实时回调 headline 增量文本。
func (s *Service) GenerateDashboardBriefStream(
	ctx context.Context,
	snapshot string,
	onHeadline func(headline string) error,
) (*DashboardBriefResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, _, _, err := s.buildClientExt(ctx, taskKeySummaryGeneration)
	if err != nil {
		client, modelID, _, _, err = s.buildClientExt(ctx, taskKeyContentRewrite)
		if err != nil {
			return nil, fmt.Errorf("build ai client: %w", err)
		}
	}

	temp := 0.35
	var collected strings.Builder
	lastHeadline := ""

	streamErr := client.ChatStream(ctx, infraai.ChatRequest{
		Model:       modelID,
		Temperature: &temp,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: defaultDashboardBriefPrompt},
			{Role: "user", Content: snapshot},
		},
	}, func(chunk string) error {
		collected.WriteString(chunk)
		if onHeadline != nil {
			if partial, ok := extractPartialJSONString(collected.String(), "headline"); ok && len(partial) > len(lastHeadline) {
				lastHeadline = partial
				if err := onHeadline(partial); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if streamErr != nil {
		return nil, fmt.Errorf("ai chat stream: %w", streamErr)
	}

	return parseDashboardBriefResult(collected.String())
}

func parseDashboardBriefResult(rawContent string) (*DashboardBriefResult, error) {
	raw := extractJSON(rawContent)
	var parsed DashboardBriefResult
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("parse dashboard brief: %w", err)
	}

	parsed.Headline = strings.TrimSpace(parsed.Headline)
	if parsed.Headline == "" {
		return nil, fmt.Errorf("empty dashboard brief headline")
	}

	switch parsed.Mood {
	case "growing", "steady", "quiet", "attention":
	default:
		parsed.Mood = "steady"
	}

	allowedRoutes := map[string]struct{}{
		"/articles/list":             {},
		"/articles/edit/new":         {},
		"/notes/list":                {},
		"/notes/edit/new":            {},
		"/comments":                  {},
		"/friend-links/applications": {},
		"/audience/visitors":         {},
	}

	highlights := make([]DashboardBriefHighlight, 0, 3)
	for _, h := range parsed.Highlights {
		label := strings.TrimSpace(h.Label)
		detail := strings.TrimSpace(h.Detail)
		if label == "" || detail == "" {
			continue
		}
		highlights = append(highlights, DashboardBriefHighlight{Label: label, Detail: detail})
		if len(highlights) >= 3 {
			break
		}
	}
	parsed.Highlights = highlights

	actions := make([]DashboardBriefAction, 0, 3)
	for _, a := range parsed.Actions {
		title := strings.TrimSpace(a.Title)
		reason := strings.TrimSpace(a.Reason)
		if title == "" {
			continue
		}
		route := strings.TrimSpace(a.Route)
		if _, ok := allowedRoutes[route]; !ok {
			route = ""
		}
		actions = append(actions, DashboardBriefAction{
			Title:  title,
			Reason: reason,
			Route:  route,
		})
		if len(actions) >= 3 {
			break
		}
	}
	parsed.Actions = actions
	parsed.GeneratedAt = time.Now().UTC()
	parsed.Cached = false

	return &parsed, nil
}

// extractPartialJSONString 从流式累积的 JSON 文本中提取指定字段的部分字符串值，用于实时流式展示。
func extractPartialJSONString(raw, key string) (string, bool) {
	needle := `"` + key + `"`
	idx := strings.Index(raw, needle)
	if idx < 0 {
		return "", false
	}
	rest := raw[idx+len(needle):]
	colonIdx := strings.IndexByte(rest, ':')
	if colonIdx < 0 {
		return "", false
	}
	rest = strings.TrimLeft(rest[colonIdx+1:], " \t\r\n")
	if len(rest) == 0 || rest[0] != '"' {
		return "", false
	}
	rest = rest[1:]

	var sb strings.Builder
	escaped := false
	for i := 0; i < len(rest); i++ {
		ch := rest[i]
		if escaped {
			switch ch {
			case '"', '\\', '/':
				sb.WriteByte(ch)
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				// ignore
			default:
				sb.WriteByte(ch)
			}
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			return sb.String(), true
		}
		sb.WriteByte(ch)
	}
	return sb.String(), sb.Len() > 0
}

const defaultOpsReportPrompt = `你是一位资深、克制、懂独立中文博客的专栏主编与运营顾问，正在为站长撰写一份「完整版站点运营复盘报告」。
请严格基于提供的真实数据（站点趋势、分类标签、热门文章/手记、访客画像与漏斗、近期真实读者评论），给出有洞察、有温度、不编造数据、无官话套话的分析。
要求：
1. "summary"：2~3 句话的执行摘要（约 70~110 字），先讲核心态势，再点出最值得把握的一个机会或待改善点。
2. "mood"：只能是 "growing" | "steady" | "quiet" | "attention"。
3. "topics"（主题梳理）：
   - "overview"：1~2 句话概括博客的内容结构与更新节奏。
   - "pillars"：2~4 个当前内容支柱，每项含 "name"（主题名）、"strength"（"core"|"growing"|"quiet"）、"observation"（不超过 45 字的真实观察）。
   - "ideas"：2~3 个贴合站长现有风格的后续选题方向，每项含 "title"（具体可写的标题方向）、"angle"（切入角度，<=40字）、"why"（为什么值得写，<=35字）。
4. "audience"（访客画像）：
   - "overview"：1~2 句话描绘读者轮廓（地区/设备/回访特征）。
   - "highlights"：2~3 条简短画像观察（每条 <=45 字，结合真实数字）。
   - "retentionNote"：1 句话指出从「浏览 → 点赞 → 评论」漏斗中的关键机会或建议。
5. "comments"（评论区声音）：
   - "overview"：1~2 句话概括近期互动热度与讨论氛围（若近期评论较少，实事求是说明，绝不编造）。
   - "themes"：2~3 条读者常聊/关心的话题或高频反馈（每条 <=35 字）。
   - "suggestion"：1 句话给出自然提升互动的具体方式。
6. "operations"（整体运营建议）：
   - "overview"：1 句话总结下一阶段的运营重点。
   - "actions"：3~4 条分优先级的可执行行动项，每项含 "priority"（"high"|"medium"|"low"）、"title"（<=12字）、"detail"（<=55字）、"route"（只能从 "/articles/list", "/articles/edit/new", "/notes/list", "/notes/edit/new", "/comments", "/friend-links/applications", "/audience/visitors", "/taxonomy/tags", "" 中选）。
7. 严格输出合法 JSON，不要输出任何额外说明或 Markdown 包裹。`

// GenerateOpsReportStream 流式生成完整版深度运营复盘报告，实时回调 summary 增量文本。
func (s *Service) GenerateOpsReportStream(
	ctx context.Context,
	snapshot string,
	days int,
	onSummary func(summary string) error,
) (*OpsReportResult, error) {
	if err := s.checkEnabled(ctx); err != nil {
		return nil, err
	}

	client, modelID, _, _, err := s.buildClientExt(ctx, taskKeySummaryGeneration)
	if err != nil {
		client, modelID, _, _, err = s.buildClientExt(ctx, taskKeyContentRewrite)
		if err != nil {
			return nil, fmt.Errorf("build ai client: %w", err)
		}
	}

	temp := 0.4
	var collected strings.Builder
	lastSummary := ""

	streamErr := client.ChatStream(ctx, infraai.ChatRequest{
		Model:       modelID,
		Temperature: &temp,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: defaultOpsReportPrompt},
			{Role: "user", Content: snapshot},
		},
	}, func(chunk string) error {
		collected.WriteString(chunk)
		if onSummary != nil {
			if partial, ok := extractPartialJSONString(collected.String(), "summary"); ok && len(partial) > len(lastSummary) {
				lastSummary = partial
				if err := onSummary(partial); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if streamErr != nil {
		return nil, fmt.Errorf("ai chat stream: %w", streamErr)
	}

	return parseOpsReportResult(collected.String(), days)
}

func parseOpsReportResult(rawContent string, days int) (*OpsReportResult, error) {
	raw := extractJSON(rawContent)
	var parsed OpsReportResult
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("parse ops report: %w", err)
	}

	parsed.Summary = strings.TrimSpace(parsed.Summary)
	if parsed.Summary == "" {
		return nil, fmt.Errorf("empty ops report summary")
	}

	switch parsed.Mood {
	case "growing", "steady", "quiet", "attention":
	default:
		parsed.Mood = "steady"
	}

	allowedRoutes := map[string]struct{}{
		"/articles/list":             {},
		"/articles/edit/new":         {},
		"/notes/list":                {},
		"/notes/edit/new":            {},
		"/comments":                  {},
		"/friend-links/applications": {},
		"/audience/visitors":         {},
		"/taxonomy/tags":             {},
	}

	pillars := make([]OpsReportTopicPillar, 0, len(parsed.Topics.Pillars))
	for _, p := range parsed.Topics.Pillars {
		name := strings.TrimSpace(p.Name)
		obs := strings.TrimSpace(p.Observation)
		if name == "" || obs == "" {
			continue
		}
		strength := strings.TrimSpace(p.Strength)
		if strength != "core" && strength != "growing" && strength != "quiet" {
			strength = "core"
		}
		pillars = append(pillars, OpsReportTopicPillar{Name: name, Strength: strength, Observation: obs})
		if len(pillars) >= 4 {
			break
		}
	}
	parsed.Topics.Pillars = pillars

	ideas := make([]OpsReportTopicIdea, 0, len(parsed.Topics.Ideas))
	for _, idea := range parsed.Topics.Ideas {
		title := strings.TrimSpace(idea.Title)
		if title == "" {
			continue
		}
		ideas = append(ideas, OpsReportTopicIdea{
			Title: title,
			Angle: strings.TrimSpace(idea.Angle),
			Why:   strings.TrimSpace(idea.Why),
		})
		if len(ideas) >= 3 {
			break
		}
	}
	parsed.Topics.Ideas = ideas

	actions := make([]OpsReportActionItem, 0, len(parsed.Operations.Actions))
	for _, act := range parsed.Operations.Actions {
		title := strings.TrimSpace(act.Title)
		detail := strings.TrimSpace(act.Detail)
		if title == "" || detail == "" {
			continue
		}
		priority := strings.TrimSpace(act.Priority)
		if priority != "high" && priority != "medium" && priority != "low" {
			priority = "medium"
		}
		route := strings.TrimSpace(act.Route)
		if _, ok := allowedRoutes[route]; !ok {
			route = ""
		}
		actions = append(actions, OpsReportActionItem{
			Priority: priority,
			Title:    title,
			Detail:   detail,
			Route:    route,
		})
		if len(actions) >= 4 {
			break
		}
	}
	parsed.Operations.Actions = actions
	parsed.Days = days
	parsed.GeneratedAt = time.Now().UTC()
	parsed.Cached = false
	return &parsed, nil
}

// GenerateSummaryStream 使用 AI 流式生成文章摘要，通过 onChunk 回调增量返回。
func (s *Service) GenerateSummaryStream(ctx context.Context, content string, onChunk func(string) error) error {
	if err := s.checkEnabled(ctx); err != nil {
		return err
	}

	client, modelID, modelName, providerName, err := s.buildClientExt(ctx, taskKeySummaryGeneration)
	if err != nil {
		return fmt.Errorf("build ai client: %w", err)
	}

	taskLog, startTime := s.recordTaskStart(ctx, domainai.TaskTypeSummaryGeneration, content, domainai.TriggerManual, modelName, providerName)

	prompt := s.readPrompt(ctx, promptKeySummaryGeneration, defaultSummaryPrompt)

	var collected strings.Builder
	streamErr := client.ChatStream(ctx, infraai.ChatRequest{
		Model: modelID,
		Messages: []infraai.ChatMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	}, func(chunk string) error {
		collected.WriteString(chunk)
		return onChunk(chunk)
	})

	s.recordTaskEnd(ctx, taskLog, collected.String(), streamErr, startTime)
	return streamErr
}

// ── TaskLog 查询代理 ──

func (s *Service) ListTaskLogs(ctx context.Context, opts domainai.TaskLogListOptions) ([]*domainai.TaskLog, int64, error) {
	return s.repo.ListTaskLogs(ctx, opts)
}

func (s *Service) GetTaskLogByID(ctx context.Context, id int64) (*domainai.TaskLog, error) {
	return s.repo.GetTaskLogByID(ctx, id)
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

// buildClientExt extends buildClient to also return model name and provider name.
func (s *Service) buildClientExt(ctx context.Context, taskConfigKey string) (infraai.Client, string, string, string, error) {
	modelIDStr, err := s.cfgGet.GetConfigValue(ctx, taskConfigKey)
	if err != nil || strings.TrimSpace(modelIDStr) == "" {
		return nil, "", "", "", fmt.Errorf("该任务未配置 AI 模型 (key: %s)", taskConfigKey)
	}

	modelID, err := strconv.ParseInt(strings.TrimSpace(modelIDStr), 10, 64)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("无效的模型 ID: %s", modelIDStr)
	}

	m, p, err := s.repo.GetModelWithProvider(ctx, modelID)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("获取模型信息失败: %w", err)
	}

	if !p.IsActive {
		return nil, "", "", "", fmt.Errorf("AI 提供商 %q 已禁用", p.Name)
	}
	if !m.IsActive {
		return nil, "", "", "", fmt.Errorf("AI 模型 %q 已禁用", m.Name)
	}

	client, err := infraai.NewClient(p.Type, p.APIURL, p.APIKey)
	if err != nil {
		return nil, "", "", "", err
	}

	return client, m.ModelID, m.Name, p.Name, nil
}

func (s *Service) recordTaskStart(ctx context.Context, taskType, input, triggerSource, modelName, providerName string) (*domainai.TaskLog, time.Time) {
	startTime := time.Now()
	taskLog := &domainai.TaskLog{
		TaskType:      taskType,
		ModelName:     modelName,
		ProviderName:  providerName,
		Status:        domainai.TaskStatusRunning,
		InputText:     truncateText(input, 10000),
		TriggerSource: triggerSource,
	}
	if err := s.repo.CreateTaskLog(ctx, taskLog); err != nil {
		log.Printf("[AI] create task log failed: %v", err)
	}
	return taskLog, startTime
}

func (s *Service) recordTaskEnd(ctx context.Context, taskLog *domainai.TaskLog, output string, taskErr error, startTime time.Time) {
	if taskLog == nil || taskLog.ID == 0 {
		return
	}
	taskLog.DurationMs = int(time.Since(startTime).Milliseconds())
	taskLog.OutputText = truncateText(output, 10000)
	if taskErr != nil {
		if isTaskInterruptedError(taskErr) {
			taskLog.Status = domainai.TaskStatusInterrupted
		} else {
			taskLog.Status = domainai.TaskStatusFailed
		}
		taskLog.ErrorMessage = taskErr.Error()
	} else {
		taskLog.Status = domainai.TaskStatusCompleted
	}

	// task 结束落库不要绑定调用链 context，避免请求取消或超时后状态卡在 running。
	updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.repo.UpdateTaskLog(updateCtx, taskLog); err != nil {
		log.Printf("[AI] update task log failed: %v", err)
	}
}

func isTaskInterruptedError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		return true
	}

	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset by peer") ||
		strings.Contains(msg, "context canceled") ||
		strings.Contains(msg, "client disconnected")
}

func (s *Service) readPrompt(ctx context.Context, promptKey, defaultPrompt string) string {
	val, err := s.cfgGet.GetConfigValue(ctx, promptKey)
	if err != nil || strings.TrimSpace(val) == "" {
		return defaultPrompt
	}
	return val
}

// extractJSON 从 AI 回复中提取 JSON 内容（处理推理标签、markdown code block 与前后说明文字）。
func extractJSON(raw string) string {
	trimmed := strings.TrimSpace(raw)
	// 移除部分推理模型可能带出的 <think>...</think> 块
	for {
		start := strings.Index(trimmed, "<think>")
		end := strings.Index(trimmed, "</think>")
		if start >= 0 && end > start {
			trimmed = strings.TrimSpace(trimmed[:start] + trimmed[end+len("</think>"):])
		} else {
			break
		}
	}
	// 移除 markdown code block 包裹
	if idx := strings.Index(trimmed, "```json"); idx >= 0 {
		sub := trimmed[idx+len("```json"):]
		if end := strings.Index(sub, "```"); end >= 0 {
			trimmed = strings.TrimSpace(sub[:end])
		} else {
			trimmed = strings.TrimSpace(sub)
		}
	} else if idx := strings.Index(trimmed, "```"); idx >= 0 {
		sub := trimmed[idx+len("```"):]
		if end := strings.Index(sub, "```"); end >= 0 {
			trimmed = strings.TrimSpace(sub[:end])
		} else {
			trimmed = strings.TrimSpace(sub)
		}
	}
	// 如果仍有前置或尾随说明文字，截取最外层 JSON 对象或数组
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		objStart := strings.IndexByte(trimmed, '{')
		arrStart := strings.IndexByte(trimmed, '[')
		if objStart >= 0 && (arrStart < 0 || objStart < arrStart) {
			if objEnd := strings.LastIndexByte(trimmed, '}'); objEnd > objStart {
				return strings.TrimSpace(trimmed[objStart : objEnd+1])
			}
		} else if arrStart >= 0 {
			if arrEnd := strings.LastIndexByte(trimmed, ']'); arrEnd > arrStart {
				return strings.TrimSpace(trimmed[arrStart : arrEnd+1])
			}
		}
	}
	return strings.TrimSpace(trimmed)
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
