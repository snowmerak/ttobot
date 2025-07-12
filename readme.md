# ttobot

## Overview

`ttobot` is a Go-based intelligent assistant that combines the power of Ollama LLM with Model Context Protocol (MCP) servers. It enables users to interact with AI models while providing the model with access to various tools and capabilities through MCP servers. The project serves as both a command-line tool and a framework for building AI assistants with extensible tool support.

## Project Structure

```
ttobot/
├── cmd/                    # Command-line tools and MCP servers
│   └── filesystem/         # Filesystem MCP server (file operations)
│       └── main.go
├── lib/                    # Core libraries
│   ├── mcp/               # MCP configuration management
│   │   └── config.go
│   └── tool/              # Tool abstraction and execution
│       └── tool.go
├── pkg/                    # Reusable packages
│   ├── mcp/               # MCP client implementation
│   │   ├── client.go      # MCP client with multi-server support
│   │   └── convert.go     # Tool conversion utilities
│   └── ollama/            # Ollama client integration
│       └── client.go      # Ollama client with tool support
├── go.mod                  # Go module definition
├── go.sum                  # Go dependencies checksum
├── main.go                 # Main CLI application
├── mcp.yaml               # MCP servers and Ollama configuration
└── readme.md              # Project documentation
```

## Getting Started

### Prerequisites
- Go 1.24 or newer
- Ollama server running (default: http://localhost:11434)
- Compatible MCP servers (optional - filesystem server included)

### Installation
Clone the repository:

```zsh
git clone https://github.com/snowmerak/ttobot.git
cd ttobot
```

Install dependencies:

```zsh
go mod tidy
```

### Configuration
Configure your MCP servers and Ollama settings in `mcp.yaml`:

```yaml
servers:
  - name: "filesystem"
    command: "go"
    args:
      - "run"
      - "./cmd/filesystem/."
ollama:
  url: "http://localhost:11434"
  model: "qwen3:14b"
```

### Usage

#### Basic Usage
Ask questions and let the AI use available tools:

```zsh
go run main.go "What files are in the current directory?"
```

```zsh
go run main.go "Create a new file called test.txt with some content"
```

```zsh
go run main.go "Search for all Go files in this project"
```

#### Running the Filesystem MCP Server
The filesystem server can be run independently:

```zsh
go run ./cmd/filesystem/main.go
```

#### Building
Build the main application:

```zsh
go build -o ttobot main.go
./ttobot "your question here"
```

## Features

### Core Functionality
- **AI Assistant**: Interactive command-line AI assistant powered by Ollama
- **Tool Integration**: Seamlessly integrates AI models with MCP tools
- **Multi-Server Support**: Connect to multiple MCP servers simultaneously
- **Flexible Configuration**: YAML-based configuration for easy customization

### Built-in Tools (Filesystem Server)
- **File Operations**: Create, read, write, and delete files
- **Directory Management**: Create and remove directories
- **File Search**: Find files by pattern with regex support
- **Text Search**: Search for text content within files
- **Recursive Operations**: Support for recursive directory operations

### Technical Features
- **MCP Client**: Full Model Context Protocol client implementation
- **Ollama Integration**: Native Ollama API support with tool calling
- **Configuration Management**: Automatic config loading with fallbacks
- **Error Handling**: Comprehensive error handling and logging
- **Concurrent Execution**: Efficient tool execution with context support

## Architecture

The project follows a clean architecture pattern:

- **`main.go`**: Entry point that orchestrates MCP connections and Ollama interactions
- **`lib/mcp`**: Configuration management and YAML parsing
- **`lib/tool`**: Tool abstraction layer for consistent tool execution
- **`pkg/mcp`**: MCP client implementation with multi-server support
- **`pkg/ollama`**: Ollama client wrapper with tool integration
- **`cmd/filesystem`**: Standalone filesystem MCP server implementation

## Example Interactions

```bash
# File operations
./ttobot "Create a README.md file with project description"
./ttobot "List all Go files in the project"
./ttobot "Find all functions named 'main' in the codebase"

# Directory operations  
./ttobot "Create a new directory called 'docs'"
./ttobot "What's the current directory structure?"

# Search operations
./ttobot "Search for 'func main' in all Go files"
./ttobot "Find all YAML files recursively"
```

## Dependencies

- **[Ollama](https://github.com/ollama/ollama)**: Local LLM inference server
- **[Model Context Protocol Go SDK](https://github.com/modelcontextprotocol/go-sdk)**: MCP implementation
- **[YAML v3](https://gopkg.in/yaml.v3)**: Configuration file parsing

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for:

- New MCP server implementations
- Additional tool integrations
- Performance improvements
- Documentation enhancements
- Bug fixes

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Contact

For questions or support, contact [snowmerak](https://github.com/snowmerak).
