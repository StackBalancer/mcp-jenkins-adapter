package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ClaudeProvider struct {
	apiKey string
	client *http.Client
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (p *ClaudeProvider) CreateCompletion(ctx context.Context, messages []Message, maxTokens int) (string, error) {
	if maxTokens == 0 {
		maxTokens = 1000
	}

	claudeMessages := make([]claudeMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Role == "system" {
			continue // Claude handles system messages differently
		}
		claudeMessages = append(claudeMessages, claudeMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	reqBody := claudeRequest{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: maxTokens,
		Messages:  claudeMessages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", err
	}

	if claudeResp.Error != nil {
		return "", fmt.Errorf("claude API error: %s", claudeResp.Error.Message)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in Claude response")
	}

	return claudeResp.Content[0].Text, nil
}

func (p *ClaudeProvider) GetName() string {
	return "Claude"
}