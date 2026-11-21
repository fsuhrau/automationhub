package api

import (
	"math"
	"net/http"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func defaultNaN(protocol *models.TestProtocol) {
	if math.IsNaN(protocol.AvgCPU) {
		protocol.AvgCPU = 0
	}
	if math.IsNaN(protocol.AvgFPS) {
		protocol.AvgFPS = 0
	}
	if math.IsNaN(protocol.AvgMEM) {
		protocol.AvgMEM = 0
	}
	if math.IsNaN(protocol.AvgTriangles) {
		protocol.AvgTriangles = 0
	}
	if math.IsNaN(protocol.AvgVertexCount) {
		protocol.AvgVertexCount = 0
	}

	if math.IsNaN(protocol.HistAvgFPS) {
		protocol.HistAvgFPS = 0
	}
	if math.IsNaN(protocol.HistAvgMEM) {
		protocol.HistAvgMEM = 0
	}
	if math.IsNaN(protocol.HistAvgCPU) {
		protocol.HistAvgCPU = 0
	}
}

func (s *Service) getTestRunProtocol(c *gin.Context, project *models.Project, application *models.App) {
	_, runId, ok := s.scopedRun(c, application)
	if !ok {
		return
	}
	protocolId, ok := s.idParam(c, "protocol_id")
	if !ok {
		return
	}

	var run models.TestRun
	if err := s.db.Preload("Protocols", "id = ?", protocolId).Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Protocols.Performance").Preload("Protocols.Keys", func(db *gorm.DB) *gorm.DB { return db.Order("key") }).Preload("Log").Preload("AppBinary").Preload("Test").First(&run, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if len(run.Protocols) > 0 {
		defaultNaN(&run.Protocols[0])

		// earlier runs of the same test on the same device that measured
		// performance: not this run (its runner protocol shares the test
		// name but has no measurements), not later runs, and not runners
		// of earlier runs, which would count as zeros
		current := &run.Protocols[0]
		var histProtocols []models.TestProtocol
		if err := s.db.Where("device_id = ? and test_name = ? and test_run_id < ?", current.DeviceID, current.TestName, run.ID).
			Where("test_run_id in (select tr.id from test_runs tr where tr.test_id in "+appTestIDs+")", application.ID).
			Where("exists (select 1 from protocol_performance_entries pe where pe.test_protocol_id = test_protocols.id and pe.deleted_at is null)").
			Order("id desc").Limit(AVG_HISTORY_LIMIT).Preload("Performance").Find(&histProtocols).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		for i := range histProtocols {
			defaultNaN(&histProtocols[i])

			current.HistAvgFPS += histProtocols[i].AvgFPS
			current.HistAvgMEM += histProtocols[i].AvgMEM
			current.HistAvgCPU += histProtocols[i].AvgCPU
			current.HistAvgVertexCount += histProtocols[i].AvgVertexCount
			current.HistAvgTriangles += histProtocols[i].AvgTriangles
		}
		if n := float64(len(histProtocols)); n > 0 {
			current.HistAvgFPS /= n
			current.HistAvgMEM /= n
			current.HistAvgCPU /= n
			current.HistAvgVertexCount /= n
			current.HistAvgTriangles /= n
		}
		run.Protocols[0].TestProtocolHistory = histProtocols
	}

	c.JSON(http.StatusOK, run)
}
