package api

import (
	"net/http"
	"strconv"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

const (
	deviceProtocolsDefault = 10
	deviceProtocolsMax     = 50
)

// getDeviceProtocols lists the latest test results on a device, newest
// first, for the device page. Only tests of this project count, and only
// the tests themselves: a runner protocol that wraps a test has the same
// name and would show every test twice.
func (s *Service) getDeviceProtocols(c *gin.Context, project *models.Project) {
	device, ok := s.findProjectDevice(c, project)
	if !ok {
		return
	}
	limit := deviceProtocolsDefault
	if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 {
		limit = min(v, deviceProtocolsMax)
	}

	var protocols []models.TestProtocol
	if err := s.db.
		Joins("JOIN test_runs ON test_runs.id = test_protocols.test_run_id").
		Joins("JOIN tests ON tests.id = test_runs.test_id").
		Joins("JOIN apps ON apps.id = tests.app_id").
		Where("test_protocols.device_id = ? AND apps.project_id = ?", device.ID, project.ID).
		Where("NOT EXISTS (SELECT 1 FROM test_protocols c WHERE c.parent_test_protocol_id = test_protocols.id AND c.deleted_at IS NULL)").
		Preload("TestRun").Preload("TestRun.Test").
		Order("test_protocols.id desc").Limit(limit).
		Find(&protocols).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	for i := range protocols {
		defaultNaN(&protocols[i])
	}
	c.JSON(http.StatusOK, protocols)
}
