package tools

import (
	"fmt"
	"os"
	"strings"

	mcp "github.com/metoro-io/mcp-golang"
	"mcp-server/internal/config"
	"mcp-server/internal/utils"
)

// ShowFileArgs defines the arguments for the show_file tool
type ShowFileArgs struct {
	FilePath  string `json:"file_path" jsonschema:"required,description=Path to the file to display"`
	StartLine int    `json:"start_line" jsonschema:"description=Line number to start from (1-based indexing)"`
	NumLines  *int   `json:"num_lines" jsonschema:"description=Number of lines to display (defaults to all lines)"`
}

// ShowFileResult defines the result of the show_file tool
type ShowFileResult struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Content    string `json:"content"`
	LinesShown int    `json:"lines_shown"`
	TotalLines int    `json:"total_lines"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
}

// ShowFileTool implements the show_file tool
type ShowFileTool struct {
	config *config.ServerConfig
}

// NewShowFileTool creates a new ShowFileTool instance
func NewShowFileTool() *ShowFileTool {
	return &ShowFileTool{}
}

// SetConfig sets the server configuration
func (t *ShowFileTool) SetConfig(cfg *config.ServerConfig) {
	t.config = cfg
}

// Name returns the tool name
func (t *ShowFileTool) Name() string {
	return "show_file"
}

// Description returns the tool description
func (t *ShowFileTool) Description() string {
	return "Show contents of a file with options to display specific line ranges"
}

// Execute shows file contents with the provided arguments
func (t *ShowFileTool) Execute(args ShowFileArgs) (*mcp.ToolResponse, error) {
	// Check if path is allowed by configuration
	if t.config != nil {
		allowed, err := t.config.IsPathAllowed(args.FilePath)
		if err != nil || !allowed {
			errorMsg := "Access to this file path is not allowed by server configuration"
			if err != nil {
				errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
			}
			result := ShowFileResult{
				Success:    false,
				Error:      errorMsg,
				Content:    "",
				LinesShown: 0,
				TotalLines: 0,
			}
			return utils.CreateSuccessResponse(result), nil
		}
	}

	// Check if file exists
	fileInfo, err := os.Stat(args.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			result := ShowFileResult{
				Success:    false,
				Error:      fmt.Sprintf("File %s does not exist", args.FilePath),
				Content:    "",
				LinesShown: 0,
				TotalLines: 0,
			}
			return utils.CreateSuccessResponse(result), nil
		}
		return utils.CreateErrorResponse(fmt.Sprintf("Error checking file: %v", err)), nil
	}

	// Don't read directories
	if fileInfo.IsDir() {
		result := ShowFileResult{
			Success:    false,
			Error:      fmt.Sprintf("%s is a directory, not a file", args.FilePath),
			Content:    "",
			LinesShown: 0,
			TotalLines: 0,
		}
		return utils.CreateSuccessResponse(result), nil
	}

	// Read file content
	content, err := os.ReadFile(args.FilePath)
	if err != nil {
		result := ShowFileResult{
			Success:    false,
			Error:      fmt.Sprintf("Error reading file: %v", err),
			Content:    "",
			LinesShown: 0,
			TotalLines: 0,
		}
		return utils.CreateSuccessResponse(result), nil
	}

	// Split into lines
	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	// Ensure start line is valid
	startLine := args.StartLine
	if startLine < 1 {
		startLine = 1
	}

	// Check if start line is beyond file length
	if startLine > totalLines {
		result := ShowFileResult{
			Success:    false,
			Error:      fmt.Sprintf("Start line %d is beyond the file length (%d lines)", startLine, totalLines),
			Content:    "",
			LinesShown: 0,
			TotalLines: totalLines,
		}
		return utils.CreateSuccessResponse(result), nil
	}

	// Convert to 0-based index
	startIndex := startLine - 1

	// Determine end index
	endIndex := totalLines
	if args.NumLines != nil {
		endIndex = startIndex + *args.NumLines
		if endIndex > totalLines {
			endIndex = totalLines
		}
	}

	// Extract requested lines
	selectedLines := lines[startIndex:endIndex]
	selectedContent := strings.Join(selectedLines, "\n")

	// Create result
	result := ShowFileResult{
		Success:    true,
		Content:    selectedContent,
		LinesShown: len(selectedLines),
		TotalLines: totalLines,
		StartLine:  startLine,
		EndLine:    startIndex + len(selectedLines) + 1,
	}

	return utils.CreateSuccessResponse(result), nil
}
