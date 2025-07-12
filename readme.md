# ttobot

## Overview

`ttobot` is a Go-based project designed to provide tools and libraries for working with MCP (Model Context Protocol) and Ollama clients. The repository is organized for modularity and extensibility, supporting command-line utilities and reusable packages.

## Project Structure

```
ttobot/
├── cmd/                # Command-line tools
│   └── filesystem/     # Filesystem-related CLI
│       └── main.go
├── lib/                # Libraries for MCP and tools
│   └── mcp/
│       └── config.go
├── tool/               # Utility tools
│   └── tool.go
├── pkg/                # Packages for MCP and Ollama clients
│   └── mcp/
│       ├── client.go
│       └── convert.go
│   └── ollama/
│       └── client.go
├── go.mod              # Go module definition
├── go.sum              # Go dependencies checksum
├── main.go             # Main entry point
├── mcp.yaml            # MCP configuration
└── readme.md           # Project documentation
```

## Getting Started

### Prerequisites
- Go 1.20 or newer
- (Optional) MCP and Ollama services for client integration

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

### Usage
Build and run the main application:

```zsh
go run main.go
```

Or build CLI tools:

```zsh
go build -o ttobot-cmd ./cmd/filesystem/main.go
./ttobot-cmd
```

## Features
- MCP client and configuration utilities
- Ollama client integration
- Filesystem command-line tool
- Modular package structure for easy extension

## Contributing
Contributions are welcome! Please open issues or submit pull requests for improvements and bug fixes.

## License
This project is licensed under the MIT License.

## Contact
For questions or support, contact [snowmerak](https://github.com/snowmerak).
