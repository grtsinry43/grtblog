package contract

import "time"

// ── Provider ──

type CreateAIProviderReq struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // openai / openrouter / gemini
	APIURL   string `json:"apiUrl"`
	APIKey   string `json:"apiKey"`
	IsActive *bool  `json:"isActive,omitempty"`
}

type UpdateAIProviderReq struct {
	Name     *string `json:"name,omitempty"`
	Type     *string `json:"type,omitempty"`
	APIURL   *string `json:"apiUrl,omitempty"`
	APIKey   *string `json:"apiKey,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
}

type AIProviderResp struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	APIURL    string    `json:"apiUrl"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ── Model ──

type CreateAIModelReq struct {
	ProviderID int64  `json:"providerId"`
	Name       string `json:"name"`
	ModelID    string `json:"modelId"`
	IsActive   *bool  `json:"isActive,omitempty"`
}

type UpdateAIModelReq struct {
	ProviderID *int64  `json:"providerId,omitempty"`
	Name       *string `json:"name,omitempty"`
	ModelID    *string `json:"modelId,omitempty"`
	IsActive   *bool   `json:"isActive,omitempty"`
}

type AIModelResp struct {
	ID           int64     `json:"id"`
	ProviderID   int64     `json:"providerId"`
	ProviderName string    `json:"providerName,omitempty"`
	ProviderType string    `json:"providerType,omitempty"`
	Name         string    `json:"name"`
	ModelID      string    `json:"modelId"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ── AI 功能 ──

type AIModerateCommentReq struct {
	Content string `json:"content"`
}

type AIModerateCommentResp struct {
	Approved bool    `json:"approved"`
	Reason   string  `json:"reason"`
	Score    float64 `json:"score"`
}

type AIGenerateTitleReq struct {
	Content string `json:"content"`
}

type AIGenerateTitleResp struct {
	Title    string `json:"title"`
	ShortURL string `json:"shortUrl"`
}

type AIRewriteContentReq struct {
	Content     string `json:"content"`
	Instruction string `json:"instruction"`
}

type AIRewriteContentResp struct {
	Content string `json:"content"`
}

type AIGenerateSummaryReq struct {
	Content string `json:"content"`
}

type AIGenerateSummaryResp struct {
	Summary string `json:"summary"`
}

type AICompleteReq struct {
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix,omitempty"`
	Title  string `json:"title,omitempty"`
}

type AICheckWritingReq struct {
	Content string `json:"content"`
}

type AIWritingIssue struct {
	Original   string `json:"original"`
	Suggestion string `json:"suggestion"`
	Reason     string `json:"reason"`
	Severity   string `json:"severity"`
}

type AICheckWritingResp struct {
	Issues []AIWritingIssue `json:"issues"`
}

type AIAnnotateParagraphReq struct {
	Paragraph   string `json:"paragraph"`
	PrevContext string `json:"prevContext,omitempty"`
	Title       string `json:"title,omitempty"`
}

type AIMarginAnnotation struct {
	Kind       string `json:"kind"`
	Quote      string `json:"quote,omitempty"`
	Comment    string `json:"comment"`
	Suggestion string `json:"suggestion,omitempty"`
}

type AIAnnotateParagraphResp struct {
	Annotations []AIMarginAnnotation `json:"annotations"`
}

type AIDashboardBriefHighlight struct {
	Label  string `json:"label"`
	Detail string `json:"detail"`
}

type AIDashboardBriefAction struct {
	Title  string `json:"title"`
	Reason string `json:"reason"`
	Route  string `json:"route,omitempty"`
}

type AIDashboardBriefResp struct {
	Enabled     bool                        `json:"enabled"`
	Headline    string                      `json:"headline,omitempty"`
	Mood        string                      `json:"mood,omitempty"`
	Highlights  []AIDashboardBriefHighlight `json:"highlights,omitempty"`
	Actions     []AIDashboardBriefAction    `json:"actions,omitempty"`
	GeneratedAt *time.Time                  `json:"generatedAt,omitempty"`
	Cached      bool                        `json:"cached"`
}

type AIOpsReportTopicPillar struct {
	Name        string `json:"name"`
	Strength    string `json:"strength"`
	Observation string `json:"observation"`
}

type AIOpsReportTopicIdea struct {
	Title string `json:"title"`
	Angle string `json:"angle"`
	Why   string `json:"why"`
}

type AIOpsReportTopicsSection struct {
	Overview string                   `json:"overview"`
	Pillars  []AIOpsReportTopicPillar `json:"pillars"`
	Ideas    []AIOpsReportTopicIdea   `json:"ideas"`
}

type AIOpsReportAudienceSection struct {
	Overview      string   `json:"overview"`
	Highlights    []string `json:"highlights"`
	RetentionNote string   `json:"retentionNote"`
}

type AIOpsReportCommentsSection struct {
	Overview   string   `json:"overview"`
	Themes     []string `json:"themes"`
	Suggestion string   `json:"suggestion"`
}

type AIOpsReportActionItem struct {
	Priority string `json:"priority"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Route    string `json:"route,omitempty"`
}

type AIOpsReportOperationsSection struct {
	Overview string                  `json:"overview"`
	Actions  []AIOpsReportActionItem `json:"actions"`
}

type AIOpsReportResp struct {
	Enabled     bool                          `json:"enabled"`
	Summary     string                        `json:"summary,omitempty"`
	Mood        string                        `json:"mood,omitempty"`
	Days        int                           `json:"days,omitempty"`
	Topics      *AIOpsReportTopicsSection     `json:"topics,omitempty"`
	Audience    *AIOpsReportAudienceSection   `json:"audience,omitempty"`
	Comments    *AIOpsReportCommentsSection   `json:"comments,omitempty"`
	Operations  *AIOpsReportOperationsSection `json:"operations,omitempty"`
	GeneratedAt *time.Time                    `json:"generatedAt,omitempty"`
	Cached      bool                          `json:"cached"`
}

type AIOpsReportIdeasResp struct {
	Ideas       []AIOpsReportTopicIdea `json:"ideas"`
	GeneratedAt *time.Time             `json:"generatedAt,omitempty"`
}

type AIFeatureFlagsResp struct {
	Enabled         bool `json:"enabled"`
	RealtimeAssist  bool `json:"realtimeAssist"`
	GhostCompletion bool `json:"ghostCompletion"`
	MarginNotes     bool `json:"marginNotes"`
	InlineLint      bool `json:"inlineLint"`
	DashboardBrief  bool `json:"dashboardBrief"`
	OpsReport       bool `json:"opsReport"`
	OnboardingSeen  bool `json:"onboardingSeen"`
}

type UpdateAIFeatureFlagsReq struct {
	RealtimeAssist  *bool `json:"realtimeAssist,omitempty"`
	GhostCompletion *bool `json:"ghostCompletion,omitempty"`
	MarginNotes     *bool `json:"marginNotes,omitempty"`
	InlineLint      *bool `json:"inlineLint,omitempty"`
	DashboardBrief  *bool `json:"dashboardBrief,omitempty"`
	OpsReport       *bool `json:"opsReport,omitempty"`
	OnboardingSeen  *bool `json:"onboardingSeen,omitempty"`
}

// ── TaskLog ──

type AITaskLogResp struct {
	ID            int64     `json:"id"`
	TaskType      string    `json:"taskType"`
	ModelName     string    `json:"modelName"`
	ProviderName  string    `json:"providerName"`
	Status        string    `json:"status"`
	InputText     string    `json:"inputText,omitempty"`
	OutputText    string    `json:"outputText,omitempty"`
	ErrorMessage  string    `json:"errorMessage,omitempty"`
	DurationMs    int       `json:"durationMs"`
	TriggerSource string    `json:"triggerSource"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type AITaskLogListResp struct {
	Items []AITaskLogResp `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Size  int             `json:"size"`
}
