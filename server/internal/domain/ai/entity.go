package ai

import (
	"errors"
	"time"
)

var (
	ErrProviderNotFound = errors.New("ai provider not found")
	ErrModelNotFound    = errors.New("ai model not found")
)

type Provider struct {
	ID        int64
	Name      string
	Type      string // "openai" / "openrouter" / "gemini"
	APIURL    string
	APIKey    string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Model struct {
	ID         int64
	ProviderID int64
	Name       string
	ModelID    string // 实际 API model identifier，如 "gpt-4o"
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Task types
const (
	TaskTypeCommentModeration = "comment_moderation"
	TaskTypeTitleGeneration   = "title_generation"
	TaskTypeContentRewrite    = "content_rewrite"
	TaskTypeSummaryGeneration = "summary_generation"
)

// Task statuses
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// Trigger sources
const (
	TriggerManual = "manual"
	TriggerAuto   = "auto"
)

type TaskLog struct {
	ID            int64
	TaskType      string
	ModelName     string
	ProviderName  string
	Status        string
	InputText     string
	OutputText    string
	ErrorMessage  string
	DurationMs    int
	TriggerSource string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TaskLogListOptions struct {
	Page     int
	PageSize int
	TaskType *string
	Status   *string
	Search   *string
}
