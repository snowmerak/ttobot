package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpConfig "github.com/snowmerak/ttobot/lib/mcp"
	"github.com/snowmerak/ttobot/lib/tool"
)

// generateUUID generates a simple UUID-like string
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateServerID generates a unique server ID
func generateServerID(originalID string) string {
	if originalID != "" {
		return originalID
	}

	// Generate ID with timestamp and UUID
	timestamp := time.Now().Format("20060102-150405")
	uuid := generateUUID()
	return fmt.Sprintf("mcp-server-%s-%s", timestamp, uuid[:8])
}

type Client struct {
	client      *mcp.Client
	servers     map[string]*mcp.ClientSession
	serverIDs   map[*mcp.ClientSession]string // Maps session to our generated ID
	serversLock sync.RWMutex
}

func NewClient(name string, version string) *Client {
	return &Client{
		client:    mcp.NewClient(&mcp.Implementation{Name: name, Version: version}, nil),
		servers:   make(map[string]*mcp.ClientSession),
		serverIDs: make(map[*mcp.ClientSession]string),
	}
}

func (c *Client) Connect(ctx context.Context, filepath string, args ...string) error {
	ct := mcp.NewCommandTransport(exec.CommandContext(ctx, filepath, args...))
	return c.connectWithTransport(ctx, ct)
}

// ConnectWithCommand connects to an MCP server using a pre-configured command
func (c *Client) ConnectWithCommand(ctx context.Context, cmd *exec.Cmd) error {
	ct := mcp.NewCommandTransport(cmd)
	return c.connectWithTransport(ctx, ct)
}

// connectWithTransport handles the common connection logic
func (c *Client) connectWithTransport(ctx context.Context, ct *mcp.CommandTransport) error {
	ss, err := c.client.Connect(ctx, ct)
	if err != nil {
		return fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	c.serversLock.Lock()
	defer c.serversLock.Unlock()

	// Generate a unique server ID if the original ID is empty
	originalID := ss.ID()
	serverID := generateServerID(originalID)

	// Check if server with this ID already exists
	_, ok := c.servers[serverID]
	if ok {
		ss.Close()
		return fmt.Errorf("server with ID %s already exists", serverID)
	}

	// Store the server with the generated ID
	c.servers[serverID] = ss
	c.serverIDs[ss] = serverID

	return nil
}

func (c *Client) Tools(ctx context.Context) ([]tool.Tool, error) {
	c.serversLock.RLock()
	defer c.serversLock.RUnlock()

	if len(c.servers) == 0 {
		return nil, fmt.Errorf("no servers connected")
	}

	var result []tool.Tool

	for _, server := range c.servers {
		for mcpTool, err := range server.Tools(ctx, &mcp.ListToolsParams{}) {
			if err != nil {
				return nil, fmt.Errorf("failed to list tools: %w", err)
			}

			if mcpTool == nil {
				continue
			} // Create the common tool structure with server ID prefix
			serverID := c.serverIDs[server]
			toolName := fmt.Sprintf("%s:%s", serverID, mcpTool.Name)

			commonTool := tool.Tool{
				Name:        toolName,
				Description: mcpTool.Description,
				Title:       mcpTool.Title,
				Function: tool.ToolFunction{
					Name:        toolName,
					Description: mcpTool.Description,
					Parameters: tool.ParameterSchema{
						Type:       "object",
						Properties: make(map[string]tool.PropertyDefinition),
						Required:   []string{},
					},
				},
				Executor: &MCPToolExecutor{
					client:       c,
					serverID:     serverID,
					toolName:     mcpTool.Name, // Original tool name without server prefix
					originalTool: mcpTool,
				},
			}

			// Convert MCP input schema to common parameter schema
			if mcpTool.InputSchema != nil {
				if err := ConvertViaJSON(mcpTool.InputSchema, &commonTool.Function.Parameters); err != nil {
					return nil, fmt.Errorf("failed to convert input schema for tool %s: %w", mcpTool.Name, err)
				}
			}

			result = append(result, commonTool)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no tools found")
	}

	return result, nil
}

// MCPToolExecutor implements the ToolExecutor interface for MCP tools
type MCPToolExecutor struct {
	client       *Client
	serverID     string
	toolName     string
	originalTool *mcp.Tool
}

// Execute executes the MCP tool with the given arguments
func (e *MCPToolExecutor) Execute(ctx context.Context, arguments map[string]any) (string, error) {
	e.client.serversLock.RLock()
	server, exists := e.client.servers[e.serverID]
	e.client.serversLock.RUnlock()

	if !exists {
		return "", fmt.Errorf("server %s not found", e.serverID)
	}

	// Convert arguments to MCP format
	params := &mcp.CallToolParams{
		Name:      e.toolName,
		Arguments: arguments,
	}

	// Call the tool
	result, err := server.CallTool(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to call tool %s: %w", e.toolName, err)
	}

	// Convert result to string
	if result.Content != nil {
		// Handle different content types
		var content strings.Builder
		for _, c := range result.Content {
			// Try to convert to TextContent
			if textContent, ok := c.(*mcp.TextContent); ok {
				content.WriteString(textContent.Text)
			} else {
				// For other content types, try to marshal as JSON
				if jsonBytes, err := c.MarshalJSON(); err == nil {
					content.Write(jsonBytes)
				}
			}
		}
		return content.String(), nil
	}

	return "Tool executed successfully", nil
}

// ConnectFromConfig connects to an MCP server using the configuration
func (c *Client) ConnectFromConfig(ctx context.Context, config mcpConfig.Config) error {
	// Create command from config
	cmd := config.CreateCommand(ctx)

	// Connect to the server
	return c.ConnectWithCommand(ctx, cmd)
}

// ConnectFromConfigs connects to multiple MCP servers from configurations
func (c *Client) ConnectFromConfigs(ctx context.Context, configs []mcpConfig.Config) error {
	for _, config := range configs {
		if err := c.ConnectFromConfig(ctx, config); err != nil {
			return fmt.Errorf("failed to connect to server %s: %w", config.Name, err)
		}
	}
	return nil
}
