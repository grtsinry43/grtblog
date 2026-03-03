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
