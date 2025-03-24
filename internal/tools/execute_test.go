package tools

import (
	"encoding/json"
	"strings"
	"testing"

	mcp "github.com/metoro-io/mcp-golang"
)

func TestExecuteShellTool_Name(t *testing.T) {
	tool := NewExecuteShellTool()
	expected := "execute_shell_command"

	if tool.Name() != expected {
		t.Errorf("Expected tool name to be %s, got %s", expected, tool.Name())
	}
}

func TestExecuteShellTool_Description(t *testing.T) {
	tool := NewExecuteShellTool()
	if tool.Description() == "" {
		t.Error("Expected tool description to be non-empty")
	}
}

func TestExecuteShellTool_Execute_EmptyCommand(t *testing.T) {
	tool := NewExecuteShellTool()
	args := ExecuteShellCommandArgs{
		Command: []string{},
	}

	resp, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to be non-nil")
	}

	// Verify error is returned in the response
	if len(resp.Content) == 0 {
		t.Fatal("Expected Content to be non-empty")
	}

	// Use the mcp package here to fix "not used" error
	_ = mcp.ContentTypeText // Just to use the import

	content := resp.Content[0].TextContent.Text
	if !strings.Contains(content, "Empty command") {
		t.Errorf("Expected response to contain error message, got: %s", content)
	}
}

func TestExecuteShellTool_Execute_Echo(t *testing.T) {
	tool := NewExecuteShellTool()
	testMessage := "hello world"

	// Skip test on Windows as the command would be different
	if isWindows() {
		t.Skip("Skipping test on Windows")
	}

	args := ExecuteShellCommandArgs{
		Command: []string{"echo", testMessage},
	}

	resp, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response to be non-nil")
	}

	if len(resp.Content) == 0 {
		t.Fatal("Expected Content to be non-empty")
	}

	// Parse the response
	var result ExecuteShellCommandResult
	if err := json.Unmarshal([]byte(resp.Content[0].TextContent.Text), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Check success flag
	if !result.Success {
		t.Errorf("Expected command to succeed, got success=%v with stderr: %s",
			result.Success, result.Stderr)
	}

	// Check stdout contains our test message
	if !strings.Contains(result.Stdout, testMessage) {
		t.Errorf("Expected stdout to contain '%s', got: %s", testMessage, result.Stdout)
	}

	// Check exit code
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

// Helper function to check if running on Windows
func isWindows() bool {
	return false // For this example, just return false
	// In a real implementation:
	// return runtime.GOOS == "windows"
}
