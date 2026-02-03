package cli

import (
	"fmt"

	"cnb.cool/znb/learn-skills/internal/llm"
)

// printPendingMCPCallEndings prints all pending MCP call ending info and clears the list
func (a *Assistant) printPendingMCPCallEndings() {
	for _, info := range a.pendingMCPCallEnding {
		fmt.Print(formatMCPCallEnd(info))
	}
	// Clear the list
	a.pendingMCPCallEnding = nil
}

// Initialize sets up the assistant with system prompt
func (a *Assistant) Initialize() error {
	// Add skill as system message
	a.Messages = append(a.Messages, llm.Message{
		Role:    "system",
		Content: a.Skill,
	})

	return nil
}

// ProcessMessage handles a user message and returns the assistant's response
func (a *Assistant) ProcessMessage(userMessage string) (string, error) {
	// Add user message
	a.Messages = append(a.Messages, llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	// Get CNB tools
	tools := GetCNBTools()

	// Tool calling loop (max 10 iterations to prevent infinite loops)
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		// Use streaming for final response (when no tool calls expected)
		// For now, we'll always use non-streaming to handle tool calls properly
		// In the future, we can detect if tools might be needed and choose accordingly

		resp, err := a.LLMClient.Chat(a.Messages, tools)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no response from LLM")
		}

		assistantMsg := resp.Choices[0].Message
		finishReason := resp.Choices[0].FinishReason

		// Add assistant response to history
		a.Messages = append(a.Messages, assistantMsg)

		// Check if LLM wants to call tools
		if finishReason == "tool_calls" && len(assistantMsg.ToolCalls) > 0 {
			// Execute each tool call
			for _, toolCall := range assistantMsg.ToolCalls {
				result, err := a.ExecuteTool(toolCall.Function.Name, toolCall.Function.Arguments)

				var toolResultContent string
				if err != nil {
					toolResultContent = fmt.Sprintf("Error: %v", err)
				} else {
					toolResultContent = result
				}

				// Add tool result to messages
				a.Messages = append(a.Messages, llm.Message{
					Role:       "tool",
					Content:    toolResultContent,
					ToolCallID: toolCall.ID,
				})
			}
			// Continue loop to let LLM process tool results
			continue
		}

		// LLM finished (no more tool calls)
		// Print any pending MCP call ending info
		a.printPendingMCPCallEndings()
		return assistantMsg.Content, nil
	}

	return "", fmt.Errorf("exceeded maximum tool calling iterations")
}

// ProcessMessageStream handles a user message with streaming output
// callback is called for each chunk of the response
func (a *Assistant) ProcessMessageStream(userMessage string, callback llm.StreamCallback) (string, error) {
	// Add user message
	a.Messages = append(a.Messages, llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	// Get CNB tools
	tools := GetCNBTools()

	// Tool calling loop (max 10 iterations to prevent infinite loops)
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		// Call LLM with streaming
		resp, err := a.LLMClient.ChatStream(a.Messages, tools, callback)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no response from LLM")
		}

		assistantMsg := resp.Choices[0].Message
		finishReason := resp.Choices[0].FinishReason

		// Add assistant response to history
		a.Messages = append(a.Messages, assistantMsg)

		// Check if LLM wants to call tools
		if finishReason == "tool_calls" && len(assistantMsg.ToolCalls) > 0 {
			// Execute each tool call (non-streaming)
			for _, toolCall := range assistantMsg.ToolCalls {
				result, err := a.ExecuteTool(toolCall.Function.Name, toolCall.Function.Arguments)

				var toolResultContent string
				if err != nil {
					toolResultContent = fmt.Sprintf("Error: %v", err)
				} else {
					toolResultContent = result
				}

				// Add tool result to messages
				a.Messages = append(a.Messages, llm.Message{
					Role:       "tool",
					Content:    toolResultContent,
					ToolCallID: toolCall.ID,
				})
			}
			// Continue loop with tool results, use streaming for next response
			continue
		}

		// LLM finished (no more tool calls)
		// Print any pending MCP call ending info
		a.printPendingMCPCallEndings()
		return assistantMsg.Content, nil
	}

	return "", fmt.Errorf("exceeded maximum tool calling iterations")
}

// Reset clears conversation history (keeps system message)
func (a *Assistant) Reset() {
	systemMsg := a.Messages[0]
	a.Messages = []llm.Message{systemMsg}
}
