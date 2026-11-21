package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// testCreate is what the new test page sends: the test type and the same
// settings the edit page saves
type testCreate struct {
	Type models.TestType `json:"type"`
	testUpdate
}

// newTest creates a test with the edit page's rules; only Unity tests can be
// created here
func (s *Service) newTest(c *gin.Context, project *models.Project, application *models.App) {
	var req testCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	if req.Type != models.TestTypeUnity {
		s.error(c, http.StatusBadRequest, fmt.Errorf("only Unity tests can be created here"))
		return
	}
	if err := req.validate(true); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	if err := s.checkProjectDevices(project, req.Devices); err != nil {
		s.error(c, statusForError(err), err)
		return
	}

	test := models.Test{AppID: application.ID, Name: req.Name}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&test).Error; err != nil {
			return err
		}
		config := models.TestConfig{TestID: test.ID, Type: req.Type, AllDevices: req.AllDevices, ExecutionType: req.ExecutionType, RecordSession: req.RecordSession}
		if err := tx.Create(&config).Error; err != nil {
			return err
		}
		for _, d := range req.Devices {
			if err := tx.Create(&models.TestConfigDevice{TestConfigID: config.ID, DeviceID: d}).Error; err != nil {
				return err
			}
		}
		unity := models.TestConfigUnity{TestConfigID: config.ID, UnityTestCategoryType: req.UnityTestCategoryType, Categories: req.Categories}
		if err := tx.Create(&unity).Error; err != nil {
			return err
		}
		for _, f := range req.TestFunctions {
			f.TestConfigUnityID = unity.ID
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
	c.JSON(http.StatusCreated, test)
}

// checkProjectDevices makes sure every device exists and is visible to the
// project
func (s *Service) checkProjectDevices(project *models.Project, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	var visible int64
	if err := s.projectDevices(project).Model(&models.Device{}).Where("id IN ?", ids).Count(&visible).Error; err != nil {
		return err
	}
	if int(visible) != len(ids) {
		return badRequest("some of the picked devices don't exist")
	}
	return nil
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
	RecordSession         bool                         `json:"recordSession"`
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
	if err := s.checkProjectDevices(project, req.Devices); err != nil {
		s.error(c, statusForError(err), err)
		return
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Test{}).Where("id = ?", test.ID).Update("name", req.Name).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.TestConfig{}).Where("id = ?", test.TestConfig.ID).
			Updates(map[string]interface{}{"execution_type": req.ExecutionType, "all_devices": req.AllDevices, "record_session": req.RecordSession}).Error; err != nil {
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

// extractParams reads run parameters: key=value pairs separated by ";" as
// the run dialog sends them, or by line breaks as a run stores them and
// "Run again" sends them back. Only the first "=" separates key and value,
// so values such as URLs stay whole. Reading a stored run with ";" only made
// "MarketId=xu\nProxy=http://..." the market "xu\nProxy" and lost the proxy.
func extractParams(param string) map[string]string {
	env := make(map[string]string)
	params := strings.FieldsFunc(param, func(r rune) bool { return r == ';' || r == '\n' || r == '\r' })
	for _, p := range params {
		key, value, ok := strings.Cut(p, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			continue
		}
		env[key] = strings.TrimSpace(value)
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
	case errors.Is(err, ErrBundleExists), errors.Is(err, ErrNoDevices):
		return http.StatusConflict
	case errors.As(err, new(invalidRequest)):
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
