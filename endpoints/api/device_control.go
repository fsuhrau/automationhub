package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/device"
	controlapi "github.com/fsuhrau/automationhub/endpoints/control"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/base"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (s *Service) registerDeviceControl(projectApi *gin.RouterGroup) {
	controlapi.NewHandler(s.control, s.resolveControlDevice, s.resolveControlApp, requestOwner, s.nodeUrl).Register(projectApi)
}

func projectFromContext(c *gin.Context) *models.Project {
	p, _ := c.Get("project")
	project, _ := p.(*models.Project)
	return project
}

// resolveControlDevice returns the live device for :device_id within the project.
func (s *Service) resolveControlDevice(c *gin.Context) (device.Device, bool) {
	project := projectFromContext(c)
	if project == nil {
		s.error(c, http.StatusNotFound, fmt.Errorf("project not found"))
		return nil, false
	}
	model, ok := s.findProjectDevice(c, project)
	if !ok {
		return nil, false
	}
	dev, _ := s.devicesManager.GetDevice(model.DeviceIdentifier)
	if dev == nil {
		s.error(c, http.StatusConflict, fmt.Errorf("device %s is not connected", model.DeviceIdentifier))
		return nil, false
	}
	return dev, true
}

// resolveControlApp loads an app bundle that belongs to the project.
func (s *Service) resolveControlApp(c *gin.Context, binaryID uint) (*app.Parameter, error) {
	project := projectFromContext(c)
	if project == nil {
		return nil, fmt.Errorf("project not found")
	}
	var binary models.AppBinary
	err := s.db.Preload("App").
		Joins("JOIN apps ON apps.id = app_binaries.app_id").
		Where("apps.project_id = ?", project.ID).
		First(&binary, "app_binaries.id = ?", binaryID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("app bundle %d not found", binaryID)
	}
	if err != nil {
		return nil, err
	}
	params := base.GetParams(&binary, "")
	return &params, nil
}

// requestOwner names the authenticated caller for device leases.
func requestOwner(c *gin.Context) string {
	value, _ := c.Get(authentication.UserKey)
	switch u := value.(type) {
	case string:
		if u != "" {
			return "token:" + u
		}
	case authentication.User:
		if u.Name != "" {
			return u.Name
		}
		if u.Login != "" {
			return u.Login
		}
	case *authentication.User:
		if u != nil && u.Name != "" {
			return u.Name
		}
		if u != nil && u.Login != "" {
			return u.Login
		}
	}
	return "anonymous"
}
