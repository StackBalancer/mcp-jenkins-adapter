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

type ProviderConfig struct {
	APIKey string
	Model  string
}

func NewProvider(providerType string, cfg ProviderConfig) (Provider, error) {
	switch providerType {
	case "claude":
		if cfg.APIKey == "" {
			cfg.APIKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY not set")
		}
		model := ResolveModel(cfg.Model, DefaultClaudeModel)
		return NewClaudeProvider(cfg.APIKey, model), nil

	case "openai":
		if cfg.APIKey == "" {
			cfg.APIKey = os.Getenv("OPENAI_API_KEY")
		}
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY not set")
		}
		model := ResolveModel(cfg.Model, DefaultOpenAIModel)
		return NewOpenAIProvider(cfg.APIKey, model), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerType)
	}
}
