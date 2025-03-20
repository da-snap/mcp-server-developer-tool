package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"mcp-server/internal/server"
	"mcp-server/internal/tools"
)

func main() {
	// Set up logging to a file so we don't interfere with stdio communication
	logFile, err := os.OpenFile("mcp-server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
	}

	// Create MCP server
	mcpServer, err := server.NewServer()
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	// Register all tools
	registerTools(mcpServer)

	// Set up signal handling for graceful shutdown
	setupSignalHandling(mcpServer)

	// Start the server
	log.Printf("Starting MCP server with stdio transport...")
	if err := mcpServer.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait forever (handled by signal handling)
	select {}
}

// registerTools registers all tools with the server
func registerTools(mcpServer *server.Server) {
	// Create tool instances
	executeShellTool := tools.NewExecuteShellTool()
	showFileTool := tools.NewShowFileTool()
	searchFileTool := tools.NewSearchFileTool()
	writeFileTool := tools.NewWriteFileTool()

	// Register tools with server
	if err := mcpServer.RegisterTool(executeShellTool); err != nil {
		log.Fatalf("Failed to register execute_shell_command tool: %v", err)
	}

	if err := mcpServer.RegisterTool(showFileTool); err != nil {
		log.Fatalf("Failed to register show_file tool: %v", err)
	}

	if err := mcpServer.RegisterTool(searchFileTool); err != nil {
		log.Fatalf("Failed to register search_in_file tool: %v", err)
	}

	if err := mcpServer.RegisterTool(writeFileTool); err != nil {
		log.Fatalf("Failed to register write_file tool: %v", err)
	}
}

// setupSignalHandling sets up handlers for OS signals
func setupSignalHandling(mcpServer *server.Server) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		log.Printf("Received signal %v, shutting down...", sig)
		mcpServer.Stop()
		os.Exit(0)
	}()
}
