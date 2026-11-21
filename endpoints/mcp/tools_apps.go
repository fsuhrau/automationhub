package mcp

import (
	"context"
	"time"

	appsvc "github.com/fsuhrau/automationhub/service/app"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultBundleLimit = 20
	maxBundleLimit     = 200
)

type AppArgs struct {
	AppID uint `json:"appId" jsonschema:"app id from list_apps"`
}

type ListBundlesArgs struct {
	AppID uint `json:"appId" jsonschema:"app id from list_apps"`
	Limit int  `json:"limit,omitempty" jsonschema:"max bundles, newest first (default 20)"`
}

type appInfo struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
	Platform   string `json:"platform"`
}

type bundleInfo struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Version    string    `json:"version"`
	Identifier string    `json:"identifier"`
	Platform   string    `json:"platform"`
	Size       int       `json:"size"`
	Tags       string    `json:"tags,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	// Archived bundles have no file: they can't be installed or run
	Archived bool `json:"archived,omitempty"`
}

type testInfo struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	AllDevices bool   `json:"allDevices"`
	DeviceIDs  []uint `json:"deviceIds,omitempty"`
	LastRunID  uint   `json:"lastRunId,omitempty"`
}

func (t *tools) registerApps(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "list_apps", Description: "List the apps of the project.", Annotations: readOnly()}, t.listApps)
	mcp.AddTool(s, &mcp.Tool{Name: "list_bundles", Description: "List uploaded builds (bundles) of an app, newest first.", Annotations: readOnly()}, t.listBundles)
	mcp.AddTool(s, &mcp.Tool{Name: "list_tests", Description: "List the tests of an app with their type and last run.", Annotations: readOnly()}, t.listTests)
}

func toBundleInfo(b *models.AppBinary) bundleInfo {
	return bundleInfo{b.ID, b.Name, b.Version, b.Identifier, b.Platform, b.Size, b.Tags, b.CreatedAt, b.Archived()}
}

// bundleLimit is the number of bundles to list: the default for a missing or
// excessive limit.
func bundleLimit(asked int) int {
	if asked <= 0 || asked > maxBundleLimit {
		return defaultBundleLimit
	}
	return asked
}

func toTestInfo(e *testsvc.ListEntry) testInfo {
	info := testInfo{ID: e.ID, Name: e.Name, Type: testTypeNames()[e.TestConfig.Type], AllDevices: e.TestConfig.AllDevices}
	for _, d := range e.TestConfig.Devices {
		info.DeviceIDs = append(info.DeviceIDs, d.DeviceID)
	}
	if len(e.RecentRuns) > 0 {
		info.LastRunID = e.RecentRuns[0].ID
	}
	return info
}

// app finds an app of the project.
func (t *tools) app(ctx context.Context, id uint) (*models.App, error) {
	return t.Apps.App(ctx, appsvc.Scope{ProjectID: t.project.ID, AppID: id})
}

func (t *tools) listApps(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	list, err := t.Apps.Apps(ctx, t.project.ID)
	if err != nil {
		return nil, nil, err
	}
	result := make([]appInfo, 0, len(list))
	for _, a := range list {
		result = append(result, appInfo{a.ID, a.Name, a.Identifier, platformNames()[a.Platform]})
	}
	return jsonResult(result)
}

func (t *tools) listBundles(ctx context.Context, req *mcp.CallToolRequest, in ListBundlesArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.app(ctx, in.AppID); err != nil {
		return nil, nil, err
	}
	list, err := t.Apps.Bundles(ctx, appsvc.Scope{ProjectID: t.project.ID, AppID: in.AppID})
	if err != nil {
		return nil, nil, err
	}
	list = list[:min(bundleLimit(in.Limit), len(list))]
	result := make([]bundleInfo, 0, len(list))
	for i := range list {
		result = append(result, toBundleInfo(&list[i].AppBinary))
	}
	return jsonResult(result)
}

func (t *tools) listTests(ctx context.Context, req *mcp.CallToolRequest, in AppArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.app(ctx, in.AppID); err != nil {
		return nil, nil, err
	}
	list, err := t.Tests.List(ctx, testsvc.Scope{ProjectID: t.project.ID, AppID: in.AppID})
	if err != nil {
		return nil, nil, err
	}
	result := make([]testInfo, 0, len(list))
	for i := range list {
		result = append(result, toTestInfo(&list[i]))
	}
	return jsonResult(result)
}
