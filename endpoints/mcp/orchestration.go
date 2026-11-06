package mcp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

var errNotFound = errors.New("not found")

var platformNames = map[models.PlatformType]string{
	models.PlatformTypeiOS:          "ios",
	models.PlatformTypeAndroid:      "android",
	models.PlatformTypeMac:          "mac",
	models.PlatformTypeWindows:      "windows",
	models.PlatformTypeLinux:        "linux",
	models.PlatformTypeWeb:          "web",
	models.PlatformTypeEditor:       "unity_editor",
	models.PlatformTypeiOSSimulator: "ios_simulator",
}

var resultNames = map[models.TestResultState]string{
	models.TestResultOpen:     "open",
	models.TestResultUnstable: "unstable",
	models.TestResultFailed:   "failed",
	models.TestResultSuccess:  "success",
}

var testTypeNames = map[models.TestType]string{
	models.TestTypeUnity:    "unity",
	models.TestTypeCocos:    "cocos",
	models.TestTypeSerenity: "serenity",
	models.TestTypeScenario: "scenario",
}

// --- inputs ---

type DeviceArgs struct {
	DeviceID uint `json:"deviceId" jsonschema:"device id from list_devices"`
}

type ListDevicesArgs struct {
	Platform string `json:"platform,omitempty" jsonschema:"optional filter: android, ios, ios_simulator, mac, windows, linux, web, unity_editor"`
}

type AppArgs struct {
	AppID uint `json:"appId" jsonschema:"app id from list_apps"`
}

type ListBundlesArgs struct {
	AppID uint `json:"appId" jsonschema:"app id from list_apps"`
	Limit int  `json:"limit,omitempty" jsonschema:"max bundles, newest first (default 20)"`
}

type UploadBundleArgs struct {
	AppID uint   `json:"appId" jsonschema:"app id from list_apps"`
	URL   string `json:"url" jsonschema:"http(s) URL of the .apk, .ipa or app archive"`
}

type RunTestArgs struct {
	TestID   uint   `json:"testId" jsonschema:"test id from list_tests"`
	BundleID uint   `json:"bundleId,omitempty" jsonschema:"app bundle id from list_bundles; required except for web and editor apps"`
	StartURL string `json:"startUrl,omitempty" jsonschema:"start URL for web apps"`
	Params   string `json:"params,omitempty" jsonschema:"environment parameters as key=value;key2=value2"`
}

type RunArgs struct {
	RunID uint `json:"runId" jsonschema:"test run id"`
}

type WaitArgs struct {
	RunID          uint `json:"runId" jsonschema:"test run id"`
	TimeoutSeconds int  `json:"timeoutSeconds,omitempty" jsonschema:"how long to wait (default 300, max 1800)"`
}

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

// --- outputs ---

type deviceInfo struct {
	ID           uint              `json:"id"`
	Identifier   string            `json:"identifier"`
	Name         string            `json:"name"`
	Alias        string            `json:"alias,omitempty"`
	Platform     string            `json:"platform"`
	OS           string            `json:"os"`
	OSVersion    string            `json:"osVersion"`
	Node         string            `json:"node,omitempty"`
	Online       bool              `json:"online"`
	Status       string            `json:"status"`
	Locked       bool              `json:"locked"`
	LockedBy     string            `json:"lockedBy,omitempty"`
	AppConnected bool              `json:"appConnected"`
	Parameter    map[string]string `json:"parameter,omitempty"`
}

func (t *tools) deviceInfo(d *models.Device, details bool) deviceInfo {
	info := deviceInfo{
		ID: d.ID, Identifier: d.DeviceIdentifier, Name: d.Name, Alias: d.Alias,
		Platform: platformNames[d.PlatformType], OS: d.OS, OSVersion: d.OSVersion,
		Status: device.StateToString(device.StateUnknown),
	}
	if d.Node != nil {
		info.Node = d.Node.Name
	}
	if dev, _ := t.Devices.GetDevice(d.DeviceIdentifier); dev != nil {
		info.Online = true
		info.Status = device.StateToString(dev.DeviceState())
		info.Locked = dev.IsLocked()
		info.AppConnected = dev.IsAppConnected()
		if l, ok := t.Control.Holder(dev.DeviceID()); ok {
			info.LockedBy = l.Owner
		}
		if details {
			info.Parameter = dev.DeviceParameter()
		}
	} else if d.NodeID > 0 {
		info.Status = device.StateToString(device.StateNodeDisconnected)
	}
	return info
}

func (t *tools) findDevice(id uint) (*models.Device, error) {
	return t.Backend.FindDevice(t.project, id)
}

func (t *tools) findApp(id uint) (*models.App, error) {
	var a models.App
	err := t.DB.First(&a, "id = ? AND project_id = ?", id, t.project.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: app %d", errNotFound, id)
	}
	return &a, err
}

// projectRuns joins a test_runs query to the project.
func (t *tools) projectRuns() *gorm.DB {
	return t.DB.Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("apps.project_id = ?", t.project.ID)
}

func (t *tools) findProtocol(id uint) (*models.TestProtocol, error) {
	var p models.TestProtocol
	err := t.DB.Preload("Device").
		Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("apps.project_id = ?", t.project.ID).
		First(&p, "test_protocols.id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: protocol %d", errNotFound, id)
	}
	return &p, err
}

func (t *tools) registerOrchestration(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "list_devices", Description: "List the devices of the project with status, lock holder and whether an app is connected.", Annotations: readOnly}, t.listDevices)
	mcp.AddTool(s, &mcp.Tool{Name: "get_device", Description: "Get one device with its parameters.", Annotations: readOnly}, t.getDevice)
	mcp.AddTool(s, &mcp.Tool{Name: "list_apps", Description: "List the apps of the project.", Annotations: readOnly}, t.listApps)
	mcp.AddTool(s, &mcp.Tool{Name: "list_bundles", Description: "List uploaded builds (bundles) of an app, newest first.", Annotations: readOnly}, t.listBundles)
	mcp.AddTool(s, &mcp.Tool{Name: "upload_bundle", Description: "Download a build from a URL and add it as a bundle of the app.", Annotations: writeSafe}, t.uploadBundle)
	mcp.AddTool(s, &mcp.Tool{Name: "list_tests", Description: "List the tests of an app with their type and last run.", Annotations: readOnly}, t.listTests)
	mcp.AddTool(s, &mcp.Tool{Name: "run_test", Description: "Start a test run on the devices of its configuration. Returns the run id immediately; use wait_for_test_run and get_test_run.", Annotations: writeSafe}, t.runTest)
	mcp.AddTool(s, &mcp.Tool{Name: "get_test_run", Description: "Status and summary of a test run: state, results per protocol (one per device and test), and the last run log lines.", Annotations: readOnly}, t.getTestRun)
	mcp.AddTool(s, &mcp.Tool{Name: "wait_for_test_run", Description: "Wait until a test run finished or the timeout passed. Sends progress notifications.", Annotations: readOnly}, t.waitForTestRun)
	mcp.AddTool(s, &mcp.Tool{Name: "get_protocol", Description: "Details of one protocol: result, duration, errors and a performance summary.", Annotations: readOnly}, t.getProtocol)
	mcp.AddTool(s, &mcp.Tool{Name: "get_protocol_logs", Description: "Log entries of a protocol, filtered and paged.", Annotations: readOnly}, t.getProtocolLogs)
	mcp.AddTool(s, &mcp.Tool{Name: "cancel_test_run", Description: "Cancel a running test run.", Annotations: destructive}, t.cancelTestRun)
	mcp.AddTool(s, &mcp.Tool{Name: "unlock_device", Description: "Force unlock a device. This aborts a test run or control session that holds it.", Annotations: destructive}, t.unlockDevice)
}

func (t *tools) listDevices(ctx context.Context, req *mcp.CallToolRequest, in ListDevicesArgs) (*mcp.CallToolResult, any, error) {
	platform := strings.ToLower(in.Platform)
	if platform != "" {
		known := false
		for _, name := range platformNames {
			known = known || name == platform
		}
		if !known {
			return nil, nil, fmt.Errorf("unknown platform %q", in.Platform)
		}
	}
	devices, err := t.Backend.ProjectDevices(t.project)
	if err != nil {
		return nil, nil, err
	}
	result := make([]deviceInfo, 0, len(devices))
	for i := range devices {
		if platform != "" && platformNames[devices[i].PlatformType] != platform {
			continue
		}
		result = append(result, t.deviceInfo(&devices[i], false))
	}
	return jsonResult(result)
}

func (t *tools) getDevice(ctx context.Context, req *mcp.CallToolRequest, in DeviceArgs) (*mcp.CallToolResult, any, error) {
	d, err := t.findDevice(in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(t.deviceInfo(d, true))
}

func (t *tools) listApps(ctx context.Context, req *mcp.CallToolRequest, in struct{}) (*mcp.CallToolResult, any, error) {
	var list []models.App
	if err := t.DB.Where("project_id = ?", t.project.ID).Order("id").Find(&list).Error; err != nil {
		return nil, nil, err
	}
	type appInfo struct {
		ID         uint   `json:"id"`
		Name       string `json:"name"`
		Identifier string `json:"identifier"`
		Platform   string `json:"platform"`
	}
	result := make([]appInfo, 0, len(list))
	for _, a := range list {
		result = append(result, appInfo{a.ID, a.Name, a.Identifier, platformNames[a.Platform]})
	}
	return jsonResult(result)
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
}

func toBundleInfo(b *models.AppBinary) bundleInfo {
	return bundleInfo{b.ID, b.Name, b.Version, b.Identifier, b.Platform, b.Size, b.Tags, b.CreatedAt}
}

func (t *tools) listBundles(ctx context.Context, req *mcp.CallToolRequest, in ListBundlesArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.findApp(in.AppID); err != nil {
		return nil, nil, err
	}
	limit := in.Limit
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	var list []models.AppBinary
	if err := t.DB.Where("app_id = ?", in.AppID).Order("id desc").Limit(limit).Find(&list).Error; err != nil {
		return nil, nil, err
	}
	result := make([]bundleInfo, 0, len(list))
	for i := range list {
		result = append(result, toBundleInfo(&list[i]))
	}
	return jsonResult(result)
}

func (t *tools) uploadBundle(ctx context.Context, req *mcp.CallToolRequest, in UploadBundleArgs) (*mcp.CallToolResult, any, error) {
	application, err := t.findApp(in.AppID)
	if err != nil {
		return nil, nil, err
	}
	u, err := url.Parse(in.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, nil, fmt.Errorf("url must be http or https")
	}
	name := path.Base(u.Path)
	if name == "" || name == "/" || name == "." {
		return nil, nil, fmt.Errorf("url has no file name")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("download failed: %s", resp.Status)
	}

	if err := os.MkdirAll(apps.AppStoragePath, os.ModePerm); err != nil {
		return nil, nil, err
	}
	f, err := os.CreateTemp(apps.AppStoragePath, "upload-*-"+filepath.Base(name))
	if err != nil {
		return nil, nil, err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, nil, fmt.Errorf("download failed: %w", err)
	}
	f.Close()
	target := filepath.Join(apps.AppStoragePath, filepath.Base(name))
	if err := os.Rename(f.Name(), target); err != nil {
		os.Remove(f.Name())
		return nil, nil, err
	}

	bundle, err := t.Backend.StoreBundle(application, target)
	if bundle != nil && err != nil {
		// identical build exists already; report it instead of failing
		return jsonResult(map[string]any{"exists": true, "bundle": toBundleInfo(bundle)})
	}
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(toBundleInfo(bundle))
}

func (t *tools) listTests(ctx context.Context, req *mcp.CallToolRequest, in AppArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.findApp(in.AppID); err != nil {
		return nil, nil, err
	}
	var tests []models.Test
	if err := t.DB.Where("app_id = ?", in.AppID).Preload("TestConfig").Preload("TestConfig.Devices").Order("id").Find(&tests).Error; err != nil {
		return nil, nil, err
	}
	type testInfo struct {
		ID         uint   `json:"id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		AllDevices bool   `json:"allDevices"`
		DeviceIDs  []uint `json:"deviceIds,omitempty"`
		LastRunID  uint   `json:"lastRunId,omitempty"`
	}
	result := make([]testInfo, 0, len(tests))
	for _, test := range tests {
		info := testInfo{ID: test.ID, Name: test.Name, Type: testTypeNames[test.TestConfig.Type], AllDevices: test.TestConfig.AllDevices}
		for _, d := range test.TestConfig.Devices {
			info.DeviceIDs = append(info.DeviceIDs, d.DeviceID)
		}
		t.DB.Table("test_runs").Where("test_id = ? AND deleted_at IS NULL", test.ID).Select("MAX(id)").Scan(&info.LastRunID)
		result = append(result, info)
	}
	return jsonResult(result)
}

func (t *tools) runTest(ctx context.Context, req *mcp.CallToolRequest, in RunTestArgs) (*mcp.CallToolResult, any, error) {
	run, err := t.Backend.StartTestRun(t.project, in.TestID, in.BundleID, in.StartURL, in.Params)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]any{"runId": run.ID, "sessionId": run.SessionID, "state": runs.RunRunning})
}

type protocolInfo struct {
	ID       uint     `json:"id"`
	ParentID *uint    `json:"parentId,omitempty"`
	Name     string   `json:"name"`
	DeviceID *uint    `json:"deviceId,omitempty"`
	Device   string   `json:"device,omitempty"`
	Result   string   `json:"result"`
	Duration string   `json:"duration"`
	AvgFPS   *float64 `json:"avgFps,omitempty"`
	AvgMEM   *float64 `json:"avgMem,omitempty"`
	AvgCPU   *float64 `json:"avgCpu,omitempty"`
}

func duration(p *models.TestProtocol) string {
	if p.EndedAt == nil {
		return "running"
	}
	return p.EndedAt.Sub(p.StartedAt).Round(time.Second).String()
}

func nonZero(v float64) *float64 {
	if v == 0 || v != v { // skip zero and NaN
		return nil
	}
	return &v
}

func toProtocolInfo(p *models.TestProtocol) protocolInfo {
	info := protocolInfo{
		ID: p.ID, ParentID: p.ParentTestProtocolID, Name: p.TestName, DeviceID: p.DeviceID,
		Result: resultNames[p.TestResult], Duration: duration(p),
		AvgFPS: nonZero(p.AvgFPS), AvgMEM: nonZero(p.AvgMEM), AvgCPU: nonZero(p.AvgCPU),
	}
	if p.Device != nil {
		info.Device = p.Device.Name
	}
	return info
}

func (t *tools) getTestRun(ctx context.Context, req *mcp.CallToolRequest, in RunArgs) (*mcp.CallToolResult, any, error) {
	summary, err := t.runSummary(in.RunID)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(summary)
}

func (t *tools) runSummary(runID uint) (map[string]any, error) {
	var run models.TestRun
	err := t.projectRuns().Preload("Test").Preload("AppBinary").
		Preload("Protocols", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		Preload("Protocols.Device").
		First(&run, "test_runs.id = ?", runID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: test run %d", errNotFound, runID)
	}
	if err != nil {
		return nil, err
	}

	state, result := t.Runs.Status(run.ID)
	counts := map[string]int{}
	protocols := make([]protocolInfo, 0, len(run.Protocols))
	for i := range run.Protocols {
		p := &run.Protocols[i]
		counts[resultNames[p.TestResult]]++
		protocols = append(protocols, toProtocolInfo(p))
	}

	var logs []models.TestRunLogEntry
	t.DB.Where("test_run_id = ?", run.ID).Order("id desc").Limit(20).Find(&logs)
	logLines := make([]string, 0, len(logs))
	for i := len(logs) - 1; i >= 0; i-- {
		logLines = append(logLines, fmt.Sprintf("%s [%s] %s", logs[i].CreatedAt.Format(time.RFC3339), logs[i].Level, logs[i].Log))
	}

	summary := map[string]any{
		"runId":        run.ID,
		"state":        state,
		"createdAt":    run.CreatedAt,
		"parameter":    run.Parameter,
		"resultCounts": counts,
		"protocols":    protocols,
		"log":          logLines,
	}
	if result != nil {
		summary["result"] = result
	}
	if run.Test != nil {
		summary["test"] = map[string]any{"id": run.Test.ID, "name": run.Test.Name}
	}
	if run.AppBinary != nil {
		summary["bundle"] = toBundleInfo(run.AppBinary)
	}
	return summary, nil
}

func (t *tools) waitForTestRun(ctx context.Context, req *mcp.CallToolRequest, in WaitArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.runSummary(in.RunID); err != nil {
		return nil, nil, err
	}
	timeout := time.Duration(in.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	if timeout > 30*time.Minute {
		timeout = 30 * time.Minute
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// progress: number of protocols that ended so far
	if token := req.Params.GetProgressToken(); token != nil && req.Session != nil {
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-waitCtx.Done():
					return
				case <-ticker.C:
					var total, ended int64
					t.DB.Model(&models.TestProtocol{}).Where("test_run_id = ?", in.RunID).Count(&total)
					t.DB.Model(&models.TestProtocol{}).Where("test_run_id = ? AND ended_at IS NOT NULL", in.RunID).Count(&ended)
					_ = req.Session.NotifyProgress(waitCtx, &mcp.ProgressNotificationParams{
						ProgressToken: token,
						Progress:      float64(ended),
						Total:         float64(total),
						Message:       fmt.Sprintf("%d of %d protocols finished", ended, total),
					})
				}
			}
		}()
	}

	state, _, err := t.Runs.Wait(waitCtx, in.RunID)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return nil, nil, err
	}
	summary, err := t.runSummary(in.RunID)
	if err != nil {
		return nil, nil, err
	}
	if state == runs.RunRunning {
		summary["note"] = fmt.Sprintf("still running after %v; call wait_for_test_run again", timeout)
	}
	return jsonResult(summary)
}

func (t *tools) getProtocol(ctx context.Context, req *mcp.CallToolRequest, in ProtocolArgs) (*mcp.CallToolResult, any, error) {
	p, err := t.findProtocol(in.ProtocolID)
	if err != nil {
		return nil, nil, err
	}
	var entryCount int64
	t.DB.Model(&models.ProtocolEntry{}).Where("test_protocol_id = ?", p.ID).Count(&entryCount)

	var errorsList []models.ProtocolEntry
	t.DB.Where("test_protocol_id = ? AND level IN ?", p.ID, []string{"error", "exception"}).Order("id").Limit(50).Find(&errorsList)
	errorLines := make([]string, 0, len(errorsList))
	for _, e := range errorsList {
		errorLines = append(errorLines, fmt.Sprintf("%s [%s/%s] %s", e.Timestamp.Format(time.RFC3339), e.Level, e.Source, e.Message))
	}

	var perf struct {
		Samples int64
		AvgFPS  float64
		MinFPS  float64
		AvgMEM  float64
		MaxMEM  float64
		AvgCPU  float64
	}
	t.DB.Model(&models.ProtocolPerformanceEntry{}).Where("test_protocol_id = ?", p.ID).
		Select("COUNT(*) AS samples, COALESCE(AVG(fps),0) AS avg_fps, COALESCE(MIN(fps),0) AS min_fps, COALESCE(AVG(mem),0) AS avg_mem, COALESCE(MAX(mem),0) AS max_mem, COALESCE(AVG(cpu),0) AS avg_cpu").
		Scan(&perf)

	var children []models.TestProtocol
	t.DB.Preload("Device").Where("parent_test_protocol_id = ?", p.ID).Order("id").Find(&children)
	childInfos := make([]protocolInfo, 0, len(children))
	for i := range children {
		childInfos = append(childInfos, toProtocolInfo(&children[i]))
	}

	return jsonResult(map[string]any{
		"protocol":    toProtocolInfo(p),
		"testRunId":   p.TestRunID,
		"startedAt":   p.StartedAt,
		"endedAt":     p.EndedAt,
		"entries":     entryCount,
		"errors":      errorLines,
		"performance": perf,
		"children":    childInfos,
	})
}

func (t *tools) getProtocolLogs(ctx context.Context, req *mcp.CallToolRequest, in ProtocolLogsArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.findProtocol(in.ProtocolID); err != nil {
		return nil, nil, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	query := t.DB.Model(&models.ProtocolEntry{}).Where("test_protocol_id = ?", in.ProtocolID)
	if in.Level != "" {
		query = query.Where("level = ?", strings.ToLower(in.Level))
	}
	if in.Contains != "" {
		query = query.Where("message LIKE ?", "%"+strings.ReplaceAll(in.Contains, "%", `\%`)+"%")
	}
	var total int64
	query.Count(&total)

	var entries []models.ProtocolEntry
	if err := query.Order("id").Offset(in.Offset).Limit(limit).Find(&entries).Error; err != nil {
		return nil, nil, err
	}
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		line := fmt.Sprintf("%s [%s/%s] %s", e.Timestamp.Format(time.RFC3339), e.Level, e.Source, e.Message)
		if e.Data != "" {
			line += " (data: " + e.Data + ")"
		}
		lines = append(lines, line)
	}
	return jsonResult(map[string]any{"total": total, "offset": in.Offset, "entries": lines})
}

func (t *tools) cancelTestRun(ctx context.Context, req *mcp.CallToolRequest, in RunArgs) (*mcp.CallToolResult, any, error) {
	if err := t.Backend.CancelTestRun(t.project, in.RunID); err != nil {
		return nil, nil, err
	}
	return textResult("test run %d cancelled", in.RunID)
}

func (t *tools) unlockDevice(ctx context.Context, req *mcp.CallToolRequest, in DeviceArgs) (*mcp.CallToolResult, any, error) {
	d, err := t.findDevice(in.DeviceID)
	if err != nil {
		return nil, nil, err
	}
	dev, _ := t.Devices.GetDevice(d.DeviceIdentifier)
	if dev == nil {
		return nil, nil, fmt.Errorf("device %d is not connected", in.DeviceID)
	}
	t.Control.ForceRelease(dev.DeviceID())
	if dev.IsLocked() {
		if err := dev.Unlock(); err != nil {
			return nil, nil, err
		}
	}
	return textResult("device %d unlocked", in.DeviceID)
}
