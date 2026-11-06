package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) ResolveApp(context *gin.Context) {
	appId := context.Param("app_id")

	p, _ := context.Get("project")
	project := p.(*models.Project)

	var application models.App
	if err := s.db.Preload("Parameter").First(&application, "project_id = ? and id = ?", project.ID, appId).Error; err != nil {
		s.error(context, http.StatusNotFound, err)
		return
	}
	context.Set("app", &application)
}

func (s *Service) WithApp(wrapperFunction func(*gin.Context, *models.Project, *models.App)) func(*gin.Context) {
	return func(context *gin.Context) {
		p, _ := context.Get("project")
		project := p.(*models.Project)
		a, _ := context.Get("app")
		application := a.(*models.App)
		wrapperFunction(context, project, application)
	}
}

func (s *Service) getApps(c *gin.Context, project *models.Project) {

	var apps []models.App
	if err := s.db.Where("project_id = ?", project.ID).Preload("Parameter").Find(&apps).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, apps)
}

// bundleListEntry is a bundle with how many test runs used it, so the
// bundles page can warn before deleting one that runs point to
type bundleListEntry struct {
	models.AppBinary
	RunCount int64 `json:"runCount"`
}

// getBinaries lists the bundles of an app, newest first
func (s *Service) getBinaries(c *gin.Context, project *models.Project, application *models.App) {
	var bundles []models.AppBinary
	if err := s.db.Where("app_id = ?", application.ID).Order("id desc").Find(&bundles).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	ids := make([]uint, len(bundles))
	for i := range bundles {
		ids[i] = bundles[i].ID
	}
	var counts []struct {
		AppBinaryID uint
		Runs        int64
	}
	if len(ids) > 0 {
		if err := s.db.Model(&models.TestRun{}).Select("app_binary_id, count(*) as runs").
			Where("app_binary_id IN ?", ids).Group("app_binary_id").Scan(&counts).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
	}
	runs := map[uint]int64{}
	for _, n := range counts {
		runs[n.AppBinaryID] = n.Runs
	}
	list := make([]bundleListEntry, len(bundles))
	for i := range bundles {
		list[i] = bundleListEntry{AppBinary: bundles[i], RunCount: runs[bundles[i].ID]}
	}
	c.JSON(http.StatusOK, list)
}

/*
func (s *Service) getApp(c *gin.Context) {
	appId := c.Param("app_id")
	var app models.AppBinary
	if err := s.db.First(&app, appId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, app)
}
*/

// bundleFile is the stored file of a bundle, or "" when the bundle has none
// or its path would leave the storage folder. An empty AppPath used to make
// a delete remove the whole folder, with every bundle in it.
func bundleFile(appPath string) string {
	if appPath == "" {
		return ""
	}
	root := filepath.Clean(apps.AppStoragePath)
	file := filepath.Join(root, appPath)
	if file == root || !strings.HasPrefix(file, root+string(filepath.Separator)) {
		return ""
	}
	return file
}

func (s *Service) deleteBinary(c *gin.Context, project *models.Project, application *models.App) {
	binaryId := c.Param("binary_id")
	var binary models.AppBinary
	if err := s.db.First(&binary, "app_id = ? and id = ?", application.ID, binaryId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.error(c, http.StatusNotFound, fmt.Errorf("bundle %s not found", binaryId))
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return
	}
	if file := bundleFile(binary.AppPath); file != "" {
		if err := os.RemoveAll(file); err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
	}

	if err := s.db.Delete(&binary).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}

// createApp adds an app to the project; it needs a name, a platform an app
// can target, and an identifier for the platforms that install bundles
func (s *Service) createApp(c *gin.Context, project *models.Project) {
	var application models.App
	if err := c.ShouldBindJSON(&application); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	application.Name = strings.TrimSpace(application.Name)
	application.Identifier = strings.TrimSpace(application.Identifier)
	if application.Name == "" {
		s.error(c, http.StatusBadRequest, fmt.Errorf("the app needs a name"))
		return
	}
	switch application.Platform {
	case models.PlatformTypeAndroid, models.PlatformTypeiOS, models.PlatformTypeMac:
		if application.Identifier == "" {
			s.error(c, http.StatusBadRequest, fmt.Errorf("the app needs its bundle identifier or package name"))
			return
		}
	case models.PlatformTypeWindows, models.PlatformTypeLinux, models.PlatformTypeWeb, models.PlatformTypeEditor:
	default:
		s.error(c, http.StatusBadRequest, fmt.Errorf("unknown platform %d", application.Platform))
		return
	}
	application.ID = 0
	application.ProjectID = project.ID
	application.Parameter = nil
	if err := s.db.Create(&application).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, application)
}

func (s *Service) updateApp(c *gin.Context, project *models.Project, application *models.App) {
	var request models.App
	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	application.Name = request.Name

	// allow change it once since it's a default generated one
	if application.Identifier == "default_app" {
		application.Identifier = request.Identifier
	}

	if err := s.db.Save(&application).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, application)
}

// parameterType is the JSON the parameter's type is stored as
type parameterType struct {
	Type         string   `json:"type"`
	DefaultValue string   `json:"defaultValue"`
	Options      []string `json:"options,omitempty"`
}

// validateParameter checks a parameter of an app: a name no other parameter
// of the app has (but itself, by id), a text or choice type, and for a
// choice its options with the default among them. It stores the tidied type.
func (s *Service) validateParameter(application *models.App, p *models.AppParameter, id uint) error {
	p.Name = strings.TrimSpace(p.Name)
	if p.Name == "" {
		return badRequest("the parameter needs a name")
	}
	var others []models.AppParameter
	if err := s.db.Where("app_id = ? AND id <> ?", application.ID, id).Find(&others).Error; err != nil {
		return err
	}
	for _, o := range others {
		if strings.EqualFold(o.Name, p.Name) {
			return badRequest("the app has a parameter %s already", o.Name)
		}
	}
	var t parameterType
	if err := json.Unmarshal([]byte(p.Type), &t); err != nil {
		return badRequest("the parameter's type isn't valid")
	}
	switch t.Type {
	case "string":
		t.Options = nil
	case "option":
		seen := map[string]bool{}
		var options []string
		for _, o := range t.Options {
			o = strings.TrimSpace(o)
			if o != "" && !seen[o] {
				seen[o] = true
				options = append(options, o)
			}
		}
		if len(options) == 0 {
			return badRequest("a choice needs at least one option")
		}
		t.Options = options
		if t.DefaultValue != "" && !seen[t.DefaultValue] {
			return badRequest("the default %s isn't one of the options", t.DefaultValue)
		}
	default:
		return badRequest("unknown parameter type %q", t.Type)
	}
	tidy, err := json.Marshal(t)
	if err != nil {
		return err
	}
	p.Type = string(tidy)
	return nil
}

func (s *Service) addAppParameter(c *gin.Context, project *models.Project, application *models.App) {
	var request models.AppParameter
	if err := c.ShouldBindJSON(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	if err := s.validateParameter(application, &request, 0); err != nil {
		s.error(c, statusForError(err), err)
		return
	}

	request.ID = 0
	request.AppID = application.ID
	if err := s.db.Create(&request).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &request)
}

func (s *Service) removeAppParameter(c *gin.Context, project *models.Project, application *models.App) {

	paramId := c.Param("parameter_id")

	if err := s.db.Delete(&models.AppParameter{}, "app_id = ? and id = ?", application.ID, paramId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}

// updateAppParameter changes a parameter and returns it (it used to return
// the whole app)
func (s *Service) updateAppParameter(c *gin.Context, project *models.Project, application *models.App) {
	var request models.AppParameter
	if err := c.ShouldBindJSON(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	var param models.AppParameter
	if err := s.db.First(&param, "app_id = ? and id = ?", application.ID, c.Param("parameter_id")).Error; err != nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("parameter %s not found", c.Param("parameter_id")))
		return
	}
	if err := s.validateParameter(application, &request, param.ID); err != nil {
		s.error(c, statusForError(err), err)
		return
	}

	param.Type = request.Type
	param.Name = request.Name
	if err := s.db.Save(&param).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, &param)
}

// normalizeTags trims the comma separated tags and drops empty and repeated ones
func normalizeTags(tags string) string {
	var out []string
	seen := map[string]bool{}
	for _, t := range strings.Split(tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return strings.Join(out, ",")
}

// updateBinary changes the tags of a bundle; nothing else of a bundle is
// editable
func (s *Service) updateBinary(c *gin.Context, project *models.Project, application *models.App) {
	var req struct {
		Tags string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	var binary models.AppBinary
	if err := s.db.First(&binary, "app_id = ? and id = ?", application.ID, c.Param("binary_id")).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.error(c, http.StatusNotFound, fmt.Errorf("bundle %s not found", c.Param("binary_id")))
		} else {
			s.error(c, http.StatusInternalServerError, err)
		}
		return
	}

	binary.Tags = normalizeTags(req.Tags)
	if err := s.db.Model(&binary).Update("tags", binary.Tags).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, binary)
}

func (s *Service) uploadBinary(c *gin.Context, project *models.Project, application *models.App) {
	file, err := c.FormFile("test_target")
	if err != nil {
		s.error(c, http.StatusBadRequest, fmt.Errorf("get form err: %s", err.Error()))
		return
	}

	if err := os.MkdirAll(apps.AppStoragePath, os.ModePerm); err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	filePath := filepath.Join(apps.AppStoragePath, filepath.Base(file.Filename))
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		s.error(c, http.StatusInternalServerError, fmt.Errorf("upload file err: %s", err.Error()))
		return
	}

	bundle, err := s.StoreBundle(application, filePath)
	if err != nil {
		s.error(c, statusForError(err), err)
		return
	}
	c.JSON(http.StatusCreated, bundle)
}
