package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"jenkins/mcp-client/llm"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get LLM provider from environment (default: openai)
	providerType := os.Getenv("LLM_PROVIDER")
	if providerType == "" {
		providerType = "claude"
	}

	llmCfg := llm.ProviderConfig{
		Model: os.Getenv("LLM_MODEL"),
	}
	llmProvider, err := llm.NewProvider(providerType, llmCfg)
	if err != nil {
		log.Fatalf("Failed to initialize LLM provider: %v", err)
	}

	// MCP server SSE URL
	mcpURL := "http://localhost:8081/sse"
	if u := os.Getenv("MCP_SERVER_URL"); u != "" {
		mcpURL = u
	}

	// Connect to MCP server
	mcpClient, err := client.NewSSEMCPClient(mcpURL)
	if err != nil {
		log.Fatalf("failed to connect to MCP server: %v", err)
	}
	defer mcpClient.Close()

	// Start client connection before Initialize
	ready := make(chan error, 1)

	go func() {
		// Start blocks until error or close
		ready <- mcpClient.Start(ctx)
	}()

	// Wait for transport to be ready
	if err := <-ready; err != nil {
		log.Fatalf("mcp connection start failed: %v", err)
	}

	// Initialize MCP session
	initResp, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			ClientInfo: mcp.Implementation{
				Name:    "jenkins-llm-bridge",
				Version: "0.1",
			},
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize MCP client: %v", err)
	}

	fmt.Printf("MCP initialized. Server: %s v%s\n", initResp.ServerInfo.Name, initResp.ServerInfo.Version)
	fmt.Printf("Using LLM provider: %s\n", llmProvider.GetName())

	// Conversation history
	history := []llm.Message{
		{
			Role: "system",
			Content: `You are a DevOps assistant.
				- "run", "start", "trigger job" → TOOL: trigger_job {"job_name": "<name>"}
				- "status", "check build" → TOOL: get_build_status {"job_name": "<name>", "build_number": <number>}
				- "logs", "console output" → TOOL: get_console_log {"job_name": "<name>", "build_number": <number>}
				- "troubleshoot", "analyze logs", "debug", "why did it fail" → TOOL: analyze_logs {"job_name": "<name>", "build_number": <number>}

				Never answer in natural language for these cases.`,
		},
	}

	fmt.Println("Jenkins LLM Bridge started. Type your prompts:")

	// Limit history size
	const maxHistory = 50

	// Allowed tools
	var allowedTools = map[string]bool{
		"trigger_job":      true,
		"get_build_status": true,
		"get_console_log":  true,
		"analyze_logs":     true,
	}

	// REPL loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		var input string
		if !scanner.Scan() {
			break
		}
		input = scanner.Text()

		// Append user message
		history = append(history, llm.Message{
			Role:    "user",
			Content: input,
		})

		if len(history) > maxHistory {
			// Keep system message + recent history
			systemMsg := history[0]
			history = append([]llm.Message{systemMsg}, history[len(history)-maxHistory+1:]...)
		}

		// Send prompt to LLM
		llmReply, err := llmProvider.CreateCompletion(ctx, history, 0)
		if err != nil {
			log.Printf("LLM error: %v", err)
			continue
		}
		fmt.Printf("LLM: %s\n", llmReply)

		// Check if LLM wants to call a tool
		if toolCall := parseToolCall(llmReply); toolCall != nil {
			fmt.Printf("→ Detected MCP tool call: %+v\n", toolCall)

			// Validate tool call
			if !allowedTools[toolCall.Name] {
				log.Printf("Unauthorized tool: %s", toolCall.Name)
				continue
			}

			if toolCall.Name == "analyze_logs" {
				// Special call: fetch logs first, then analyze with OpenAI
				jobName, ok := toolCall.Params["job_name"].(string)
				if !ok || !isValidJobName(jobName) {
					log.Printf("Invalid job_name format: %s", jobName)
					continue
				}
				buildNumFloat, ok := toolCall.Params["build_number"].(float64)
				if !ok || buildNumFloat < 0 {
					log.Printf("Invalid build_number")
					continue
				}

				// Step 1: get logs from MCP
				logReq := mcp.CallToolRequest{
					Params: mcp.CallToolParams{
						Name: "get_console_log",
						Arguments: map[string]any{
							"job_name":     jobName,
							"build_number": buildNumFloat,
						},
					},
				}
				logResp, err := mcpClient.CallTool(ctx, logReq)
				if err != nil {
					fmt.Printf("MCP log fetch error: %v\n", err)
					continue
				}

				logContent := logResp.Content
				// Convert MCP tool output (array of Content) into plain text
				toolText := extractTextFromContent(logContent)
				fmt.Println("→ Jenkins logs fetched, sending to OpenAI...")

				// Step 2: send logs to LLM for troubleshooting
				analysisMessages := []llm.Message{
					{Role: "system", Content: "You are a DevOps expert. Analyze Jenkins logs and explain errors, causes, and fixes."},
					{Role: "user", Content: toolText},
				}
				result, err := llmProvider.CreateCompletion(ctx, analysisMessages, 500)
				if err != nil {
					fmt.Printf("LLM log analysis error: %v\n", err)
					continue
				}
				fmt.Printf("Analysis: %s\n", result)

				// Add back into history
				history = append(history, llm.Message{
					Role:    "assistant",
					Content: "[Log analysis completed]",
				})
			} else {
				// Validate parameters for other tools
				jobName, ok := toolCall.Params["job_name"].(string)
				if !ok || jobName == "" {
					log.Printf("Invalid or missing job_name")
					continue
				}

				// Validate job_name pattern (alphanumeric, dash, underscore only)
				if !isValidJobName(jobName) {
					log.Printf("Invalid job_name format: %s", jobName)
					continue
				}

				// For tools that need build_number
				if toolCall.Name == "get_build_status" || toolCall.Name == "get_console_log" {
					buildNumFloat, ok := toolCall.Params["build_number"].(float64)
					if !ok || buildNumFloat < 0 {
						log.Printf("Invalid build_number")
						continue
					}
				}
				// Normal MCP tool call
				req := mcp.CallToolRequest{
					Params: mcp.CallToolParams{
						Name:      toolCall.Name,
						Arguments: toolCall.Params,
					},
				}
				_, err := mcpClient.CallTool(ctx, req)
				if err != nil {
					fmt.Printf("MCP call error: %v\n", err)
					continue
				}
				log.Printf("→ Tool '%s' completed successfully\n", toolCall.Name)

				history = append(history, llm.Message{
					Role:    "assistant",
					Content: "[Tool executed successfully]",
				})
			}
		} else {
			history = append(history, llm.Message{Role: "assistant", Content: llmReply})
		}
	}
}

// Validate job name
func isValidJobName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}
	// Allow alphanumeric, dash, underscore, and dot
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '_' || ch == '.') {
			return false
		}
	}
	return true
}

// extractTextFromContent flattens []mcp.Content into a readable string
func extractTextFromContent(contents []mcp.Content) string {
	var sb strings.Builder
	for _, c := range contents {
		switch v := c.(type) {
		case *mcp.TextContent:
			sb.WriteString(v.Text)
			sb.WriteString("\n")
		default:
			// fallback: dump raw JSON if it's not text
			b, _ := json.Marshal(v)
			sb.Write(b)
			sb.WriteString("\n")
		}
	}
	return strings.TrimSpace(sb.String())
}

// ToolCall struct
type ToolCall struct {
	Name   string
	Params map[string]any
}

// parseToolCall parses LLM output for a simple "call tool" syntax
// Example expected format: TOOL: trigger_job {"job_name":"demo-job"}
func parseToolCall(reply string) *ToolCall {
	reply = strings.TrimSpace(reply)
	if !strings.HasPrefix(reply, "TOOL:") {
		return nil
	}

	// Split into tool name and the rest
	raw := strings.TrimSpace(reply[len("TOOL:"):])
	parts := strings.SplitN(raw, " ", 2)
	if len(parts) < 2 {
		return nil
	}

	name := strings.TrimSpace(parts[0])
	rawParams := strings.TrimSpace(parts[1])

	// Extract only the {...} part for JSON safety
	start := strings.Index(rawParams, "{")
	end := strings.LastIndex(rawParams, "}")
	if start == -1 || end == -1 || end <= start {
		log.Printf("invalid tool params, no JSON object found: %s", rawParams)
		return nil
	}
	jsonStr := rawParams[start : end+1]

	// Fix common LLM issues: True/False/None → true/false/null
	jsonStr = strings.ReplaceAll(jsonStr, "True", "true")
	jsonStr = strings.ReplaceAll(jsonStr, "False", "false")
	jsonStr = strings.ReplaceAll(jsonStr, "None", "null")

	// Parse JSON
	var params map[string]any
	err := json.Unmarshal([]byte(jsonStr), &params)
	if err != nil {
		log.Printf("failed to parse tool params JSON")
		return nil
	}

	return &ToolCall{Name: name, Params: params}
}
