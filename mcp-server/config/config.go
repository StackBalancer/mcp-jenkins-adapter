package config

import (
	"fmt"
	"os"
)

type Config struct {
	Jenkins struct {
		URL       string
		User      string
		Token     string
		TokenFile string
	}
	LLM struct {
		Provider  string
		OpenAIKey string
		ClaudeKey string
		Model     string
	}
	MCP struct {
		ServerURL string
		Port      string
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Jenkins configuration
	cfg.Jenkins.URL = getEnvOrDefault("JENKINS_URL", "http://jenkins:8080/jenkins")
	cfg.Jenkins.User = getEnvOrDefault("JENKINS_MCP_USER", "mcp-user")
	cfg.Jenkins.TokenFile = getEnvOrDefault("JENKINS_TOKEN_FILE", "/run/secrets/mcp-user.token")

	// LLM configuration
	cfg.LLM.Provider = getEnvOrDefault("LLM_PROVIDER", "openai")
	cfg.LLM.OpenAIKey = os.Getenv("OPENAI_API_KEY")
	cfg.LLM.ClaudeKey = os.Getenv("ANTHROPIC_API_KEY")
	cfg.LLM.Model = os.Getenv("LLM_MODEL")

	// MCP configuration
	cfg.MCP.ServerURL = getEnvOrDefault("MCP_SERVER_URL", "http://localhost:8081/sse")
	cfg.MCP.Port = getEnvOrDefault("MCP_PORT", "8081")

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	switch c.LLM.Provider {
	case "openai":
		if c.LLM.OpenAIKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is required when using OpenAI provider")
		}
	case "claude":
		if c.LLM.ClaudeKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY is required when using Claude provider")
		}
	default:
		return fmt.Errorf("unsupported LLM provider: %s", c.LLM.Provider)
	}

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
