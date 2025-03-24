# Modular MCP Server

This is a Go implementation of the Model Context Protocol (MCP) server using the `github.com/metoro-io/mcp-golang` library, restructured in a modular way.

## Project Structure

```
mcp-server/
├── cmd/
│   └── mcp-server/
│       └── main.go           # Entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Server configuration
│   ├── server/
│   │   ├── server.go         # MCP server implementation
│   │   └── server_test.go    # Server tests
│   ├── tools/
│   │   ├── tool.go           # Tool interface
│   │   ├── execute.go        # Execute shell command tool
│   │   ├── showfile.go       # Show file tool
│   │   ├── searchfile.go     # Search in file tool
│   │   └── writefile.go      # Write file tool
│   └── utils/
│       └── response.go       # Common response utilities
├── go.mod
├── go.sum
└── README.md
```

## Building and Running

```bash
# Build the server
cd cmd/mcp-server
go build -o mcp-server

# Run the server
./mcp-server
```

## Security Features

The server now includes a path restriction system to limit file operations to specified directories.

### Configuring Allowed Paths

You can configure allowed paths in several ways:

#### 1. Using Command-line Flags

```bash
# Allow operations only in specific directories
./mcp-server --paths=/home/user/safe:/tmp/workspace

# Explicitly deny specific paths even if within allowed paths
./mcp-server --paths=/home/user --deny-paths=/home/user/.ssh:/home/user/credentials
```

#### 2. Using Environment Variables

```bash
# Set allowed paths
export MCP_ALLOWED_PATHS=/home/user/safe:/tmp/workspace

# Set denied paths
export MCP_DENIED_PATHS=/home/user/.ssh:/home/user/credentials

# Run the server
./mcp-server
```

#### 3. Programmatically

You can also create a custom configuration programmatically:

```go
cfg := config.DefaultConfig()
cfg.AddAllowedPath("/path/to/allow")
cfg.AddDeniedPath("/path/to/deny")

server, err := server.NewServerWithConfig(cfg)
```

### Default Behavior

- If no paths are specified, the server defaults to allowing only the current working directory.
- Common sensitive directories like `.git` and `.env` are automatically added to the deny list.

### Shell Command Security

For the `execute_shell_command` tool:

- Commands are restricted to a whitelist of common utilities
- Custom executable paths are checked against the allowed paths configuration
- Working directories must be within allowed paths

## Adding New Tools

To add a new tool:

1. Create a new file in the `internal/tools` directory
2. Implement the `Tool` interface
3. Optionally implement the `ConfigAware` interface if your tool needs access to server configuration
4. Register the tool in `cmd/mcp-server/main.go`

Example:

```go
// internal/tools/newtool.go
package tools

import (
    "github.com/metoro-io/mcp-golang"
    "mcp-server/internal/config"
    "mcp-server/internal/utils"
)

type NewToolArgs struct {
    // Tool arguments
}

type NewTool struct{
    config *config.ServerConfig
}

func NewNewTool() *NewTool {
    return &NewTool{}
}

// Implement ConfigAware interface
func (t *NewTool) SetConfig(cfg *config.ServerConfig) {
    t.config = cfg
}

func (t *NewTool) Name() string {
    return "new_tool"
}

func (t *NewTool) Description() string {
    return "Description of the new tool"
}

func (t *NewTool) Execute(args NewToolArgs) (*mcp.ToolResponse, error) {
    // Access configuration if needed
    if t.config != nil {
        // Use configuration for security checks
    }

    // Tool implementation
    return utils.CreateSuccessResponse(result), nil
}
```

## Testing

```bash
go test ./...
```
