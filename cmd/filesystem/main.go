package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetCurrentDirParams represents parameters for getting current directory
type GetCurrentDirParams struct{}

// FindFilesParams represents parameters for finding files
type FindFilesParams struct {
	Pattern   string `json:"pattern" mcp:"regular expression pattern to match file names"`
	Directory string `json:"directory,omitempty" mcp:"directory to search in (default: current directory)"`
	Recursive bool   `json:"recursive,omitempty" mcp:"whether to search recursively (default: false)"`
}

// SearchInFilesParams represents parameters for searching text in files
type SearchInFilesParams struct {
	SearchText string `json:"search_text" mcp:"text to search for in files"`
	Directory  string `json:"directory,omitempty" mcp:"directory to search in (default: current directory)"`
	FileFilter string `json:"file_filter,omitempty" mcp:"regex pattern to filter files (default: match all files)"`
	Recursive  bool   `json:"recursive,omitempty" mcp:"whether to search recursively (default: true)"`
}

// CreateFileParams represents parameters for creating a file
type CreateFileParams struct {
	Path    string `json:"path" mcp:"path of the file to create"`
	Content string `json:"content,omitempty" mcp:"content to write to the file (default: empty)"`
}

// CreateDirParams represents parameters for creating a directory
type CreateDirParams struct {
	Path string `json:"path" mcp:"path of the directory to create"`
}

// RemoveParams represents parameters for removing files/directories
type RemoveParams struct {
	Path string `json:"path" mcp:"path of the file or directory to remove"`
}

// WriteFileParams represents parameters for writing to a file
type WriteFileParams struct {
	Path    string `json:"path" mcp:"path of the file to write to"`
	Content string `json:"content" mcp:"content to write to the file"`
}

// ReadFileParams represents parameters for reading a file
type ReadFileParams struct {
	Path string `json:"path" mcp:"path of the file to read"`
}

// GetCurrentDir returns the current working directory
func GetCurrentDir(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[GetCurrentDirParams]) (*mcp.CallToolResultFor[any], error) {
	cwd, err := os.Getwd()
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting current directory: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Current directory: %s", cwd)}},
	}, nil
}

// FindFiles finds files matching a regular expression pattern
func FindFiles(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[FindFilesParams]) (*mcp.CallToolResultFor[any], error) {
	directory := params.Arguments.Directory
	if directory == "" {
		var err error
		directory, err = os.Getwd()
		if err != nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting current directory: %v", err)}},
				IsError: true,
			}, nil
		}
	}

	regex, err := regexp.Compile(params.Arguments.Pattern)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid regex pattern: %v", err)}},
			IsError: true,
		}, nil
	}

	var matches []string

	if params.Arguments.Recursive {
		err = filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && regex.MatchString(filepath.Base(path)) {
				matches = append(matches, path)
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error reading directory: %v", err)}},
				IsError: true,
			}, nil
		}

		for _, entry := range entries {
			if !entry.IsDir() && regex.MatchString(entry.Name()) {
				matches = append(matches, filepath.Join(directory, entry.Name()))
			}
		}
	}

	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error searching files: %v", err)}},
			IsError: true,
		}, nil
	}

	result := fmt.Sprintf("Found %d files matching pattern '%s':\n", len(matches), params.Arguments.Pattern)
	for _, match := range matches {
		result += fmt.Sprintf("- %s\n", match)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// SearchInFiles searches for text within files
func SearchInFiles(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[SearchInFilesParams]) (*mcp.CallToolResultFor[any], error) {
	directory := params.Arguments.Directory
	if directory == "" {
		var err error
		directory, err = os.Getwd()
		if err != nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting current directory: %v", err)}},
				IsError: true,
			}, nil
		}
	}

	var fileFilter *regexp.Regexp
	if params.Arguments.FileFilter != "" {
		var err error
		fileFilter, err = regexp.Compile(params.Arguments.FileFilter)
		if err != nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid file filter regex: %v", err)}},
				IsError: true,
			}, nil
		}
	}

	var matches []string

	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Apply file filter if specified
		if fileFilter != nil && !fileFilter.MatchString(filepath.Base(path)) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			// Skip files that can't be read
			return nil
		}

		if strings.Contains(string(content), params.Arguments.SearchText) {
			matches = append(matches, path)
		}
		return nil
	}

	if params.Arguments.Recursive {
		err := filepath.WalkDir(directory, walkFunc)
		if err != nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error searching in files: %v", err)}},
				IsError: true,
			}, nil
		}
	} else {
		entries, err := os.ReadDir(directory)
		if err != nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error reading directory: %v", err)}},
				IsError: true,
			}, nil
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				path := filepath.Join(directory, entry.Name())
				walkFunc(path, entry, nil)
			}
		}
	}

	result := fmt.Sprintf("Found text '%s' in %d files:\n", params.Arguments.SearchText, len(matches))
	for _, match := range matches {
		result += fmt.Sprintf("- %s\n", match)
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: result}},
	}, nil
}

// CreateFile creates a new file
func CreateFile(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateFileParams]) (*mcp.CallToolResultFor[any], error) {
	// Create parent directories if they don't exist
	dir := filepath.Dir(params.Arguments.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating parent directories: %v", err)}},
			IsError: true,
		}, nil
	}

	file, err := os.Create(params.Arguments.Path)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating file: %v", err)}},
			IsError: true,
		}, nil
	}
	defer file.Close()

	if params.Arguments.Content != "" {
		_, err = file.WriteString(params.Arguments.Content)
		if err != nil {
			return &mcp.CallToolResultFor[any]{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error writing to file: %v", err)}},
				IsError: true,
			}, nil
		}
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Successfully created file: %s", params.Arguments.Path)}},
	}, nil
}

// CreateDir creates a new directory
func CreateDir(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateDirParams]) (*mcp.CallToolResultFor[any], error) {
	err := os.MkdirAll(params.Arguments.Path, 0755)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating directory: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Successfully created directory: %s", params.Arguments.Path)}},
	}, nil
}

// RemoveFileOrDir removes a file or directory
func RemoveFileOrDir(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[RemoveParams]) (*mcp.CallToolResultFor[any], error) {
	err := os.RemoveAll(params.Arguments.Path)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error removing: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Successfully removed: %s", params.Arguments.Path)}},
	}, nil
}

// WriteFile writes content to a file
func WriteFile(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[WriteFileParams]) (*mcp.CallToolResultFor[any], error) {
	// Create parent directories if they don't exist
	dir := filepath.Dir(params.Arguments.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating parent directories: %v", err)}},
			IsError: true,
		}, nil
	}

	err := os.WriteFile(params.Arguments.Path, []byte(params.Arguments.Content), 0644)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error writing to file: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Successfully wrote to file: %s", params.Arguments.Path)}},
	}, nil
}

// ReadFile reads content from a file
func ReadFile(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[ReadFileParams]) (*mcp.CallToolResultFor[any], error) {
	content, err := os.ReadFile(params.Arguments.Path)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error reading file: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: string(content)}},
	}, nil
}

// CopyFile copies a file from source to destination
func CopyFile(ctx context.Context, cc *mcp.ServerSession, params *mcp.CallToolParamsFor[struct {
	Source string `json:"source" mcp:"source file path"`
	Dest   string `json:"dest" mcp:"destination file path"`
}]) (*mcp.CallToolResultFor[any], error) {
	sourceFile, err := os.Open(params.Arguments.Source)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error opening source file: %v", err)}},
			IsError: true,
		}, nil
	}
	defer sourceFile.Close()

	// Create parent directories if they don't exist
	dir := filepath.Dir(params.Arguments.Dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating parent directories: %v", err)}},
			IsError: true,
		}, nil
	}

	destFile, err := os.Create(params.Arguments.Dest)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating destination file: %v", err)}},
			IsError: true,
		}, nil
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return &mcp.CallToolResultFor[any]{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error copying file: %v", err)}},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResultFor[any]{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Successfully copied file from %s to %s", params.Arguments.Source, params.Arguments.Dest)}},
	}, nil
}

func main() {
	// Create a server for file system operations
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "filesystem",
		Version: "v1.0.0",
	}, nil)

	// Register tools
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_current_dir",
		Description: "Get the current working directory",
	}, GetCurrentDir)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_files",
		Description: "Find files matching a regular expression pattern",
	}, FindFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_in_files",
		Description: "Search for text within files",
	}, SearchInFiles)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_file",
		Description: "Create a new file with optional content",
	}, CreateFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_dir",
		Description: "Create a new directory",
	}, CreateDir)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove",
		Description: "Remove a file or directory",
	}, RemoveFileOrDir)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "write_file",
		Description: "Write content to a file",
	}, WriteFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_file",
		Description: "Read content from a file",
	}, ReadFile)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "copy_file",
		Description: "Copy a file from source to destination",
	}, CopyFile)

	// Run the server over stdin/stdout, until the client disconnects
	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatal(err)
	}
}
