package tools

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	mcp "github.com/metoro-io/mcp-golang"
	"mcp-server/internal/config"
	"mcp-server/internal/utils"
)

// SearchInFileArgs defines the arguments for the search_in_file tool
type SearchInFileArgs struct {
	FilePath      string `json:"file_path" jsonschema:"required,description=Path to the file to search"`
	Pattern       string `json:"pattern" jsonschema:"required,description=Regular expression pattern to search for"`
	CaseSensitive bool   `json:"case_sensitive" jsonschema:"description=Whether the search should be case-sensitive"`
	MaxMatches    int    `json:"max_matches" jsonschema:"description=Maximum number of matches to return (use -1 for all matches)"`
}

// MatchResult represents a single match result
type MatchResult struct {
	LineNumber int    `json:"line_number"`
	Content    string `json:"content"`
}

// SearchInFileResult defines the result of the search_in_file tool
type SearchInFileResult struct {
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	Matches    []MatchResult `json:"matches"`
	MatchCount int           `json:"match_count"`
	Truncated  bool          `json:"truncated"`
}

// SearchFileTool implements the search_in_file tool
type SearchFileTool struct {
	config *config.ServerConfig
}

// NewSearchFileTool creates a new SearchFileTool instance
func NewSearchFileTool() *SearchFileTool {
	return &SearchFileTool{}
}

// SetConfig sets the server configuration
func (t *SearchFileTool) SetConfig(cfg *config.ServerConfig) {
	t.config = cfg
}

// Name returns the tool name
func (t *SearchFileTool) Name() string {
	return "search_in_file"
}

// Description returns the tool description
func (t *SearchFileTool) Description() string {
	return "Search for patterns in a file using regular expressions"
}

// Execute searches in a file with the provided arguments
func (t *SearchFileTool) Execute(args SearchInFileArgs) (*mcp.ToolResponse, error) {
	// Check if path is allowed by configuration
	if t.config != nil {
		allowed, err := t.config.IsPathAllowed(args.FilePath)
		if err != nil || !allowed {
			errorMsg := "Access to this file path is not allowed by server configuration"
			if err != nil {
				errorMsg = fmt.Sprintf("%s: %v", errorMsg, err)
			}
			result := SearchInFileResult{
				Success:    false,
				Error:      errorMsg,
				Matches:    []MatchResult{},
				MatchCount: 0,
			}
			return utils.CreateSuccessResponse(result), nil
		}
	}

	// Check if file exists
	_, err := os.Stat(args.FilePath)
	if os.IsNotExist(err) {
		result := SearchInFileResult{
			Success:    false,
			Error:      fmt.Sprintf("File %s does not exist", args.FilePath),
			Matches:    []MatchResult{},
			MatchCount: 0,
		}
		return utils.CreateSuccessResponse(result), nil
	} else if err != nil {
		return utils.CreateErrorResponse(fmt.Sprintf("Error checking file: %v", err)), nil
	}

	// Open the file
	file, err := os.Open(args.FilePath)
	if err != nil {
		result := SearchInFileResult{
			Success:    false,
			Error:      fmt.Sprintf("Error opening file: %v", err),
			Matches:    []MatchResult{},
			MatchCount: 0,
		}
		return utils.CreateSuccessResponse(result), nil
	}
	defer file.Close()

	// Prepare regex options
	var regexPattern string
	if args.CaseSensitive {
		regexPattern = args.Pattern
	} else {
		regexPattern = "(?i)" + args.Pattern
	}

	// Compile regex
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		result := SearchInFileResult{
			Success:    false,
			Error:      fmt.Sprintf("Invalid regular expression: %v", err),
			Matches:    []MatchResult{},
			MatchCount: 0,
		}
		return utils.CreateSuccessResponse(result), nil
	}

	// Search file
	matches := []MatchResult{}
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if regex.MatchString(line) {
			matches = append(matches, MatchResult{
				LineNumber: lineNum,
				Content:    line,
			})

			// Check if we've reached max matches
			if args.MaxMatches > 0 && len(matches) >= args.MaxMatches {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		result := SearchInFileResult{
			Success:    false,
			Error:      fmt.Sprintf("Error reading file: %v", err),
			Matches:    []MatchResult{},
			MatchCount: 0,
		}
		return utils.CreateSuccessResponse(result), nil
	}

	// Create result
	result := SearchInFileResult{
		Success:    true,
		Matches:    matches,
		MatchCount: len(matches),
		Truncated:  args.MaxMatches > 0 && len(matches) >= args.MaxMatches,
	}

	return utils.CreateSuccessResponse(result), nil
}
