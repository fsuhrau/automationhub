package api_test

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mcpCall calls a tool of the project's MCP endpoint on the real router.
func mcpCall(t *testing.T, cs *mcp.ClientSession, tool string, args map[string]any) (string, bool) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: tool, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", tool, err)
	}
	var out strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			out.WriteString(tc.Text)
		}
	}
	return out.String(), res.IsError
}

// The MCP tools read the same seeded project through the services the router
// builds, and stay in the project of their URL.
func TestMCPToolsRunOnTheServicesOfTheRouter(t *testing.T) {
	env := apitest.New(t)
	seedContract(env)
	srv := httptest.NewServer(env.Router)
	t.Cleanup(srv.Close)
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).
		Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: srv.URL + "/api/proj/mcp"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cs.Close() })

	tests := []struct {
		name  string
		tool  string
		args  map[string]any
		want  string
		fails bool
	}{
		{"apps", "list_apps", nil, `"name": "App"`, false},
		{"bundles", "list_bundles", map[string]any{"appId": 1}, `"name": "app.apk"`, false},
		{"tests", "list_tests", map[string]any{"appId": 1}, `"lastRunId": 2`, false},
		{"run", "get_test_run", map[string]any{"runId": 1}, `"name": "Suite Case"`, false},
		{"protocol", "get_protocol", map[string]any{"protocolId": 2}, `"result": "success"`, false},
		{"protocol logs", "get_protocol_logs", map[string]any{"protocolId": 1}, `"total": 0`, false},
		{"devices", "list_devices", nil, `"name": "Pixel"`, false},
		{"device", "get_device", map[string]any{"deviceId": 1}, `"status": "unknown"`, false},
		{"unknown app", "list_tests", map[string]any{"appId": 9}, "app 9 not found", true},
		{"unlock a device that is not connected", "unlock_device", map[string]any{"deviceId": 1}, "dev-1 is not connected", true},
		{"run without a connected device", "run_test", map[string]any{"testId": 1, "bundleId": 1}, "no device available", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, failed := mcpCall(t, cs, tc.tool, tc.args)

			if failed != tc.fails || !strings.Contains(got, tc.want) {
				t.Fatalf("failed %v, result %s; want %q", failed, got, tc.want)
			}
		})
	}
}
