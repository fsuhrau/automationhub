package api

import (
	"net/http"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

const overviewLatestRuns = 10

// overviewRun is a run on the project overview, with its test and app
type overviewRun struct {
	runSummary
	TestID   uint   `json:"testId"`
	TestName string `json:"testName"`
	AppID    uint   `json:"appId"`
}

type projectOverview struct {
	Apps          int64 `json:"apps"`
	Tests         int   `json:"tests"`
	Bundles       int64 `json:"bundles"`
	BundleStorage int64 `json:"bundleStorage"`
	// tests whose latest run failed, was unstable or didn't run
	Failing []overviewRun `json:"failing"`
	// latest runs that are still running
	Running []overviewRun `json:"running"`
	// the latest runs of the project's tests, newest first
	Latest []overviewRun `json:"latest"`
}

// getOverview sums up a project: what it has, which tests need attention,
// what runs now and what ran last. The hub wide /stats mixed all projects.
func (s *Service) getOverview(c *gin.Context, project *models.Project) {
	var overview projectOverview
	appsOfProject := s.db.Model(&models.App{}).Select("id").Where("project_id = ?", project.ID)
	if err := s.db.Model(&models.App{}).Where("project_id = ?", project.ID).Count(&overview.Apps).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	var storage struct {
		N    int64
		Size int64
	}
	if err := s.db.Model(&models.AppBinary{}).Select("count(*) as n, coalesce(sum(size), 0) as size").
		Where("app_id IN (?)", appsOfProject).Scan(&storage).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	overview.Bundles, overview.BundleStorage = storage.N, storage.Size

	var tests []models.Test
	if err := s.db.Where("app_id IN (?)", appsOfProject).Order("id").Find(&tests).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	overview.Tests = len(tests)
	byID := map[uint]models.Test{}
	ids := make([]uint, len(tests))
	for i, t := range tests {
		byID[t.ID] = t
		ids[i] = t.ID
	}
	entry := func(r models.TestRun, summary runSummary) overviewRun {
		t := byID[r.TestID]
		return overviewRun{runSummary: summary, TestID: t.ID, TestName: t.Name, AppID: t.AppID}
	}

	overview.Failing, overview.Running, overview.Latest = []overviewRun{}, []overviewRun{}, []overviewRun{}
	if len(ids) == 0 {
		c.JSON(http.StatusOK, overview)
		return
	}

	// the latest run of each test says how the test is doing
	var lastRuns []models.TestRun
	if err := s.db.Where("id IN (?)", s.db.Model(&models.TestRun{}).Select("max(id)").Where("test_id IN ?", ids).Group("test_id")).
		Order("id desc").Find(&lastRuns).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	lastSummaries, err := s.summarizeRuns(lastRuns)
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	for i, r := range lastRuns {
		switch lastSummaries[i].Status {
		case "failed", "unstable", "not_run":
			overview.Failing = append(overview.Failing, entry(r, lastSummaries[i]))
		case "running":
			overview.Running = append(overview.Running, entry(r, lastSummaries[i]))
		}
	}

	var latest []models.TestRun
	if err := s.db.Where("test_id IN ?", ids).Order("id desc").Limit(overviewLatestRuns).Find(&latest).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	latestSummaries, err := s.summarizeRuns(latest)
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	for i, r := range latest {
		overview.Latest = append(overview.Latest, entry(r, latestSummaries[i]))
	}
	c.JSON(http.StatusOK, overview)
}
