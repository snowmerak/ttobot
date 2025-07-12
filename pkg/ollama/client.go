package ollama

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
	"github.com/snowmerak/ttobot/lib/tool"
)

type Client struct {
	model  string
	client *api.Client
	tools  []tool.Tool
}

type ClientOptions struct {
	URL   string
	Model string
}

func NewClient(opt ClientOptions) (*Client, error) {
	u, err := url.Parse(opt.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %s: %w", opt.URL, err)
	}

	hc := &http.Client{}

	client := api.NewClient(u, hc)

	return &Client{
		model:  opt.Model,
		client: client,
		tools:  []tool.Tool{},
	}, nil
}

// SetTools sets the available tools for the client
func (c *Client) SetTools(tools []tool.Tool) {
	c.tools = tools
}

// GetTools returns the currently available tools
func (c *Client) GetTools() []tool.Tool {
	return c.tools
}

// convertToOllamaTools converts common tool format to Ollama API format
func (c *Client) convertToOllamaTools() []api.Tool {
	ollamaTools := make([]api.Tool, 0, len(c.tools))

	for _, t := range c.tools {
		ollamaTool := api.Tool{
			Type: "function",
			Function: api.ToolFunction{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters: struct {
					Type       string   `json:"type"`
					Defs       any      `json:"$defs,omitempty"`
					Items      any      `json:"items,omitempty"`
					Required   []string `json:"required"`
					Properties map[string]struct {
						Type        api.PropertyType `json:"type"`
						Items       any              `json:"items,omitempty"`
						Description string           `json:"description"`
						Enum        []any            `json:"enum,omitempty"`
					} `json:"properties"`
				}{
					Type:     t.Function.Parameters.Type,
					Defs:     t.Function.Parameters.Defs,
					Items:    t.Function.Parameters.Items,
					Required: t.Function.Parameters.Required,
					Properties: make(map[string]struct {
						Type        api.PropertyType `json:"type"`
						Items       any              `json:"items,omitempty"`
						Description string           `json:"description"`
						Enum        []any            `json:"enum,omitempty"`
					}),
				},
			},
		}

		// Convert properties
		for propName, propDef := range t.Function.Parameters.Properties {
			ollamaTool.Function.Parameters.Properties[propName] = struct {
				Type        api.PropertyType `json:"type"`
				Items       any              `json:"items,omitempty"`
				Description string           `json:"description"`
				Enum        []any            `json:"enum,omitempty"`
			}{
				Type:        api.PropertyType{propDef.Type},
				Items:       propDef.Items,
				Description: propDef.Description,
				Enum:        propDef.Enum,
			}
		}

		ollamaTools = append(ollamaTools, ollamaTool)
	}

	return ollamaTools
}

// Chat sends a chat request with tool support
func (c *Client) Chat(ctx context.Context, messages []api.Message) (*api.ChatResponse, error) {
	req := &api.ChatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   new(bool), // Disable streaming for complete response
	}

	// Add tools if available
	if len(c.tools) > 0 {
		req.Tools = c.convertToOllamaTools()
		log.Printf("Ollama chat: Sending request with %d tools available", len(c.tools))
	} else {
		log.Printf("Ollama chat: Sending request without tools")
	}

	var finalResponse api.ChatResponse
	var responseContent string

	err := c.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		finalResponse = resp
		if resp.Message.Content != "" {
			responseContent += resp.Message.Content
		}
		return nil
	})

	if err != nil {
		log.Printf("Ollama chat: Request failed: %v", err)
		return nil, fmt.Errorf("chat request failed: %w", err)
	}

	// Combine all content
	finalResponse.Message.Content = responseContent

	// Log tool calls if any
	if len(finalResponse.Message.ToolCalls) > 0 {
		log.Printf("Ollama chat: Response contains %d tool calls", len(finalResponse.Message.ToolCalls))
		for i, toolCall := range finalResponse.Message.ToolCalls {
			log.Printf("  Tool call %d: %s", i+1, toolCall.Function.Name)
			if toolCall.Function.Arguments != nil {
				log.Printf("    Arguments: %v", toolCall.Function.Arguments)
			}
		}
	} else {
		log.Printf("Ollama chat: Response completed without tool calls")
	}

	return &finalResponse, nil
}

// ChatStream sends a streaming chat request with tool support
func (c *Client) ChatStream(ctx context.Context, messages []api.Message, callback func(api.ChatResponse) error) error {
	req := &api.ChatRequest{
		Model:    c.model,
		Messages: messages,
	}

	// Add tools if available
	if len(c.tools) > 0 {
		req.Tools = c.convertToOllamaTools()
		log.Printf("Ollama chat stream: Starting with %d tools available", len(c.tools))
	} else {
		log.Printf("Ollama chat stream: Starting without tools")
	}

	// Wrap callback to add logging
	wrappedCallback := func(resp api.ChatResponse) error {
		// Log tool calls if any
		if len(resp.Message.ToolCalls) > 0 {
			log.Printf("Ollama chat stream: Received %d tool calls", len(resp.Message.ToolCalls))
			for i, toolCall := range resp.Message.ToolCalls {
				log.Printf("  Tool call %d: %s", i+1, toolCall.Function.Name)
				if toolCall.Function.Arguments != nil {
					log.Printf("    Arguments: %v", toolCall.Function.Arguments)
				}
			}
		}

		// Call the original callback
		return callback(resp)
	}

	err := c.client.Chat(ctx, req, wrappedCallback)
	if err != nil {
		log.Printf("Ollama chat stream: Request failed: %v", err)
		return fmt.Errorf("streaming chat request failed: %w", err)
	}

	log.Printf("Ollama chat stream: Completed successfully")
	return nil
}

// ExecuteToolCall executes a tool call and returns the result
func (c *Client) ExecuteToolCall(ctx context.Context, toolCall api.ToolCall) (string, error) {
	log.Printf("Ollama tool execution: Executing tool call %s", toolCall.Function.Name)

	// Find the tool by name
	var targetTool *tool.Tool
	for _, t := range c.tools {
		if t.Function.Name == toolCall.Function.Name {
			targetTool = &t
			break
		}
	}

	if targetTool == nil {
		return "", fmt.Errorf("tool %s not found", toolCall.Function.Name)
	}
	// Parse arguments
	arguments := map[string]any(toolCall.Function.Arguments)

	log.Printf("Ollama tool execution: Tool name: %s", toolCall.Function.Name)
	log.Printf("Ollama tool execution: Arguments: %v", arguments)

	// Execute the tool using its executor
	result, err := targetTool.Execute(ctx, arguments)
	if err != nil {
		log.Printf("Ollama tool execution: Execution failed: %v", err)
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	log.Printf("Ollama tool execution: Result: %s", result)
	return result, nil
}

// HandleToolCallsInResponse processes tool calls in a chat response and returns updated messages
func (c *Client) HandleToolCallsInResponse(ctx context.Context, response *api.ChatResponse) ([]api.Message, error) {
	if len(response.Message.ToolCalls) == 0 {
		return nil, nil
	}

	log.Printf("Ollama tool handling: Processing %d tool calls", len(response.Message.ToolCalls))

	var newMessages []api.Message

	for _, toolCall := range response.Message.ToolCalls {
		result, err := c.ExecuteToolCall(ctx, toolCall)
		if err != nil {
			log.Printf("Ollama tool handling: Tool call failed: %v", err)
			result = fmt.Sprintf("Tool execution failed: %v", err)
		}

		// Add tool result as a message
		toolMessage := api.Message{
			Role:    "tool",
			Content: result,
		}
		newMessages = append(newMessages, toolMessage)
	}

	log.Printf("Ollama tool handling: Created %d tool result messages", len(newMessages))
	return newMessages, nil
}
