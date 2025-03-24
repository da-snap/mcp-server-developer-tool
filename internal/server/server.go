package server

import (
	"log"
	"reflect"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"mcp-server/internal/config"
	"mcp-server/internal/tools"
)

// Server wraps the MCP server functionality
type Server struct {
	mcpServer *mcp.Server
	done      chan struct{}
	Config    *config.ServerConfig
}

// NewServer creates a new MCP server instance with the default configuration
func NewServer() (*Server, error) {
	return NewServerWithConfig(config.NewConfigFromEnv())
}

// NewServerWithConfig creates a new MCP server instance with the provided configuration
func NewServerWithConfig(cfg *config.ServerConfig) (*Server, error) {
	// Create a stdio transport
	transport := stdio.NewStdioServerTransport()

	// Create a new MCP server
	mcpServer := mcp.NewServer(transport)

	return &Server{
		mcpServer: mcpServer,
		done:      make(chan struct{}),
		Config:    cfg,
	}, nil
}

// RegisterTool registers a tool with the MCP server
func (s *Server) RegisterTool(tool tools.Tool) error {
	// Get tool name, description, and handler function
	name := tool.Name()
	description := tool.Description()

	log.Printf("Registering tool: %s", name)

	// Inject configuration into the tool if it implements ConfigAware
	if configAware, ok := tool.(tools.ConfigAware); ok {
		configAware.SetConfig(s.Config)
	}

	// Get the Execute method using reflection
	toolValue := reflect.ValueOf(tool)
	executeMethod := toolValue.MethodByName("Execute")

	// Register the tool with the MCP server
	err := s.mcpServer.RegisterTool(name, description, executeMethod.Interface())
	if err != nil {
		return err
	}

	return nil
}

// Start begins the MCP server
func (s *Server) Start() error {
	log.Printf("Server starting with configuration: Allowed paths: %v", s.Config.AllowedPaths)

	go func() {
		if err := s.mcpServer.Serve(); err != nil {
			log.Printf("Error in server: %v", err)
			close(s.done)
		}
	}()

	return nil
}

// Stop gracefully stops the MCP server
func (s *Server) Stop() {
	log.Println("Stopping MCP server...")
	// Currently the mcp-golang library doesn't have a built-in way to stop the server,
	// but we could implement one if needed by closing connections, etc.
	close(s.done)
}
