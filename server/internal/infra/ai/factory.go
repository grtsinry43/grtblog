package ai

import "fmt"

// NewClient 根据提供商类型创建对应的 AI 客户端。
func NewClient(providerType, apiURL, apiKey string) (Client, error) {
	switch providerType {
	case "openai":
		return NewOpenAIClient(apiURL, apiKey), nil
	case "openrouter":
		return NewOpenRouterClient(apiURL, apiKey), nil
	case "gemini":
		return NewGeminiClient(apiURL, apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported ai provider type: %s", providerType)
	}
}
