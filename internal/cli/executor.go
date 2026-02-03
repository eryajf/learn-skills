package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// parseMCPCommand parses an MCP tool call command
// Returns: (tool name, arguments map, is MCP call)
func parseMCPCommand(command string) (string, map[string]interface{}, bool) {
	// Check if this is an MCP call command
	if !strings.Contains(command, "cnb-mcp.py call") {
		return "", nil, false
	}

	// Remove newlines and extra spaces
	command = strings.ReplaceAll(command, "\n", " ")
	command = strings.ReplaceAll(command, "\\", "")
	command = regexp.MustCompile(`\s+`).ReplaceAllString(command, " ")
	command = strings.TrimSpace(command)

	// Find "call" keyword and extract everything after it
	callIdx := strings.Index(command, "call ")
	if callIdx == -1 {
		return "", nil, false
	}

	// Get the part after "call "
	afterCall := strings.TrimSpace(command[callIdx+5:])
	if afterCall == "" {
		return "", nil, false
	}

	// Split into tokens, respecting quotes
	tokens := parseTokens(afterCall)
	if len(tokens) == 0 {
		return "", nil, false
	}

	// First token is the tool name
	toolName := tokens[0]
	args := make(map[string]interface{})

	// Parse remaining tokens as key=value pairs
	for _, token := range tokens[1:] {
		if strings.Contains(token, "=") {
			parts := strings.SplitN(token, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]

				// Try to parse value as JSON
				var jsonValue interface{}
				if err := json.Unmarshal([]byte(value), &jsonValue); err == nil {
					args[key] = jsonValue
				} else {
					// If JSON parsing fails, use as string
					args[key] = value
				}
			}
		}
	}

	return toolName, args, true
}

// parseTokens splits a string into tokens, respecting quoted strings
func parseTokens(s string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for i, r := range s {
		switch r {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteRune(r)
		case ' ':
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}

		// Add last token if we're at the end
		if i == len(s)-1 && current.Len() > 0 {
			tokens = append(tokens, current.String())
		}
	}

	return tokens
}

// formatMCPCallStart formats the output at the start of a tool call
func formatMCPCallStart(toolName string, args map[string]interface{}) string {
	var sb strings.Builder

	sb.WriteString("\n📡 正在调用 MCP 工具：")
	sb.WriteString(toolName)
	sb.WriteString("\n   参数：")

	// Format arguments as JSON
	if len(args) == 0 {
		sb.WriteString("{}")
	} else {
		argsJSON, err := json.MarshalIndent(args, "   ", "  ")
		if err != nil {
			sb.WriteString(fmt.Sprintf("(无法格式化: %v)", args))
		} else {
			sb.WriteString(string(argsJSON))
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

// formatMCPCallEnd formats the output at the end of a tool call
func formatMCPCallEnd(info MCPToolInfo) string {
	var sb strings.Builder

	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("ℹ️  数据来源：")
	sb.WriteString(info.ToolName)
	sb.WriteString("\n   参数：")

	// Format arguments as JSON
	if len(info.Arguments) == 0 {
		sb.WriteString("{}")
	} else {
		argsJSON, err := json.MarshalIndent(info.Arguments, "   ", "  ")
		if err != nil {
			sb.WriteString(fmt.Sprintf("(无法格式化: %v)", info.Arguments))
		} else {
			sb.WriteString(string(argsJSON))
		}
	}

	// Add duration
	duration := info.Duration()
	sb.WriteString(fmt.Sprintf("\n   耗时：%.2fs", duration.Seconds()))
	sb.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return sb.String()
}

// ExecuteTool executes a tool call and returns the result
func (a *Assistant) ExecuteTool(toolName string, argumentsJSON string) (string, error) {
	switch toolName {
	case "execute_bash":
		var args struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(argumentsJSON), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}

		// Parse MCP command
		mcpToolName, mcpArgs, isMCP := parseMCPCommand(args.Command)

		if isMCP {
			// Output call start information
			fmt.Print(formatMCPCallStart(mcpToolName, mcpArgs))

			// Record start time and execute
			startTime := time.Now()
			result, err := executeBashCommand(args.Command)

			// Store call end information to be printed later (after LLM response)
			info := MCPToolInfo{
				ToolName:  mcpToolName,
				Arguments: mcpArgs,
				StartTime: startTime,
				EndTime:   time.Now(),
			}
			a.pendingMCPCallEnding = append(a.pendingMCPCallEnding, info)

			return result, err
		}

		// Non-MCP command, execute normally
		return executeBashCommand(args.Command)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// executeBashCommand runs a bash command and returns stdout
func executeBashCommand(command string) (string, error) {
	// Use bash -c to execute the command
	cmd := exec.Command("bash", "-c", command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w\nOutput: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}
