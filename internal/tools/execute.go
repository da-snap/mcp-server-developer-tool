package tools

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	mcp "github.com/metoro-io/mcp-golang"
	"mcp-server/internal/config"
	"mcp-server/internal/utils"
)

// ExecuteShellCommandArgs defines the arguments for the execute_shell_command tool
type ExecuteShellCommandArgs struct {
	Command    []string `json:"command" jsonschema:"required,description=The command to execute as an array of strings"`
	Timeout    int      `json:"timeout" jsonschema:"description=Maximum execution time in seconds"`
	WorkingDir *string  `json:"working_dir" jsonschema:"description=Working directory for command execution"`
}

// ExecuteShellCommandResult defines the result of the execute_shell_command tool
type ExecuteShellCommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Command  string `json:"command"`
	Success  bool   `json:"success"`
}

// ExecuteShellTool implements the execute_shell_command tool
type ExecuteShellTool struct {
	config *config.ServerConfig
}

// NewExecuteShellTool creates a new ExecuteShellTool instance
func NewExecuteShellTool() *ExecuteShellTool {
	return &ExecuteShellTool{}
}

// SetConfig sets the server configuration
func (t *ExecuteShellTool) SetConfig(cfg *config.ServerConfig) {
	t.config = cfg
}

// Name returns the tool name
func (t *ExecuteShellTool) Name() string {
	return "execute_shell_command"
}

// Description returns the tool description
func (t *ExecuteShellTool) Description() string {
	return "Execute a shell command and return the complete results including stdout, stderr, and exit code"
}

// Execute runs a shell command with the provided arguments
func (t *ExecuteShellTool) Execute(args ExecuteShellCommandArgs) (*mcp.ToolResponse, error) {
	// Set default timeout if not provided
	timeout := 60
	if args.Timeout > 0 {
		timeout = args.Timeout
	}

	if len(args.Command) == 0 {
		return utils.CreateErrorResponse("Empty command"), nil
	}

	// Check if the command is valid
	if !t.isCommandAllowed(args.Command[0]) {
		return t.createResponse(
			"",
			fmt.Sprintf("Command '%s' is not allowed for security reasons", args.Command[0]),
			-1,
			strings.Join(args.Command, " "),
			false,
		), nil
	}

	// Check working directory if provided
	if args.WorkingDir != nil && t.config != nil {
		allowed, err := t.config.IsPathAllowed(*args.WorkingDir)
		if err != nil || !allowed {
			errorMsg := "Working directory is not allowed by server configuration"
			if err != nil {
				errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
			}
			return t.createResponse("", errorMsg, -1, strings.Join(args.Command, " "), false), nil
		}
	}

	// Create the command
	cmd := exec.Command(args.Command[0], args.Command[1:]...)

	// Set working directory if provided
	if args.WorkingDir != nil {
		cmd.Dir = *args.WorkingDir
	}

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return t.createResponse("", fmt.Sprintf("Error creating stdout pipe: %v", err), -1, strings.Join(args.Command, " "), false), nil
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return t.createResponse("", fmt.Sprintf("Error creating stderr pipe: %v", err), -1, strings.Join(args.Command, " "), false), nil
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return t.createResponse("", fmt.Sprintf("Error starting command: %v", err), -1, strings.Join(args.Command, " "), false), nil
	}

	// Create a channel for command completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// Read stdout and stderr
	stdoutData, _ := io.ReadAll(stdout)
	stderrData, _ := io.ReadAll(stderr)

	// Wait for command to complete or timeout
	var exitCode int
	var success bool

	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		// Command timed out
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return t.createResponse(
			string(stdoutData),
			fmt.Sprintf("Command timed out after %d seconds\n%s", timeout, string(stderrData)),
			-1,
			strings.Join(args.Command, " "),
			false,
		), nil

	case err := <-done:
		// Command completed
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = -1
			}
			success = false
		} else {
			exitCode = 0
			success = true
		}
	}

	return t.createResponse(
		string(stdoutData),
		string(stderrData),
		exitCode,
		strings.Join(args.Command, " "),
		success,
	), nil
}

// isCommandAllowed checks if a command is allowed to be executed
func (t *ExecuteShellTool) isCommandAllowed(command string) bool {
	// Check if it's a path
	if filepath.IsAbs(command) || strings.Contains(command, "/") || strings.Contains(command, "\\") {
		// If it's a path and we have a config, check if it's in an allowed path
		if t.config != nil {
			allowed, _ := t.config.IsPathAllowed(command)
			return allowed
		}
	}

	// Allowed common utilities and binaries
	allowedCommands := map[string]bool{
		"ls": true, "find": true, "grep": true, "cat": true, "echo": true,
		"pwd": true, "cd": true, "mkdir": true, "rm": true, "cp": true, "mv": true,
		"touch": true, "head": true, "tail": true, "wc": true, "sort": true,
		"uniq": true, "cut": true, "tr": true, "sed": true, "awk": true,
		"ps": true, "top": true, "df": true, "du": true, "free": true,
		"which": true, "whereis": true, "whatis": true, "file": true,
		"zip": true, "unzip": true, "tar": true, "gzip": true, "gunzip": true,
		// Add more allowed commands as needed
	}

	// Default to not allowed if not in the whitelist
	return allowedCommands[command]
}

// createResponse creates a response for the execute_shell_command tool
func (t *ExecuteShellTool) createResponse(stdout, stderr string, exitCode int, command string, success bool) *mcp.ToolResponse {
	result := ExecuteShellCommandResult{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		Command:  command,
		Success:  success,
	}

	return utils.CreateSuccessResponse(result)
}
