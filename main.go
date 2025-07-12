package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ollama/ollama/api"
	mcpConfig "github.com/snowmerak/ttobot/lib/mcp"
	"github.com/snowmerak/ttobot/pkg/mcp"
	"github.com/snowmerak/ttobot/pkg/ollama"
)

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./ttobot \"your question here\"")
		os.Exit(1)
	}

	userQuery := strings.Join(os.Args[1:], " ")
	ctx := context.Background()

	// Load configuration
	configs, ollamaConfig, err := mcpConfig.LoadConfigWithOllamaFromFile("mcp.yaml")
	if err != nil {
		configs, err = mcpConfig.LoadConfigFromDefaultPath()
		if err != nil {
			configs = []mcpConfig.Config{
				{
					Name:    "memory-server",
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-memory"},
				},
			}
		}
		ollamaConfig = mcpConfig.OllamaConfig{
			URL:   "http://localhost:11434",
			Model: "qwen3:14b",
		}
	}

	// Create and connect MCP client
	mcpClient := mcp.NewClient("ttobot", "1.0.0")
	err = mcpClient.ConnectFromConfigs(ctx, configs)
	if err != nil {
		log.Fatalf("Failed to connect to MCP servers: %v", err)
	}

	// Get tools
	tools, err := mcpClient.Tools(ctx)
	if err != nil {
		log.Fatalf("Failed to get tools: %v", err)
	}

	// Create Ollama client
	ollamaClient, err := ollama.NewClient(ollama.ClientOptions{
		URL:   ollamaConfig.URL,
		Model: ollamaConfig.Model,
	})
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Set tools
	ollamaClient.SetTools(tools)

	fmt.Printf("Question: %s\n", userQuery)

	messages := []api.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant with access to various tools. Use the appropriate tools to answer user questions whenever possible.",
		},
		{
			Role:    "user",
			Content: userQuery,
		},
	}

	// Send to Ollama
	response, err := ollamaClient.Chat(ctx, messages)
	if err != nil {
		log.Fatalf("Chat request failed: %v", err)
	}

	// Show response
	if response.Message.Content != "" {
		fmt.Printf("Response: %s\n", response.Message.Content)
	}

	// Handle tool calls if any
	if len(response.Message.ToolCalls) > 0 {
		fmt.Printf("🔧 Tools called: %d\n", len(response.Message.ToolCalls))

		for i, toolCall := range response.Message.ToolCalls {
			fmt.Printf("  %d. %s\n", i+1, toolCall.Function.Name)
			if len(toolCall.Function.Arguments) > 0 {
				fmt.Printf("     Arguments: %v\n", toolCall.Function.Arguments)
			}
		}
		fmt.Println()

		fmt.Println("⚙️  Executing tools...")
		toolResults, err := ollamaClient.HandleToolCallsInResponse(ctx, response)
		if err != nil {
			log.Printf("Tool execution failed: %v", err)
		} else {
			for i, result := range toolResults {
				fmt.Printf("� Tool %d result:\n%s\n\n", i+1, result.Content)
			}
		}
	} else {
		fmt.Println("ℹ️  No tools were called for this query")
	}

	fmt.Println("✨ Done!")
}
