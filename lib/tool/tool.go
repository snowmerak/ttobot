package tool

import (
	"context"
	"fmt"
)

// ToolExecutor defines the interface for executing tools
type ToolExecutor interface {
	Execute(ctx context.Context, arguments map[string]any) (string, error)
}

// Tool represents a common tool structure that can be used across different APIs
type Tool struct {
	// The name of the tool
	Name string `json:"name"`

	// A human-readable description of the tool
	Description string `json:"description"`

	// Optional title for display purposes
	Title string `json:"title,omitempty"`

	// The function definition for the tool
	Function ToolFunction `json:"function"`

	// Executor for the tool (not serialized)
	Executor ToolExecutor `json:"-"`
}

// Execute executes the tool with the given arguments
func (t *Tool) Execute(ctx context.Context, arguments map[string]any) (string, error) {
	if t.Executor == nil {
		return "", fmt.Errorf("no executor available for tool %s", t.Name)
	}
	return t.Executor.Execute(ctx, arguments)
}

// ToolFunction represents the function definition of a tool
type ToolFunction struct {
	// The name of the function
	Name string `json:"name"`

	// A description of what the function does
	Description string `json:"description"`

	// Parameters schema for the function
	Parameters ParameterSchema `json:"parameters"`
}

// ParameterSchema represents the schema for function parameters
type ParameterSchema struct {
	// The type of the parameter (usually "object")
	Type string `json:"type"`

	// Required parameter names
	Required []string `json:"required,omitempty"`

	// Properties of the parameters
	Properties map[string]PropertyDefinition `json:"properties,omitempty"`

	// Additional schema definitions
	Defs any `json:"$defs,omitempty"`

	// Items for array types
	Items any `json:"items,omitempty"`
}

// PropertyDefinition represents a single property in the parameter schema
type PropertyDefinition struct {
	// The type of the property
	Type string `json:"type"`

	// Description of the property
	Description string `json:"description,omitempty"`

	// Items for array types
	Items any `json:"items,omitempty"`

	// Enum values for the property
	Enum []any `json:"enum,omitempty"`
}
