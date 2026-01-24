package llm

import (
	"context"
	"fmt"
	"os"
)

type Provider interface {
	CreateCompletion(ctx context.Context, messages []Message, maxTokens int) (string, error)
	GetName() string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewProvider(providerType string) (Provider, error) {
	switch providerType {
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
		}
		return NewOpenAIProvider(apiKey), nil
	case "claude":
		apiKey := os.Getenv("CLAUDE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("CLAUDE_API_KEY environment variable not set")
		}
		return NewClaudeProvider(apiKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}