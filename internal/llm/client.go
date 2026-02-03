package llm

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// Client wraps eino-ext OpenAI client with streaming support
type Client struct {
	chatModel model.ToolCallingChatModel
}

// NewClient creates a new LLM client using eino-ext
func NewClient(apiKey, baseURL, modelName string) (*Client, error) {
	ctx := context.Background()

	// Create eino-ext OpenAI chat model
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   modelName,
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	return &Client{
		chatModel: chatModel,
	}, nil
}

// Chat sends a chat completion request (non-streaming)
func (c *Client) Chat(messages []Message, tools []Tool) (*ChatResponse, error) {
	ctx := context.Background()

	// Convert messages to eino schema
	einoMessages := messagesToEino(messages)

	// Prepare options with tools if provided
	var opts []model.Option
	if len(tools) > 0 {
		toolInfos := toolsToEino(tools)
		opts = append(opts, model.WithTools(toolInfos))
	}

	// Generate response
	resp, err := c.chatModel.Generate(ctx, einoMessages, opts...)
	if err != nil {
		return nil, fmt.Errorf("generate failed: %w", err)
	}

	// Convert back to our format
	return einoToResponse(resp), nil
}

// StreamCallback is called for each chunk of streaming response
type StreamCallback func(chunk string) error

// ChatStream sends a chat completion request with streaming
func (c *Client) ChatStream(messages []Message, tools []Tool, callback StreamCallback) (*ChatResponse, error) {
	ctx := context.Background()

	// Convert messages to eino schema
	einoMessages := messagesToEino(messages)

	// Prepare options with tools if provided
	var opts []model.Option
	if len(tools) > 0 {
		toolInfos := toolsToEino(tools)
		opts = append(opts, model.WithTools(toolInfos))
	}

	// Stream response
	reader, err := c.chatModel.Stream(ctx, einoMessages, opts...)
	if err != nil {
		return nil, fmt.Errorf("stream failed: %w", err)
	}

	// Collect all chunks
	var chunks []*schema.Message
	for {
		chunk, err := reader.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("stream recv failed: %w", err)
		}

		chunks = append(chunks, chunk)

		// Call callback with chunk content
		if callback != nil && chunk.Content != "" {
			if err := callback(chunk.Content); err != nil {
				return nil, fmt.Errorf("callback failed: %w", err)
			}
		}
	}

	// Concatenate all chunks
	finalMsg, err := schema.ConcatMessages(chunks)
	if err != nil {
		return nil, fmt.Errorf("concat messages failed: %w", err)
	}

	return einoToResponse(finalMsg), nil
}

// messagesToEino converts our Message format to eino schema.Message
func messagesToEino(messages []Message) []*schema.Message {
	result := make([]*schema.Message, len(messages))
	for i, msg := range messages {
		einoMsg := &schema.Message{
			Role:    schema.RoleType(msg.Role),
			Content: msg.Content,
		}

		// Convert tool calls
		if len(msg.ToolCalls) > 0 {
			einoMsg.ToolCalls = make([]schema.ToolCall, len(msg.ToolCalls))
			for j, tc := range msg.ToolCalls {
				einoMsg.ToolCalls[j] = schema.ToolCall{
					ID: tc.ID,
					Function: schema.FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}

		// Set tool call ID if this is a tool message
		if msg.ToolCallID != "" {
			einoMsg.ToolCallID = msg.ToolCallID
		}

		result[i] = einoMsg
	}
	return result
}

// toolsToEino converts our Tool format to eino schema.ToolInfo
// Note: eino will handle the parameter schema internally
func toolsToEino(tools []Tool) []*schema.ToolInfo {
	result := make([]*schema.ToolInfo, len(tools))
	for i, tool := range tools {
		result[i] = &schema.ToolInfo{
			Name: tool.Function.Name,
			Desc: tool.Function.Description,
			Extra: map[string]any{
				"parameters": tool.Function.Parameters,
			},
		}
	}
	return result
}

// einoToResponse converts eino schema.Message back to our ChatResponse format
func einoToResponse(msg *schema.Message) *ChatResponse {
	ourMsg := Message{
		Role:    string(msg.Role),
		Content: msg.Content,
	}

	// Convert tool calls back
	if len(msg.ToolCalls) > 0 {
		ourMsg.ToolCalls = make([]ToolCall, len(msg.ToolCalls))
		for i, tc := range msg.ToolCalls {
			ourMsg.ToolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	// Determine finish reason
	finishReason := "stop"
	if len(msg.ToolCalls) > 0 {
		finishReason = "tool_calls"
	}

	return &ChatResponse{
		Choices: []Choice{
			{
				Index:        0,
				Message:      ourMsg,
				FinishReason: finishReason,
			},
		},
	}
}
