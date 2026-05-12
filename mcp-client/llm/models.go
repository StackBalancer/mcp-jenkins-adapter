package llm

const (
	DefaultClaudeModel = "claude-sonnet-4-20250514"
	DefaultOpenAIModel = "gpt-4o"
)

// ModelAliases lets users pass short names instead of full model strings
var ModelAliases = map[string]string{
	// Claude
	"claude-sonnet": "claude-sonnet-4-20250514",
	"claude-haiku":  "claude-haiku-4-5-20251001",
	"claude-opus":   "claude-opus-4-20250514",
	// OpenAI
	"gpt-4o":      "gpt-4o",
	"gpt-4o-mini": "gpt-4o-mini",
}

func ResolveModel(input, providerDefault string) string {
	if input == "" {
		return providerDefault
	}
	if resolved, ok := ModelAliases[input]; ok {
		return resolved
	}
	return input // trust the user if it's not a known alias
}
