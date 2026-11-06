package api

import (
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	AVG_HISTORY_LIMIT = 20
)

// recentRunsPerTest is how many runs the tests list shows per test
const recentRunsPerTest = 10

// runSummary is one run in the tests list: when it ran and how it went
type runSummary struct {
	ID        uint          `json:"id"`
	CreatedAt time.Time     `json:"createdAt"`
	State     runs.RunState `json:"state"`
	// running, passed, failed, unstable, interrupted, not_run or empty;
	// the same rules as the run page
	Status   string `json:"status"`
	Passed   int    `json:"passed"`
	Failed   int    `json:"failed"`
	Unstable int    `json:"unstable"`
	Open     int    `json:"open"`
}

type testListEntry struct {
	models.Test
	// the latest runs, newest first
	RecentRuns []runSummary `json:"recentRuns"`
}

// summarizeRun is the run status of the run page: a running run is running;
// otherwise failures, then unstable tests, then tests that never finished
// decide; a run without results did not run when its log has errors
func summarizeRun(state runs.RunState, passed, failed, unstable, open, logErrors int) string {
	switch {
	case state == runs.RunRunning:
		return "running"
	case passed+failed+unstable+open == 0 && logErrors > 0:
		return "not_run"
	case passed+failed+unstable+open == 0:
		return "empty"
	case failed > 0:
		return "failed"
	case unstable > 0:
		return "unstable"
	case open > 0:
		return "interrupted"
	default:
		return "passed"
	}
}

// summarizeRuns gives each run its status and result counts, in the order
// of runs. It counts the tests themselves; a runner protocol wraps one.
func (s *Service) summarizeRuns(testRuns []models.TestRun) ([]runSummary, error) {
	summaries := make([]runSummary, len(testRuns))
	if len(testRuns) == 0 {
		return summaries, nil
	}
	ids := make([]uint, len(testRuns))
	for i, r := range testRuns {
		ids[i] = r.ID
	}
	var results []struct {
		TestRunID  uint
		TestResult models.TestResultState
		N          int
	}
	if err := s.db.Model(&models.TestProtocol{}).Select("test_run_id, test_result, count(*) as n").
		Where("test_run_id IN ?", ids).
		Where("NOT EXISTS (SELECT 1 FROM test_protocols c WHERE c.parent_test_protocol_id = test_protocols.id AND c.deleted_at IS NULL)").
		Group("test_run_id, test_result").Scan(&results).Error; err != nil {
		return nil, err
	}
	var logErrors []struct {
		TestRunID uint
		N         int
	}
	if err := s.db.Model(&models.TestRunLogEntry{}).Select("test_run_id, count(*) as n").
		Where("test_run_id IN ? AND level = ?", ids, "error").
		Group("test_run_id").Scan(&logErrors).Error; err != nil {
		return nil, err
	}
	type counts struct{ passed, failed, unstable, open, logErrors int }
	byRun := map[uint]*counts{}
	get := func(id uint) *counts {
		if byRun[id] == nil {
			byRun[id] = &counts{}
		}
		return byRun[id]
	}
	for _, r := range results {
		n := get(r.TestRunID)
		switch r.TestResult {
		case models.TestResultSuccess:
			n.passed += r.N
		case models.TestResultFailed:
			n.failed += r.N
		case models.TestResultUnstable:
			n.unstable += r.N
		default:
			n.open += r.N
		}
	}
	for _, e := range logErrors {
		get(e.TestRunID).logErrors = e.N
	}
	for i, r := range testRuns {
		n := get(r.ID)
		state := s.runState(r.ID)
		summaries[i] = runSummary{
			ID: r.ID, CreatedAt: r.CreatedAt, State: state,
			Passed: n.passed, Failed: n.failed, Unstable: n.unstable, Open: n.open,
			Status: summarizeRun(state, n.passed, n.failed, n.unstable, n.open, n.logErrors),
		}
	}
	return summaries, nil
}

// getTests lists the tests of an app with their latest runs. It used to
// preload every run of every test, which the list never showed.
func (s *Service) getTests(c *gin.Context, project *models.Project, application *models.App) {
	var tests []models.Test
	if err := s.db.Where("app_id = ?", application.ID).Preload("TestConfig").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").Preload("TestConfig.Devices").Order("id").Find(&tests).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	list := make([]testListEntry, len(tests))
	for i := range tests {
		list[i].Test = tests[i]
		var recent []models.TestRun
		if err := s.db.Where("test_id = ?", tests[i].ID).Order("id desc").Limit(recentRunsPerTest).Find(&recent).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		summaries, err := s.summarizeRuns(recent)
		if err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		list[i].RecentRuns = summaries
	}
	c.JSON(http.StatusOK, list)
}

func (s *Service) newTest(c *gin.Context, project *models.Project, application *models.App) {

	type TestFunc struct {
		Assembly string `json:"assembly"`
		Class    string `json:"class"`
		Method   string `json:"method"`
	}
	type Request struct {
		Name                  string                       `json:"name"`
		TestType              models.TestType              `json:"type"`
		ExecutionType         models.ExecutionType         `json:"executionType"`
		PlatformType          models.PlatformType          `json:"platformType"`
		UnityTestCategoryType models.UnityTestCategoryType `json:"unityTestCategoryType"`
		UnitySelectedTests    []TestFunc                   `json:"unitySelectedTests"`
		AllDevices            bool                         `json:"allDevices"`
		SelectedDevices       []uint                       `json:"selectedDevices"`
		Categories            []string                     `json:"categories"`
	}

	var request Request
	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	name := strings.TrimSpace(request.Name)
	if len(name) == 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("missing name"))
		return
	}

	switch request.TestType {
	case models.TestTypeUnity:
		if request.UnityTestCategoryType == models.SelectedTestsOnly && len(request.UnitySelectedTests) == 0 {
			s.error(c, http.StatusBadRequest, fmt.Errorf("missing tests"))
			return
		}
		if request.UnityTestCategoryType == models.AllOfCategory && len(request.Categories) == 0 {
			s.error(c, http.StatusBadRequest, fmt.Errorf("missing categories"))
			return
		}
	default:
		s.error(c, http.StatusBadRequest, fmt.Errorf("unsupported Test Type"))
		return
	}

	if request.AllDevices == false && len(request.SelectedDevices) == 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("unsupported Test Type"))
		return
	}

	tx := s.db.Begin()
	defer func() {
		tx.Rollback()
	}()
	test := models.Test{
		AppID: application.ID,
		Name:  request.Name,
	}
	if err := tx.Create(&test).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	config := models.TestConfig{
		TestID:        test.ID,
		Type:          request.TestType,
		AllDevices:    request.AllDevices,
		ExecutionType: request.ExecutionType,
	}
	if err := tx.Create(&config).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	switch request.TestType {
	case models.TestTypeUnity:
		unityConfig := models.TestConfigUnity{
			TestConfigID:          config.ID,
			UnityTestCategoryType: request.UnityTestCategoryType,
		}
		if len(request.Categories) > 0 {
			unityConfig.Categories = strings.Join(request.Categories, ",")
		}
		if err := tx.Create(&unityConfig).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		for _, r := range request.UnitySelectedTests {

			function := models.UnityTestFunction{
				TestConfigUnityID: unityConfig.ID,
				Assembly:          r.Assembly,
				Class:             r.Class,
				Method:            r.Method,
			}
			if err := tx.Create(&function).Error; err != nil {
				s.error(c, http.StatusInternalServerError, err)
				return
			}
		}
	}

	for _, d := range request.SelectedDevices {
		dev := models.TestConfigDevice{
			TestConfigID: config.ID,
			DeviceID:     d,
		}
		if err := tx.Create(&dev).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusCreated, test)
}

func (s *Service) getTest(c *gin.Context, project *models.Project, application *models.App) {
	testId := c.Param("test_id")

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Devices.Device").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, "id = ? AND app_id = ?", testId, application.ID).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, test)
}

// testUpdate is what the test edit page saves: the name, how and where the
// test runs, and for Unity tests what it runs
type testUpdate struct {
	Name                  string                       `json:"name"`
	ExecutionType         models.ExecutionType         `json:"executionType"`
	AllDevices            bool                         `json:"allDevices"`
	Devices               []uint                       `json:"devices"`
	UnityTestCategoryType models.UnityTestCategoryType `json:"unityTestCategoryType"`
	Categories            string                       `json:"categories"`
	TestFunctions         []models.UnityTestFunction   `json:"testFunctions"`
}

// validate tidies the update (trimmed name and categories, no repeated
// devices or functions) and says what keeps the test from running
func (u *testUpdate) validate(unity bool) error {
	u.Name = strings.TrimSpace(u.Name)
	if u.Name == "" {
		return fmt.Errorf("the test needs a name")
	}
	if u.ExecutionType != models.ConcurrentExecutionType && u.ExecutionType != models.SimultaneouslyExecutionType {
		return fmt.Errorf("unknown execution type %d", u.ExecutionType)
	}
	seen := map[uint]bool{}
	devices := u.Devices[:0:0]
	for _, d := range u.Devices {
		if !seen[d] {
			seen[d] = true
			devices = append(devices, d)
		}
	}
	u.Devices = devices
	if u.AllDevices {
		u.Devices = nil
	} else if len(u.Devices) == 0 {
		return fmt.Errorf("pick at least one device, or run on all devices")
	}
	if !unity {
		return nil
	}
	u.Categories = normalizeTags(u.Categories)
	type key struct{ assembly, class, method string }
	known := map[key]bool{}
	functions := u.TestFunctions[:0:0]
	for _, f := range u.TestFunctions {
		k := key{strings.TrimSpace(f.Assembly), strings.TrimSpace(f.Class), strings.TrimSpace(f.Method)}
		if k.method == "" || known[k] {
			continue
		}
		known[k] = true
		functions = append(functions, models.UnityTestFunction{Assembly: k.assembly, Class: k.class, Method: k.method})
	}
	u.TestFunctions = functions
	switch u.UnityTestCategoryType {
	case models.AllTest:
	case models.AllOfCategory:
		if u.Categories == "" {
			return fmt.Errorf("name at least one category")
		}
	case models.SelectedTestsOnly:
		if len(u.TestFunctions) == 0 {
			return fmt.Errorf("pick at least one test")
		}
	default:
		return fmt.Errorf("unknown Unity test selection %d", u.UnityTestCategoryType)
	}
	return nil
}

// updateTest saves the test edit page in one transaction: the devices and
// the Unity test functions are replaced by the submitted ones. The test
// must belong to the app and the devices to the project.
func (s *Service) updateTest(c *gin.Context, project *models.Project, application *models.App) {
	var req testUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	load := func() (*models.Test, error) {
		var test models.Test
		err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Devices.Device").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").
			First(&test, "id = ? AND app_id = ?", c.Param("test_id"), application.ID).Error
		return &test, err
	}
	test, err := load()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.error(c, http.StatusNotFound, fmt.Errorf("test %s not found", c.Param("test_id")))
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return
	}
	unity := test.TestConfig.Type == models.TestTypeUnity && test.TestConfig.Unity != nil
	if err := req.validate(unity); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	if len(req.Devices) > 0 {
		var visible int64
		if err := s.projectDevices(project).Model(&models.Device{}).Where("id IN ?", req.Devices).Count(&visible).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		if int(visible) != len(req.Devices) {
			s.error(c, http.StatusBadRequest, fmt.Errorf("some of the picked devices don't exist"))
			return
		}
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Test{}).Where("id = ?", test.ID).Update("name", req.Name).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.TestConfig{}).Where("id = ?", test.TestConfig.ID).
			Updates(map[string]interface{}{"execution_type": req.ExecutionType, "all_devices": req.AllDevices}).Error; err != nil {
			return err
		}
		if err := tx.Where("test_config_id = ?", test.TestConfig.ID).Delete(&models.TestConfigDevice{}).Error; err != nil {
			return err
		}
		for _, d := range req.Devices {
			if err := tx.Create(&models.TestConfigDevice{TestConfigID: test.TestConfig.ID, DeviceID: d}).Error; err != nil {
				return err
			}
		}
		if !unity {
			return nil
		}
		if err := tx.Model(&models.TestConfigUnity{}).Where("id = ?", test.TestConfig.Unity.ID).
			Updates(map[string]interface{}{"unity_test_category_type": req.UnityTestCategoryType, "categories": req.Categories}).Error; err != nil {
			return err
		}
		if err := tx.Where("test_config_unity_id = ?", test.TestConfig.Unity.ID).Delete(&models.UnityTestFunction{}).Error; err != nil {
			return err
		}
		for _, f := range req.TestFunctions {
			f.TestConfigUnityID = test.TestConfig.Unity.ID
			if err := tx.Create(&f).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	saved, err := load()
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, saved)
}

func extractParams(param string) map[string]string {
	var env map[string]string
	env = make(map[string]string)

	params := strings.Split(param, ";")
	for _, p := range params {
		kv := strings.Split(p, "=")
		if len(kv) > 1 {
			env[kv[0]] = kv[1]
		}
	}
	return env
}

type RunTestRequest struct {
	AppBinaryID int    `json:"appBinaryId"`
	StartURL    string `json:"startUrl"`
	Params      string `json:"params"`
}

func (s *Service) runTest(c *gin.Context, project *models.Project, application *models.App) {
	testID, err := strconv.ParseUint(c.Param("test_id"), 10, 64)
	if err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("invalid test id"))
		return
	}
	var req RunTestRequest
	if err := c.Bind(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	run, err := s.StartTestRun(project, uint(testID), uint(req.AppBinaryID), req.StartURL, req.Params)
	if err != nil {
		s.error(c, statusForError(err), err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (s *Service) cancelTestRun(c *gin.Context, project *models.Project, application *models.App) {
	runID, err := strconv.ParseUint(c.Param("run_id"), 10, 64)
	if err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("invalid run id"))
		return
	}
	if err := s.CancelTestRun(project, uint(runID)); err != nil {
		s.error(c, statusForError(err), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrBundleExists):
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

const (
	testRunsPageDefault = 25
	testRunsPageMax     = 100
)

type bundleRef struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// testRunListEntry is a run on the test page: its summary, what it ran
// with, and when its last test ended
type testRunListEntry struct {
	runSummary
	Parameter string     `json:"parameter"`
	Bundle    *bundleRef `json:"bundle"`
	EndedAt   *time.Time `json:"endedAt"`
}

type testRunsPage struct {
	Runs []testRunListEntry `json:"runs"`
	// more older runs exist; ask again with before = the last id
	More bool `json:"more"`
}

// getTestRuns pages through the runs of a test, newest first: limit runs
// (default 25, at most 100) older than the run id before, if given.
func (s *Service) getTestRuns(c *gin.Context, project *models.Project, application *models.App) {
	var test models.Test
	if err := s.db.First(&test, "id = ? AND app_id = ?", c.Param("test_id"), application.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.error(c, http.StatusNotFound, fmt.Errorf("test %s not found", c.Param("test_id")))
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return
	}
	limit := testRunsPageDefault
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = min(v, testRunsPageMax)
	}
	query := s.db.Where("test_id = ?", test.ID)
	if before, err := strconv.Atoi(c.Query("before")); err == nil && before > 0 {
		query = query.Where("id < ?", before)
	}
	var testRuns []models.TestRun
	if err := query.Preload("AppBinary").Order("id desc").Limit(limit + 1).Find(&testRuns).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	page := testRunsPage{More: len(testRuns) > limit}
	if page.More {
		testRuns = testRuns[:limit]
	}
	summaries, err := s.summarizeRuns(testRuns)
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	ids := make([]uint, len(testRuns))
	for i, r := range testRuns {
		ids[i] = r.ID
	}
	// the latest end of a run's protocols; read as columns and compared
	// here, since an aggregate of a time column comes back as text on SQLite
	var ends []models.TestProtocol
	if len(ids) > 0 {
		if err := s.db.Select("test_run_id", "ended_at").Where("test_run_id IN ? AND ended_at IS NOT NULL", ids).
			Find(&ends).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
	}
	endOf := map[uint]*time.Time{}
	for i := range ends {
		e := ends[i].EndedAt
		if e != nil && (endOf[ends[i].TestRunID] == nil || e.After(*endOf[ends[i].TestRunID])) {
			endOf[ends[i].TestRunID] = e
		}
	}

	page.Runs = make([]testRunListEntry, len(testRuns))
	for i, r := range testRuns {
		entry := testRunListEntry{runSummary: summaries[i], Parameter: r.Parameter, EndedAt: endOf[r.ID]}
		if r.AppBinary != nil {
			entry.Bundle = &bundleRef{ID: r.AppBinary.ID, Name: r.AppBinary.Name, Version: r.AppBinary.Version}
		}
		page.Runs[i] = entry
	}
	c.JSON(http.StatusOK, page)
}

type Response struct {
	TestRun   models.TestRun `json:"testRun"`
	PrevRunId uint           `json:"prevRunId"`
	NextRunId uint           `json:"nextRunId"`
	// State says whether the run still executes. The run model has no status
	// of its own: "running" and "finished" come from the run tracker, and
	// "not_running" is a run this master didn't start (e.g. before a restart).
	State runs.RunState `json:"state"`
}

func (s *Service) runState(runID uint) runs.RunState {
	if s.runTracker == nil {
		return runs.RunUnknown
	}
	state, _ := s.runTracker.Status(runID)
	return state
}

func (s *Service) getLastTestRun(c *gin.Context, project *models.Project, application *models.App) {

	testId := c.Param("test_id")

	var resp Response
	if err := s.db.Preload("DeviceStatus").Preload("DeviceStatus.Device").Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("AppBinary").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").Where("test_id = ?", testId).Order("id desc").First(&resp.TestRun).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	// get prev
	s.db.Table("test_runs").Where("test_id = ? and id < ?", testId, resp.TestRun.ID).Order("created_at desc").Limit(1).Select("id").Scan(&resp.PrevRunId)

	// get next
	s.db.Table("test_runs").Where("test_id = ? and id > ?", testId, resp.TestRun.ID).Order("created_at asc").Limit(1).Select("id").Scan(&resp.NextRunId)

	for i := range resp.TestRun.Protocols {
		defaultNaN(&resp.TestRun.Protocols[i])
	}
	resp.State = s.runState(resp.TestRun.ID)

	c.JSON(http.StatusOK, resp)
}

func (s *Service) getTestRun(c *gin.Context, project *models.Project, application *models.App) {
	testId := c.Param("test_id")
	runId := c.Param("run_id")

	var resp Response
	if err := s.db.Preload("DeviceStatus").Preload("DeviceStatus.Device").Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("AppBinary").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").First(&resp.TestRun, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	for i := range resp.TestRun.DeviceStatus {
		var histStatus []models.TestRunDeviceStatus
		if err := s.db.Where("device_id = ? and test_run_id in (select tr.id from test_runs tr where tr.id < ? order by tr.id desc limit ?)", resp.TestRun.DeviceStatus[i].DeviceID, resp.TestRun.DeviceStatus[i].TestRunID, AVG_HISTORY_LIMIT).Find(&histStatus).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		for _, t := range histStatus {
			resp.TestRun.DeviceStatus[i].HistAvgStartupTime += t.StartupTime
		}
		if len(histStatus) > 0 {
			resp.TestRun.DeviceStatus[i].HistAvgStartupTime = resp.TestRun.DeviceStatus[i].HistAvgStartupTime / uint(len(histStatus))
		}
	}

	for i := range resp.TestRun.Protocols {
		defaultNaN(&resp.TestRun.Protocols[i])
	}

	// get prev
	s.db.Table("test_runs").Where("test_id = ? and id < ?", testId, runId).Order("created_at desc").Limit(1).Select("id").Scan(&resp.PrevRunId)

	// get next
	s.db.Table("test_runs").Where("test_id = ? and id > ?", testId, runId).Order("created_at asc").Limit(1).Select("id").Scan(&resp.NextRunId)
	resp.State = s.runState(resp.TestRun.ID)

	c.JSON(http.StatusOK, resp)
}

func defaultNaN(protocol *models.TestProtocol) {
	if math.IsNaN(protocol.AvgCPU) {
		protocol.AvgCPU = 0
	}
	if math.IsNaN(protocol.AvgFPS) {
		protocol.AvgFPS = 0
	}
	if math.IsNaN(protocol.AvgMEM) {
		protocol.AvgMEM = 0
	}
	if math.IsNaN(protocol.AvgTriangles) {
		protocol.AvgTriangles = 0
	}
	if math.IsNaN(protocol.AvgVertexCount) {
		protocol.AvgVertexCount = 0
	}

	if math.IsNaN(protocol.HistAvgFPS) {
		protocol.HistAvgFPS = 0
	}
	if math.IsNaN(protocol.HistAvgMEM) {
		protocol.HistAvgMEM = 0
	}
	if math.IsNaN(protocol.HistAvgCPU) {
		protocol.HistAvgCPU = 0
	}
}

func (s *Service) getTestRunProtocol(c *gin.Context, project *models.Project, application *models.App) {
	runId := c.Param("run_id")
	protocolId := c.Param("protocol_id")

	var run models.TestRun
	if err := s.db.Preload("Protocols", "id = ?", protocolId).Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Protocols.Performance").Preload("Log").Preload("AppBinary").Preload("Test").First(&run, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if len(run.Protocols) > 0 {
		defaultNaN(&run.Protocols[0])

		// earlier runs of the same test on the same device that measured
		// performance: not this run (its runner protocol shares the test
		// name but has no measurements), not later runs, and not runners
		// of earlier runs, which would count as zeros
		current := &run.Protocols[0]
		var histProtocols []models.TestProtocol
		if err := s.db.Where("device_id = ? and test_name = ? and test_run_id < ?", current.DeviceID, current.TestName, run.ID).
			Where("exists (select 1 from protocol_performance_entries pe where pe.test_protocol_id = test_protocols.id and pe.deleted_at is null)").
			Order("id desc").Limit(AVG_HISTORY_LIMIT).Preload("Performance").Find(&histProtocols).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		for i := range histProtocols {
			defaultNaN(&histProtocols[i])

			current.HistAvgFPS += histProtocols[i].AvgFPS
			current.HistAvgMEM += histProtocols[i].AvgMEM
			current.HistAvgCPU += histProtocols[i].AvgCPU
			current.HistAvgVertexCount += histProtocols[i].AvgVertexCount
			current.HistAvgTriangles += histProtocols[i].AvgTriangles
		}
		if n := float64(len(histProtocols)); n > 0 {
			current.HistAvgFPS /= n
			current.HistAvgMEM /= n
			current.HistAvgCPU /= n
			current.HistAvgVertexCount /= n
			current.HistAvgTriangles /= n
		}
		run.Protocols[0].TestProtocolHistory = histProtocols
	}

	c.JSON(http.StatusOK, run)
}

func (s *Service) getData(c *gin.Context) {
	name := c.Param("name")
	c.File(filepath.Join(apps.TestDataPath, name))
}
