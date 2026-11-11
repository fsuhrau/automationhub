package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/runs"
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

// archivedIconFile is where the icon of an archived bundle is kept, next to
// where its file was; "" when the bundle has no file path
func archivedIconFile(appPath string) string {
	file := bundleFile(appPath)
	if file == "" {
		return ""
	}
	return file + ".icon"
}

// getBinaryIcon serves the app icon out of the bundle file, or the copy kept
// when the bundle was archived
func (s *Service) getBinaryIcon(c *gin.Context, project *models.Project, application *models.App) {
	binary, ok := s.findBinary(c, application)
	if !ok {
		return
	}
	if binary.Archived() {
		iconFile := archivedIconFile(binary.AppPath)
		data, err := os.ReadFile(iconFile)
		if iconFile == "" || errors.Is(err, os.ErrNotExist) {
			s.error(c, http.StatusNotFound, app.ErrNoIcon)
			return
		}
		if err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		c.Header("Cache-Control", "private, max-age=86400")
		c.Data(http.StatusOK, http.DetectContentType(data), data)
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
	if binary.Archived() {
		s.error(c, http.StatusConflict, fmt.Errorf("bundle %d is archived; its file is gone", binary.ID))
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

// bundleInUse says whether a test run of this master runs with the bundle
func (s *Service) bundleInUse(binaryID uint) (bool, error) {
	var ids []uint
	// a running run is one of the latest that used the bundle
	if err := s.db.Model(&models.TestRun{}).Where("app_binary_id = ?", binaryID).
		Order("id desc").Limit(100).Pluck("id", &ids).Error; err != nil {
		return false, err
	}
	for _, id := range ids {
		if s.runState(id) == runs.RunRunning {
			return true, nil
		}
	}
	return false, nil
}

// archiveBinary removes the file of a bundle and keeps everything read from
// it, so the runs with it stay comparable. An archived bundle can't be
// installed or run anymore; uploading the same file again brings it back.
// Archiving twice removes what is left of the file.
func (s *Service) archiveBinary(c *gin.Context, project *models.Project, application *models.App) {
	binary, ok := s.findBinary(c, application)
	if !ok {
		return
	}
	if !binary.Archived() {
		inUse, err := s.bundleInUse(binary.ID)
		if err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		if inUse {
			s.error(c, http.StatusConflict, fmt.Errorf("a test run uses bundle %d; archive it when the run is over", binary.ID))
			return
		}
	}
	file := bundleFile(binary.AppPath)
	if file != "" && !binary.Archived() && binary.Info != nil && binary.Info.Icon != "" {
		// the icon is read out of the file, so keep a copy for the bundle pages
		if data, _, err := app.Icon(file, binary.Info.Icon); err == nil {
			_ = os.WriteFile(archivedIconFile(binary.AppPath), data, 0o644)
		}
	}
	if !binary.Archived() {
		// archived first, so no run or install starts with a file being removed
		now := time.Now()
		binary.ArchivedAt = &now
		if err := s.db.Model(binary).Update("archived_at", binary.ArchivedAt).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
	}
	if file != "" {
		if err := os.RemoveAll(file); err != nil {
			s.error(c, http.StatusInternalServerError, fmt.Errorf("bundle %d is archived, but its file stays: %w", binary.ID, err))
			return
		}
	}
	s.getBinary(c, project, application)
}
