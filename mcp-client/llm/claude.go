package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	apiKey string
	model  string
	client *anthropic.Client
}

func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)
	return &ClaudeProvider{
		apiKey: apiKey,
		model:  model,
		client: &client,
	}
}

func (p *ClaudeProvider) CreateCompletion(ctx context.Context, messages []Message, maxTokens int) (string, error) {
	if maxTokens == 0 {
		maxTokens = 1000
	}

	// extract system message and build anthropic messages
	var systemPrompt string
	anthropicMessages := make([]anthropic.MessageParam, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		default:
			anthropicMessages = append(anthropicMessages, anthropic.MessageParam{
				Role: anthropic.MessageParamRole(msg.Role),
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewTextBlock(msg.Content),
				},
			})
		}
	}

	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: int64(maxTokens),
		Messages:  anthropicMessages,
	}

	// only set system if one was found
	if systemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	resp, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}

	return resp.Content[0].Text, nil
}

func (p *ClaudeProvider) GetName() string {
	return fmt.Sprintf("Claude (%s)", p.model)
}
