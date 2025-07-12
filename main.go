package main

import (
	"context"
	"log"

	"github.com/ollama/ollama/api"
	mcpConfig "github.com/snowmerak/ttobot/lib/mcp"
	"github.com/snowmerak/ttobot/pkg/mcp"
	"github.com/snowmerak/ttobot/pkg/ollama"
)

func main() {
	// Create MCP client
	client := mcp.NewClient("ttobot", "1.0.0")

	// Context without timeout for better tool performance
	ctx := context.Background()

	// Load configuration from YAML file
	configs, ollamaConfig, err := mcpConfig.LoadConfigWithOllamaFromFile("mcp.yaml")
	if err != nil {
		log.Printf("Failed to load config from file, trying default paths: %v", err)
		configs, err = mcpConfig.LoadConfigFromDefaultPath()
		if err != nil {
			log.Printf("Failed to load config from default paths: %v", err)
			log.Println("Using hardcoded example configuration...")

			// Fallback to hardcoded configuration
			configs = []mcpConfig.Config{
				{
					Name:    "memory-server",
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-memory"},
					Environment: map[string]string{
						"NODE_ENV": "production",
					},
				},
			}
		}
		// Set default Ollama config if not loaded from file
		ollamaConfig = mcpConfig.OllamaConfig{
			URL:   "http://localhost:11434",
			Model: "qwen3:14b",
		}
	}

	// Connect to MCP servers from configuration
	log.Printf("Connecting to %d MCP servers from configuration...", len(configs))
	err = client.ConnectFromConfigs(ctx, configs)
	if err != nil {
		log.Printf("Failed to connect to some servers: %v", err)
		// Continue even if some servers fail to connect
	}

	// Get tools from all connected servers
	log.Println("Fetching tools from connected servers...")
	tools, err := client.Tools(ctx)
	if err != nil {
		log.Fatalf("Failed to get tools: %v", err)
	}

	log.Printf("Found %d tools", len(tools))

	// Create Ollama client
	log.Printf("Creating Ollama client with URL: %s, Model: %s", ollamaConfig.URL, ollamaConfig.Model)
	ollamaClient, err := ollama.NewClient(ollama.ClientOptions{
		URL:   ollamaConfig.URL,
		Model: ollamaConfig.Model,
	})
	if err != nil {
		log.Fatalf("Failed to create Ollama client: %v", err)
	}

	// Set tools in Ollama client
	ollamaClient.SetTools(tools)

	// Test multiple chat examples with tools
	testQuestions := []string{
		"Can you search for files in the current directory? I want to see what Go files are available.",
		"Please create a new entity in the knowledge graph with the name 'TToBot' and description 'A Go-based chatbot with MCP tool integration'.",
		"Search the web for information about 'Model Context Protocol'.",
		"List all the directories you can access.",
	}

	for i, question := range testQuestions {
		log.Printf("\n=== Test %d ===", i+1)
		log.Printf("Question: %s", question)

		messages := []api.Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant with access to various tools. When a user asks for something that requires using a tool, you should use the appropriate tool to help them. You have access to file system operations, knowledge graph management, web search, and more. Always try to use tools when they can help answer the user's question.",
			},
			{
				Role:    "user",
				Content: question,
			},
		}

		response, err := ollamaClient.Chat(ctx, messages)
		if err != nil {
			log.Printf("Chat request failed: %v", err)
			continue
		}

		log.Printf("Raw response: %+v", response)
		log.Printf("Message: %+v", response.Message)
		log.Printf("Chat response content: '%s'", response.Message.Content)
		log.Printf("Response done: %v", response.Done)

		// Handle tool calls if any
		if len(response.Message.ToolCalls) > 0 {
			log.Printf("Processing %d tool calls...", len(response.Message.ToolCalls))
			toolMessages, err := ollamaClient.HandleToolCallsInResponse(ctx, response)
			if err != nil {
				log.Printf("Tool call handling failed: %v", err)
			} else {
				log.Printf("Generated %d tool result messages", len(toolMessages))
				for j, msg := range toolMessages {
					log.Printf("Tool result %d: %s", j+1, msg.Content)
				}
			}
		} else {
			log.Printf("No tool calls were made for this question")
		}
	}

	log.Println("\nMCP client test completed successfully!")
}
