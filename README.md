# Modular MCP Server

This is a Go implementation of the Model Context Protocol (MCP) server using the `github.com/metoro-io/mcp-golang` library, restructured in a modular way.

## Project Structure

```
mcp-server/
├── cmd/
│   └── mcp-server/
│       └── main.go           # Entry point
├── internal/
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

## Adding New Tools

To add a new tool:

1. Create a new file in the `internal/tools` directory
2. Implement the `Tool` interface
3. Register the tool in `cmd/mcp-server/main.go`

Example:

```go
// internal/tools/newtool.go
package tools

import (
    "github.com/metoro-io/mcp-golang"
    "mcp-server/internal/utils"
)

type NewToolArgs struct {
    // Tool arguments
}

type NewTool struct{}

func NewNewTool() *NewTool {
    return &NewTool{}
}

func (t *NewTool) Name() string {
    return "new_tool"
}

func (t *NewTool) Description() string {
    return "Description of the new tool"
}

func (t *NewTool) Execute(args NewToolArgs) (*mcp.ToolResponse, error) {
    // Tool implementation
    return utils.CreateSuccessResponse(result), nil
}
```

## Testing

```bash
go test ./...
```
