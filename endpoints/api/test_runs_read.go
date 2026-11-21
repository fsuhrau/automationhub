package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
	if err := query.Preload("AppBinary", func(db *gorm.DB) *gorm.DB { return db.Omit("info") }).Order("id desc").Limit(limit + 1).Find(&testRuns).Error; err != nil {
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
	// Devices are the IDs of the devices a running run holds now; empty
	// once the run is over.
	Devices []uint `json:"devices"`
}

func (s *Service) runState(runID uint) runs.RunState {
	if s.runTracker == nil {
		return runs.RunUnknown
	}
	state, _ := s.runTracker.Status(runID)
	return state
}

// runDevices returns the devices an active run of this master holds.
func (s *Service) runDevices(runID uint) []uint {
	s.runnersMutex.Lock()
	runner, ok := s.runners[runKey(runID)]
	s.runnersMutex.Unlock()
	if !ok {
		return []uint{}
	}
	return runner.LockedDevices()
}

// appTestIDs selects the tests of one app; histories only look at runs of
// these, so they never mix in another project's data
const appTestIDs = "(select t.id from tests t where t.app_id = ? and t.deleted_at is null)"

// idParam reads a numeric id from the path; anything else is a 400
func (s *Service) idParam(c *gin.Context, name string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("invalid %s %q", name, c.Param(name)))
		return 0, false
	}
	return uint(id), true
}

// lookupFailed answers a failed lookup: 404 when the row does not exist
// (or belongs to somebody else), 500 for a database error
func (s *Service) lookupFailed(c *gin.Context, what string, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.error(c, http.StatusNotFound, fmt.Errorf("%s %s not found", what, c.Param(what+"_id")))
		return
	}
	s.error(c, http.StatusInternalServerError, err)
}

// scopedTestID is the id of the path's test when it belongs to the app
func (s *Service) scopedTestID(c *gin.Context, application *models.App) (uint, bool) {
	id, ok := s.idParam(c, "test_id")
	if !ok {
		return 0, false
	}
	var test models.Test
	if err := s.db.Select("id").First(&test, "id = ? AND app_id = ?", id, application.ID).Error; err != nil {
		s.lookupFailed(c, "test", err)
		return 0, false
	}
	return test.ID, true
}

// scopedRunID is the id of the path's run when it belongs to the test
func (s *Service) scopedRunID(c *gin.Context, testID uint) (uint, bool) {
	id, ok := s.idParam(c, "run_id")
	if !ok {
		return 0, false
	}
	var run models.TestRun
	if err := s.db.Select("id").First(&run, "id = ? AND test_id = ?", id, testID).Error; err != nil {
		s.lookupFailed(c, "run", err)
		return 0, false
	}
	return run.ID, true
}

// scopedRun resolves the path's test and run, both inside the app
func (s *Service) scopedRun(c *gin.Context, application *models.App) (testID, runID uint, ok bool) {
	if testID, ok = s.scopedTestID(c, application); !ok {
		return 0, 0, false
	}
	runID, ok = s.scopedRunID(c, testID)
	return testID, runID, ok
}

func (s *Service) getLastTestRun(c *gin.Context, project *models.Project, application *models.App) {
	testId, ok := s.scopedTestID(c, application)
	if !ok {
		return
	}

	var resp Response
	if err := s.db.Preload("DeviceStatus").Preload("DeviceStatus.Device").Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("AppBinary").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").Preload("Protocols.Keys", func(db *gorm.DB) *gorm.DB { return db.Order("key") }).Where("test_id = ?", testId).Order("id desc").First(&resp.TestRun).Error; err != nil {
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
	resp.Devices = s.runDevices(resp.TestRun.ID)

	c.JSON(http.StatusOK, resp)
}

func (s *Service) getTestRun(c *gin.Context, project *models.Project, application *models.App) {
	testId, runId, ok := s.scopedRun(c, application)
	if !ok {
		return
	}

	var resp Response
	if err := s.db.Preload("DeviceStatus").Preload("DeviceStatus.Device").Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("AppBinary").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").Preload("Recordings", func(db *gorm.DB) *gorm.DB { return db.Order("started_at") }).Preload("Protocols.Keys", func(db *gorm.DB) *gorm.DB { return db.Order("key") }).First(&resp.TestRun, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	for i := range resp.TestRun.DeviceStatus {
		var histStatus []models.TestRunDeviceStatus
		if err := s.db.Where("device_id = ? and test_run_id in (select tr.id from test_runs tr where tr.id < ? and tr.test_id in "+appTestIDs+" order by tr.id desc limit ?)", resp.TestRun.DeviceStatus[i].DeviceID, resp.TestRun.DeviceStatus[i].TestRunID, application.ID, AVG_HISTORY_LIMIT).Find(&histStatus).Error; err != nil {
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
	resp.Devices = s.runDevices(resp.TestRun.ID)

	c.JSON(http.StatusOK, resp)
}
