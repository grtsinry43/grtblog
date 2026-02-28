package ai

import "context"

// ChatMessage 表示一条对话消息。
type ChatMessage struct {
	Role    string `json:"role"`    // "system" / "user" / "assistant"
	Content string `json:"content"`
}

// ChatRequest 是统一的 AI 聊天请求。
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

// ChatResponse 是统一的 AI 聊天响应。
type ChatResponse struct {
	Content string `json:"content"`
	Model   string `json:"model"`
}

// Client 是 AI 聊天补全的统一接口，所有提供商都需实现。
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req ChatRequest, onChunk func(chunk string) error) error
}
