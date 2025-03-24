package tools

import (
	"mcp-server/internal/config"
)

// Tool defines the interface that all MCP tools must implement
type Tool interface {
	// Name returns the tool name as it will be exposed to the MCP client
	Name() string

	// Description returns a detailed description of the tool
	Description() string

	// Execute is implemented by each tool to run its specific functionality
	// The actual signature will differ for each tool based on its argument type,
	// but reflection is used to call it correctly
	// Execute(args SomeArgsType) (*mcp.ToolResponse, error)
}

// ConfigAware is an interface that tools can implement to receive server configuration
type ConfigAware interface {
	// SetConfig sets the server configuration for the tool
	SetConfig(cfg *config.ServerConfig)
}
