package mcp

import (
	"context"
	"fmt"
	"time"

	runsvc "github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ProtocolArgs struct {
	ProtocolID uint `json:"protocolId" jsonschema:"protocol id from get_test_run"`
}

type ProtocolLogsArgs struct {
	ProtocolID uint   `json:"protocolId" jsonschema:"protocol id from get_test_run"`
	Level      string `json:"level,omitempty" jsonschema:"only entries of this level, e.g. error, exception, log"`
	Contains   string `json:"contains,omitempty" jsonschema:"only entries whose message contains this text"`
	Offset     int    `json:"offset,omitempty" jsonschema:"skip this many matching entries"`
	Limit      int    `json:"limit,omitempty" jsonschema:"max entries (default 100, max 500)"`
}

// perfInfo is the performance summary of a protocol, named as the fields of
// the service's summary.
type perfInfo struct {
	Samples int64
	AvgFPS  float64
	MinFPS  float64
	AvgMEM  float64
	MaxMEM  float64
	AvgCPU  float64
}

func (t *tools) registerProtocols(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "get_protocol", Description: "Details of one protocol: result, duration, errors and a performance summary.", Annotations: readOnly()}, t.getProtocol)
	mcp.AddTool(s, &mcp.Tool{Name: "get_protocol_screenshot", Description: "A screenshot taken during a test, as an image. Give the file from get_protocol (screenshots) or get_protocol_logs (data: ...); without a file the protocol's last screenshot is returned.", Annotations: readOnly()}, t.getProtocolScreenshot)
	mcp.AddTool(s, &mcp.Tool{Name: "get_protocol_logs", Description: "Log entries of a protocol, filtered and paged.", Annotations: readOnly()}, t.getProtocolLogs)
}

// entryLine is a protocol entry as one log line.
func entryLine(e *models.ProtocolEntry) string {
	return fmt.Sprintf("%s [%s/%s] %s", e.Timestamp.Format(time.RFC3339), e.Level, e.Source, e.Message)
}

func (t *tools) getProtocol(ctx context.Context, req *mcp.CallToolRequest, in ProtocolArgs) (*mcp.CallToolResult, any, error) {
	detail, err := t.Protocols.ProtocolDetail(ctx, t.project.ID, in.ProtocolID)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(protocolView(detail))
}

// protocolView is the get_protocol result.
func protocolView(d *runsvc.ProtocolDetail) map[string]any {
	errorLines := make([]string, 0, len(d.Errors))
	for i := range d.Errors {
		errorLines = append(errorLines, entryLine(&d.Errors[i]))
	}
	return map[string]any{
		"protocol":    toProtocolInfo(d.Protocol),
		"screenshots": screenshotInfos(d.Data),
		"testRunId":   d.Protocol.TestRunID,
		"startedAt":   d.Protocol.StartedAt,
		"endedAt":     d.Protocol.EndedAt,
		"entries":     d.Entries,
		"errors":      errorLines,
		"performance": perfInfo(d.Performance),
		"children":    toProtocolInfos(d.Children),
	}
}

func (t *tools) getProtocolLogs(ctx context.Context, req *mcp.CallToolRequest, in ProtocolLogsArgs) (*mcp.CallToolResult, any, error) {
	page, err := t.Protocols.Logs(ctx, runsvc.LogQuery{
		ProjectID: t.project.ID, ProtocolID: in.ProtocolID, Level: in.Level,
		Contains: in.Contains, Offset: in.Offset, Limit: in.Limit,
	})
	if err != nil {
		return nil, nil, err
	}
	lines := make([]string, 0, len(page.Entries))
	for i := range page.Entries {
		lines = append(lines, logEntryLine(&page.Entries[i]))
	}
	return jsonResult(map[string]any{"total": page.Total, "offset": page.Offset, "entries": lines})
}

// logEntryLine is an entry line with the data the entry refers to.
func logEntryLine(e *models.ProtocolEntry) string {
	line := entryLine(e)
	if e.Data != "" {
		line += " (data: " + e.Data + ")"
	}
	return line
}
