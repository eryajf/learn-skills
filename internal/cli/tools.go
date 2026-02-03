package cli

import (
	"cnb.cool/znb/learn-skills/internal/llm"
)

// GetCNBTools returns the CNB tool definitions for the LLM
func GetCNBTools() []llm.Tool {
	return []llm.Tool{
		{
			Type: "function",
			Function: llm.Function{
				Name:        "execute_bash",
				Description: "Execute a bash command and return the output. Use this to call the CNB MCP Python script (skills/scripts/cnb-mcp.py).",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"command": map[string]interface{}{
							"type":        "string",
							"description": "The bash command to execute. For CNB operations, use: python3 skills/scripts/cnb-mcp.py <operation> [args]",
						},
					},
					"required": []string{"command"},
				},
			},
		},
	}
}
