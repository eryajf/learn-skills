package cli

import (
	"time"

	"cnb.cool/znb/learn-skills/internal/llm"
)

// MCPToolInfo stores information about an MCP tool call
type MCPToolInfo struct {
	ToolName  string                 // Tool name, e.g. "list_organizations"
	Arguments map[string]interface{} // Parameter key-value pairs
	StartTime time.Time              // Start time
	EndTime   time.Time              // End time
}

// Duration returns the execution duration
func (m *MCPToolInfo) Duration() time.Duration {
	return m.EndTime.Sub(m.StartTime)
}

// Assistant holds the core components
type Assistant struct {
	LLMClient            *llm.Client
	Skill                string
	Messages             []llm.Message
	pendingMCPCallEnding []MCPToolInfo // Store MCP call info to print after LLM response
}

// NewAssistant creates a new assistant instance
func NewAssistant(llmClient *llm.Client, skill string) *Assistant {
	return &Assistant{
		LLMClient: llmClient,
		Skill:     skill,
		Messages:  []llm.Message{},
	}
}
