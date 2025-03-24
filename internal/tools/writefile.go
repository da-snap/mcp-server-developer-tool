package tools

import (
	"fmt"
	"os"
	"path/filepath"

	mcp "github.com/metoro-io/mcp-golang"
	"mcp-server/internal/config"
	"mcp-server/internal/utils"
)

// WriteFileArgs defines the arguments for the write_file tool
type WriteFileArgs struct {
	FilePath string `json:"file_path" jsonschema:"required,description=Path to the file to write"`
	Content  string `json:"content" jsonschema:"required,description=Text content to write to the file"`
	Mode     string `json:"mode" jsonschema:"description=Write mode to use: 'w' (overwrite) or 'a' (append)"`
}

// WriteFileResult defines the result of the write_file tool
type WriteFileResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// WriteFileTool implements the write_file tool
type WriteFileTool struct {
	config *config.ServerConfig
}

// NewWriteFileTool creates a new WriteFileTool instance
func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{}
}

// SetConfig sets the server configuration
func (t *WriteFileTool) SetConfig(cfg *config.ServerConfig) {
	t.config = cfg
}

// Name returns the tool name
func (t *WriteFileTool) Name() string {
	return "write_file"
}

// Description returns the tool description
func (t *WriteFileTool) Description() string {
	return "Write content to a file with options to append or overwrite existing content"
}

// Execute writes to a file with the provided arguments
func (t *WriteFileTool) Execute(args WriteFileArgs) (*mcp.ToolResponse, error) {
	// Check if path is allowed by configuration
	if t.config != nil {
		allowed, err := t.config.IsPathAllowed(args.FilePath)
		if err != nil || !allowed {
			errorMsg := "Access to this file path is not allowed by server configuration"
			if err != nil {
				errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
			}
			result := WriteFileResult{
				Success: false,
				Error:   errorMsg,
			}
			return utils.CreateSuccessResponse(result), nil
		}
	}

	// Determine file mode
	fileMode := os.O_WRONLY | os.O_CREATE
	if args.Mode == "a" {
		fileMode |= os.O_APPEND
	} else {
		fileMode |= os.O_TRUNC
	}

	// Create parent directories if they don't exist
	dir := filepath.Dir(args.FilePath)
	if dir != "" {
		// Check if the parent directory's path is allowed too
		if t.config != nil {
			allowed, err := t.config.IsPathAllowed(dir)
			if err != nil || !allowed {
				errorMsg := "Access to the parent directory is not allowed by server configuration"
				if err != nil {
					errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
				}
				result := WriteFileResult{
					Success: false,
					Error:   errorMsg,
				}
				return utils.CreateSuccessResponse(result), nil
			}
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			result := WriteFileResult{
				Success: false,
				Error:   fmt.Sprintf("Error creating directories: %v", err),
			}
			return utils.CreateSuccessResponse(result), nil
		}
	}

	// Open file
	file, err := os.OpenFile(args.FilePath, fileMode, 0644)
	if err != nil {
		result := WriteFileResult{
			Success: false,
			Error:   fmt.Sprintf("Error opening file: %v", err),
		}
		return utils.CreateSuccessResponse(result), nil
	}
	defer file.Close()

	// Write content
	_, err = file.WriteString(args.Content)
	if err != nil {
		result := WriteFileResult{
			Success: false,
			Error:   fmt.Sprintf("Error writing to file: %v", err),
		}
		return utils.CreateSuccessResponse(result), nil
	}

	// Create result
	result := WriteFileResult{
		Success: true,
	}

	return utils.CreateSuccessResponse(result), nil
}
