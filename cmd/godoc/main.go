package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GodocParams represents parameters for godoc tool
type GodocParams struct {
	PackagePath string `json:"package_path" mcp:"package path to generate documentation for"`
	Output      string `json:"output,omitempty" mcp:"output format (html, text) - default: text"`
}

// GoFmtParams represents parameters for go fmt
type GoFmtParams struct {
	FilePath string `json:"file_path" mcp:"path to Go file to format"`
	Write    bool   `json:"write,omitempty" mcp:"write result to file instead of stdout"`
}

// GoVetParams represents parameters for go vet
type GoVetParams struct {
	PackagePath string `json:"package_path,omitempty" mcp:"package path to vet (default: current directory)"`
}

// GoTestParams represents parameters for go test
type GoTestParams struct {
	PackagePath string `json:"package_path,omitempty" mcp:"package path to test (default: current directory)"`
	Verbose     bool   `json:"verbose,omitempty" mcp:"verbose output"`
	Cover       bool   `json:"cover,omitempty" mcp:"enable coverage analysis"`
}

// GoBuildParams represents parameters for go build
type GoBuildParams struct {
	PackagePath string `json:"package_path,omitempty" mcp:"package path to build (default: current directory)"`
	Output      string `json:"output,omitempty" mcp:"output binary name"`
	Tags        string `json:"tags,omitempty" mcp:"build tags"`
}

// GoModParams represents parameters for go mod commands
type GoModParams struct {
	Command string `json:"command" mcp:"mod command (init, tidy, download, verify, why, graph)"`
	Args    string `json:"args,omitempty" mcp:"additional arguments"`
}

// GoDocTool generates documentation for Go packages
func GoDocTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GodocParams]) (*mcp.CallToolResultFor[any], error) {
	args := []string{"doc"}

	if params.Arguments.PackagePath != "" {
		args = append(args, params.Arguments.PackagePath)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error running go doc: %v\nOutput: %s", err, string(output))}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
	}, nil
}

// GoFmtTool formats Go source code
func GoFmtTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GoFmtParams]) (*mcp.CallToolResultFor[any], error) {
	args := []string{"fmt"}

	if params.Arguments.FilePath != "" {
		// Check if file exists
		if _, err := os.Stat(params.Arguments.FilePath); os.IsNotExist(err) {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("File does not exist: %s", params.Arguments.FilePath)}},
				IsError: true,
			}, nil
		}
		args = append(args, params.Arguments.FilePath)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error running go fmt: %v\nOutput: %s", err, string(output))}},
			IsError: true,
		}, nil
	}

	result := "Go fmt completed successfully"
	if len(output) > 0 {
		result += "\nOutput: " + string(output)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// GoVetTool examines Go source code and reports suspicious constructs
func GoVetTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GoVetParams]) (*mcp.CallToolResultFor[any], error) {
	args := []string{"vet"}

	if params.Arguments.PackagePath != "" {
		args = append(args, params.Arguments.PackagePath)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Go vet found issues:\n%s", string(output))}},
			IsError: false, // vet finding issues is not an error in tool execution
		}, nil
	}

	result := "Go vet completed successfully - no issues found"
	if len(output) > 0 {
		result = fmt.Sprintf("Go vet output:\n%s", string(output))
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// GoTestTool runs Go tests
func GoTestTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GoTestParams]) (*mcp.CallToolResultFor[any], error) {
	args := []string{"test"}

	if params.Arguments.Verbose {
		args = append(args, "-v")
	}

	if params.Arguments.Cover {
		args = append(args, "-cover")
	}

	if params.Arguments.PackagePath != "" {
		args = append(args, params.Arguments.PackagePath)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Go test failed:\n%s", string(output))}},
			IsError: false, // test failures are not tool execution errors
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
	}, nil
}

// GoBuildTool builds Go packages
func GoBuildTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GoBuildParams]) (*mcp.CallToolResultFor[any], error) {
	args := []string{"build"}

	if params.Arguments.Output != "" {
		args = append(args, "-o", params.Arguments.Output)
	}

	if params.Arguments.Tags != "" {
		args = append(args, "-tags", params.Arguments.Tags)
	}

	if params.Arguments.PackagePath != "" {
		args = append(args, params.Arguments.PackagePath)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Go build failed: %v\nOutput: %s", err, string(output))}},
			IsError: true,
		}, nil
	}

	result := "Go build completed successfully"
	if len(output) > 0 {
		result += "\nOutput: " + string(output)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// GoModTool handles go mod commands
func GoModTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GoModParams]) (*mcp.CallToolResultFor[any], error) {
	validCommands := map[string]bool{
		"init":     true,
		"tidy":     true,
		"download": true,
		"verify":   true,
		"why":      true,
		"graph":    true,
	}

	if !validCommands[params.Arguments.Command] {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid mod command: %s. Valid commands: init, tidy, download, verify, why, graph", params.Arguments.Command)}},
			IsError: true,
		}, nil
	}

	args := []string{"mod", params.Arguments.Command}

	if params.Arguments.Args != "" {
		// Split additional arguments
		additionalArgs := strings.Fields(params.Arguments.Args)
		args = append(args, additionalArgs...)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Go mod %s failed: %v\nOutput: %s", params.Arguments.Command, err, string(output))}},
			IsError: true,
		}, nil
	}

	result := fmt.Sprintf("Go mod %s completed successfully", params.Arguments.Command)
	if len(output) > 0 {
		result += "\nOutput: " + string(output)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// GoVersionTool gets Go version information
func GoVersionTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[struct{}]) (*mcp.CallToolResultFor[any], error) {
	cmd := exec.CommandContext(ctx, "go", "version")
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting Go version: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
	}, nil
}

// GoEnvTool gets Go environment information
func GoEnvTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[struct {
	Variable string `json:"variable,omitempty" mcp:"specific environment variable to get (optional)"`
}]) (*mcp.CallToolResultFor[any], error) {
	args := []string{"env"}

	if params.Arguments.Variable != "" {
		args = append(args, params.Arguments.Variable)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting Go environment: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
	}, nil
}

// GoListTool lists Go packages
func GoListTool(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[struct {
	Pattern string `json:"pattern,omitempty" mcp:"package pattern to list (default: all packages in current module)"`
	Json    bool   `json:"json,omitempty" mcp:"output in JSON format"`
}]) (*mcp.CallToolResultFor[any], error) {
	args := []string{"list"}

	if params.Arguments.Json {
		args = append(args, "-json")
	}

	if params.Arguments.Pattern != "" {
		args = append(args, params.Arguments.Pattern)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error listing Go packages: %v\nOutput: %s", err, string(output))}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(output)}},
	}, nil
}

func main() {
	// Create a server for Go development tools
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "godoc",
		Version: "v1.0.0",
	}, nil)

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_doc",
		Description: "Generate documentation for Go packages using 'go doc'",
	}, GoDocTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_fmt",
		Description: "Format Go source code using 'go fmt'",
	}, GoFmtTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_vet",
		Description: "Examine Go source code and report suspicious constructs using 'go vet'",
	}, GoVetTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_test",
		Description: "Run Go tests using 'go test'",
	}, GoTestTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_build",
		Description: "Build Go packages using 'go build'",
	}, GoBuildTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_mod",
		Description: "Handle Go module operations using 'go mod'",
	}, GoModTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_version",
		Description: "Get Go version information",
	}, GoVersionTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_env",
		Description: "Get Go environment information",
	}, GoEnvTool)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "go_list",
		Description: "List Go packages",
	}, GoListTool)

	// Run the server over stdin/stdout, until the client disconnects
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatal(err)
	}
}
