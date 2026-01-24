package llm

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
	}
}

func (p *OpenAIProvider) CreateCompletion(ctx context.Context, messages []Message, maxTokens int) (string, error) {
	oaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		oaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:       "gpt-4",
		Messages:    oaiMessages,
		Temperature: 0.2,
	}
	if maxTokens > 0 {
		req.MaxTokens = maxTokens
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) GetName() string {
	return "OpenAI"
}