package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/teststore"
	"github.com/gin-gonic/gin"
)

const (
	AVG_HISTORY_LIMIT = 20
)

// runSummary is one run in a list; the test service owns the type.
type runSummary = testsvc.RunSummary

// testService is the test definition service on the database of this api.
func (s *Service) testService() *testsvc.Service {
	return testsvc.New(testsvc.Deps{Repo: teststore.New(s.db), RunState: s.runState})
}

// summarizeRuns gives each run its status and result counts, in the order
// of runs. The runs, bundles and overview pages still use it; they get the
// service directly when they move out of this package.
func (s *Service) summarizeRuns(testRuns []models.TestRun) ([]runSummary, error) {
	return s.testService().Summaries(context.Background(), testRuns)
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
