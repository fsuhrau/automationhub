package api

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// A protocol's performance is compared with a baseline: another run of the
// same test that measured performance. By default that is the latest earlier
// run on the same device; these endpoints let the user pick one by bundle
// name, version, tag or run, on the same device or on any. A baseline is
// chosen by run, so it stays while moving through the tests of a run.

const (
	baselinesDefault = 50
	baselinesMax     = 200
)

type baselineBundle struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Tags     string `json:"tags"`
}

type baselineCandidate struct {
	ProtocolID     uint                   `json:"protocolId"`
	TestRunID      uint                   `json:"testRunId"`
	TestID         uint                   `json:"testId"`
	StartedAt      time.Time              `json:"startedAt"`
	TestResult     models.TestResultState `json:"testResult"`
	DeviceID       *uint                  `json:"deviceId,omitempty"`
	Device         string                 `json:"device"`
	Bundle         *baselineBundle        `json:"bundle"`
	Parameter      string                 `json:"parameter"`
	AvgFPS         float64                `json:"avgFps"`
	AvgMEM         float64                `json:"avgMem"`
	AvgCPU         float64                `json:"avgCpu"`
	AvgVertexCount float64                `json:"avgVertexCount"`
	AvgTriangles   float64                `json:"avgTriangles"`
}

type baselinesResponse struct {
	Candidates []baselineCandidate `json:"candidates"`
	// the tags of the app's bundles, to filter by
	Tags []string `json:"tags"`
}

type baselineResponse struct {
	baselineCandidate
	Performance []models.ProtocolPerformanceEntry `json:"performance"`
}

// protocolOfApp loads the protocol of the URL and checks that its run belongs
// to the test of the URL and the test to the app; it answers the request
// itself when it doesn't
func (s *Service) protocolOfApp(c *gin.Context, application *models.App) (*models.TestProtocol, bool) {
	var protocol models.TestProtocol
	err := s.db.Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Where("test_protocols.id = ? AND test_protocols.test_run_id = ? AND test_runs.test_id = ? AND tests.app_id = ?",
			c.Param("protocol_id"), c.Param("run_id"), c.Param("test_id"), application.ID).
		First(&protocol).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.error(c, http.StatusNotFound, fmt.Errorf("protocol %s not found in run %s", c.Param("protocol_id"), c.Param("run_id")))
		return nil, false
	}
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return nil, false
	}
	return &protocol, true
}

// baselineQuery selects the protocols a protocol can be compared with: the
// same test name in another run of the app, with performance measured
func (s *Service) baselineQuery(application *models.App, current *models.TestProtocol) *gorm.DB {
	return s.db.Model(&models.TestProtocol{}).
		Select("test_protocols.*").
		Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id AND test_runs.deleted_at IS NULL").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("LEFT JOIN app_binaries ON app_binaries.id = test_runs.app_binary_id").
		Where("tests.app_id = ? AND test_protocols.test_name = ? AND test_protocols.test_run_id <> ?",
			application.ID, current.TestName, current.TestRunID).
		Where("exists (select 1 from protocol_performance_entries pe where pe.test_protocol_id = test_protocols.id and pe.deleted_at is null)")
}

func toCandidate(p *models.TestProtocol) baselineCandidate {
	defaultNaN(p)
	b := baselineCandidate{
		ProtocolID: p.ID, TestRunID: p.TestRunID, StartedAt: p.StartedAt, TestResult: p.TestResult, DeviceID: p.DeviceID,
		AvgFPS: p.AvgFPS, AvgMEM: p.AvgMEM, AvgCPU: p.AvgCPU, AvgVertexCount: p.AvgVertexCount, AvgTriangles: p.AvgTriangles,
	}
	if p.Device != nil {
		b.Device = p.Device.Alias
		if b.Device == "" {
			b.Device = p.Device.Name
		}
	}
	if p.TestRun != nil {
		b.TestID = p.TestRun.TestID
		b.Parameter = p.TestRun.Parameter
		if bin := p.TestRun.AppBinary; bin != nil {
			b.Bundle = &baselineBundle{ID: bin.ID, Name: bin.Name, Version: bin.Version, Platform: bin.Platform, Tags: bin.Tags}
		}
	}
	return b
}

// getProtocolBaselines lists what the protocol can be compared with, newest
// first. Filters: q matches the bundle's name, version or tags, or a run id;
// tag matches one tag exactly; anyDevice=true drops the same device rule.
func (s *Service) getProtocolBaselines(c *gin.Context, project *models.Project, application *models.App) {
	current, ok := s.protocolOfApp(c, application)
	if !ok {
		return
	}
	limit := baselinesDefault
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = min(v, baselinesMax)
	}

	query := s.baselineQuery(application, current)
	if c.Query("anyDevice") != "true" {
		if current.DeviceID == nil {
			query = query.Where("test_protocols.device_id IS NULL")
		} else {
			query = query.Where("test_protocols.device_id = ?", *current.DeviceID)
		}
	}
	if tag := strings.TrimSpace(c.Query("tag")); tag != "" {
		query = query.Where("(',' || app_binaries.tags || ',') LIKE ? ESCAPE '\\'", "%,"+escapeLike(tag)+",%")
	}
	if q := strings.ToLower(strings.TrimSpace(c.Query("q"))); q != "" {
		like := "%" + escapeLike(q) + "%"
		cond := `LOWER(app_binaries.name) LIKE ? ESCAPE '\' OR LOWER(app_binaries.version) LIKE ? ESCAPE '\'` +
			` OR LOWER(app_binaries.tags) LIKE ? ESCAPE '\' OR LOWER(test_runs.parameter) LIKE ? ESCAPE '\'`
		args := []interface{}{like, like, like, like}
		if id, err := strconv.Atoi(strings.TrimPrefix(q, "#")); err == nil {
			cond += " OR test_runs.id = ?"
			args = append(args, id)
		}
		query = query.Where("("+cond+")", args...)
	}

	var protocols []models.TestProtocol
	if err := query.Preload("Device").Preload("TestRun").Preload("TestRun.AppBinary").
		Order("test_protocols.id desc").Limit(limit).Find(&protocols).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	resp := baselinesResponse{Candidates: make([]baselineCandidate, len(protocols)), Tags: []string{}}
	for i := range protocols {
		resp.Candidates[i] = toCandidate(&protocols[i])
	}

	var tagLists []string
	if err := s.db.Model(&models.AppBinary{}).Where("app_id = ? AND tags <> ''", application.ID).
		Distinct().Pluck("tags", &tagLists).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	seen := map[string]bool{}
	for _, list := range tagLists {
		for _, t := range strings.Split(list, ",") {
			if t = strings.TrimSpace(t); t != "" && !seen[t] {
				seen[t] = true
				resp.Tags = append(resp.Tags, t)
			}
		}
	}
	sort.Strings(resp.Tags)

	c.JSON(http.StatusOK, resp)
}

// getProtocolBaseline loads the baseline in a run with its checkpoints: the
// protocol of the same test in that run, on the device given by deviceId, or
// else preferably on the same device. It has to be one the protocol can be
// compared with.
func (s *Service) getProtocolBaseline(c *gin.Context, project *models.Project, application *models.App) {
	current, ok := s.protocolOfApp(c, application)
	if !ok {
		return
	}
	query := s.baselineQuery(application, current).Where("test_protocols.test_run_id = ?", c.Param("baseline_run_id"))
	if deviceID, err := strconv.Atoi(c.Query("deviceId")); err == nil {
		query = query.Where("test_protocols.device_id = ?", deviceID)
	} else if current.DeviceID != nil {
		// a number, so it can be part of the clause
		query = query.Order(fmt.Sprintf("CASE WHEN test_protocols.device_id = %d THEN 0 ELSE 1 END", *current.DeviceID))
	}
	var baseline models.TestProtocol
	err := query.Order("test_protocols.id desc").
		Preload("Device").Preload("TestRun").Preload("TestRun.AppBinary").
		Preload("Performance", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		First(&baseline).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.error(c, http.StatusNotFound, fmt.Errorf("run %s has no performance of %s to compare with", c.Param("baseline_run_id"), current.TestName))
		return
	}
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, baselineResponse{baselineCandidate: toCandidate(&baseline), Performance: baseline.Performance})
}

// escapeLike makes %, _ and \ in user input match literally in a LIKE with
// ESCAPE '\' (SQLite and Postgres alike)
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`).Replace(s)
}
