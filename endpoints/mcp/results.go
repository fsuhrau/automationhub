package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func ptr[T any](v T) *T { return &v }

// The annotations tell a client whether a tool only reads, writes safely or
// destroys.
func readOnly() *mcp.ToolAnnotations { return &mcp.ToolAnnotations{ReadOnlyHint: true} }

func writeSafe() *mcp.ToolAnnotations { return &mcp.ToolAnnotations{DestructiveHint: ptr(false)} }

func destructive() *mcp.ToolAnnotations { return &mcp.ToolAnnotations{DestructiveHint: ptr(true)} }

// jsonResult returns v as pretty JSON text.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, nil, nil
}

func textResult(format string, args ...any) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}}}, nil, nil
}

// jpegResult is a JPEG image with a note that says what it shows.
func jpegResult(jpg []byte, note string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.ImageContent{Data: jpg, MIMEType: "image/jpeg"},
		&mcp.TextContent{Text: note},
	}}, nil, nil
}
