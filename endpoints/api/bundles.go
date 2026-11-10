package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const bundleRunsLimit = 25

// bundleRun is a test run that used a bundle
type bundleRun struct {
	runSummary
	TestID   uint   `json:"testId"`
	TestName string `json:"testName"`
}

// bundleDetail is a bundle with all read from its file and the runs that
// used it, for the bundle page
type bundleDetail struct {
	models.AppBinary
	RunCount int64       `json:"runCount"`
	Runs     []bundleRun `json:"runs"`
}

// summaryInfo is the part of a bundle's info the bundle list shows
func summaryInfo(info *app.BundleInfo) *app.BundleInfo {
	if info == nil {
		return nil
	}
	return &app.BundleInfo{
		Format:      info.Format,
		DisplayName: info.DisplayName,
		Identifier:  info.Identifier,
		Version:     info.Version,
		Build:       info.Build,
		Icon:        info.Icon,
		Engine:      info.Engine,
		ReadAt:      info.ReadAt,
	}
}

func (s *Service) findBinary(c *gin.Context, application *models.App) (*models.AppBinary, bool) {
	binaryId := c.Param("binary_id")
	var binary models.AppBinary
	if err := s.db.First(&binary, "app_id = ? and id = ?", application.ID, binaryId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.error(c, http.StatusNotFound, fmt.Errorf("bundle %s not found", binaryId))
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return nil, false
	}
	return &binary, true
}

// getBinary is one bundle with everything known about it
func (s *Service) getBinary(c *gin.Context, project *models.Project, application *models.App) {
	binary, ok := s.findBinary(c, application)
	if !ok {
		return
	}
	detail := bundleDetail{AppBinary: *binary, Runs: []bundleRun{}}
	if err := s.db.Model(&models.TestRun{}).Where("app_binary_id = ?", binary.ID).Count(&detail.RunCount).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	var testRuns []models.TestRun
	if err := s.db.Preload("Test").Where("app_binary_id = ?", binary.ID).Order("id desc").Limit(bundleRunsLimit).Find(&testRuns).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	summaries, err := s.summarizeRuns(testRuns)
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	for i, r := range testRuns {
		run := bundleRun{runSummary: summaries[i], TestID: r.TestID}
		if r.Test != nil {
			run.TestName = r.Test.Name
		}
		detail.Runs = append(detail.Runs, run)
	}
	c.JSON(http.StatusOK, detail)
}

// getBinaryIcon serves the app icon out of the bundle file
func (s *Service) getBinaryIcon(c *gin.Context, project *models.Project, application *models.App) {
	binary, ok := s.findBinary(c, application)
	if !ok {
		return
	}
	file := bundleFile(binary.AppPath)
	if binary.Info == nil || binary.Info.Icon == "" || file == "" {
		s.error(c, http.StatusNotFound, app.ErrNoIcon)
		return
	}
	data, contentType, err := app.Icon(file, binary.Info.Icon)
	if err != nil {
		if errors.Is(err, app.ErrNoIcon) {
			s.error(c, http.StatusNotFound, err)
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return
	}
	// a bundle's file never changes, so neither does its icon
	c.Header("Cache-Control", "private, max-age=86400")
	c.Data(http.StatusOK, contentType, data)
}

// readBinaryAgain reads the stored file of a bundle again, for bundles
// uploaded before the master read all of it. The file, its hash and the tags
// stay; what the file says about the app is updated.
func (s *Service) readBinaryAgain(c *gin.Context, project *models.Project, application *models.App) {
	binary, ok := s.findBinary(c, application)
	if !ok {
		return
	}
	file := bundleFile(binary.AppPath)
	if file == "" {
		s.error(c, http.StatusConflict, fmt.Errorf("bundle %d has no file", binary.ID))
		return
	}
	analyser := app.NewAnalyser(file)
	if err := analyser.ReadInfo(); err != nil {
		s.error(c, http.StatusUnprocessableEntity, fmt.Errorf("unable to read the bundle: %w", err))
		return
	}
	params := analyser.GetParameter()
	binary.Info = params.Info
	columns := []string{"info"}
	set := func(column string, field *string, value string) {
		if value != "" {
			*field = value
			columns = append(columns, column)
		}
	}
	set("name", &binary.Name, params.Name)
	set("identifier", &binary.Identifier, params.Identifier)
	set("version", &binary.Version, params.Version)
	set("platform", &binary.Platform, params.Platform)
	if params.App.Android != nil {
		set("android_launch_activity", &binary.Android.LaunchActivity, params.App.Android.LaunchActivity)
	}
	if params.App.Executable != nil {
		set("executable_executable", &binary.Executable.Executable, params.App.Executable.Executable)
	}
	if err := s.db.Model(binary).Select(columns).Updates(binary).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	s.getBinary(c, project, application)
}
