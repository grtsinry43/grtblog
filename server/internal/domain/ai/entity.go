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
