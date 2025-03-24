package utils

import (
	"encoding/json"
	"fmt"

	mcp "github.com/metoro-io/mcp-golang"
)

// CreateSuccessResponse creates a tool response from any struct
func CreateSuccessResponse(result interface{}) *mcp.ToolResponse {
	resultJSON, _ := json.Marshal(result)
	return mcp.NewToolResponse(mcp.NewTextContent(string(resultJSON)))
}

// CreateErrorResponse creates an error response with the given message
func CreateErrorResponse(message string) *mcp.ToolResponse {
	errorJSON := fmt.Sprintf(`{"success": false, "error": "%s"}`, message)
	return mcp.NewToolResponse(mcp.NewTextContent(errorJSON))
}

// CreateErrorResponseWithData creates an error response including the provided data
func CreateErrorResponseWithData(message string, data interface{}) *mcp.ToolResponse {
	// First create a map with error info
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}

	// Add all fields from data to the response
	// This requires that data is a struct or map
	if dataMap, ok := data.(map[string]interface{}); ok {
		for k, v := range dataMap {
			response[k] = v
		}
	} else {
		// If it's a struct, convert it to JSON and back to a map
		dataJSON, _ := json.Marshal(data)
		var dataMap map[string]interface{}
		json.Unmarshal(dataJSON, &dataMap)

		for k, v := range dataMap {
			if k != "success" && k != "error" { // Don't overwrite these
				response[k] = v
			}
		}
	}

	// Marshal the combined response
	resultJSON, _ := json.Marshal(response)
	return mcp.NewToolResponse(mcp.NewTextContent(string(resultJSON)))
}
