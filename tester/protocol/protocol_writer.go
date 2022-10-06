package protocol

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"time"
)

type logProtocol struct {
	db     *gorm.DB
	p      *models.TestProtocol
	Writer *LogWriter
}

func (p *logProtocol) Close() {
	var state models.TestResultState

	if len(p.Writer.errs) > 0 {
		state = models.TestResultFailed
	} else {
		state = models.TestResultSuccess
	}

	endTime := time.Now()
	p.p.TestResult = state
	p.p.EndedAt = &endTime
	p.p.AvgCPU, p.p.AvgFPS, p.p.AvgMEM = p.Writer.GetAvgPerformanceMetrics()
	p.db.Updates(&p.p)
	events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: p.p.TestRunID, Protocol: p.p})
}

func (p *logProtocol) Errors() []error {
	return p.Writer.errs
}

type ProtocolWriter struct {
	db        *gorm.DB
	run       *models.TestRun
	testName  string
	protocols []*logProtocol
	projectId string
	appId     uint
}

func NewProtocolWriter(db *gorm.DB, testName string, run *models.TestRun) *ProtocolWriter {
	return &ProtocolWriter{db: db, testName: testName, run: run}
}

func (w *ProtocolWriter) NewProtocol(projectId string, appId uint, dev models.Device, testname string) (*logProtocol, error) {
	protocol := &models.TestProtocol{
		TestRunID: w.run.ID,
		DeviceID:  &dev.ID,
		TestName:  testname,
		StartedAt: time.Now(),
	}

	if err := w.db.Create(protocol).Error; err != nil {
		return nil, err
	}
	protocol.Device = &dev

	writer := NewLogWriter(w.db, protocol.ID)

	p := &logProtocol{w.db, protocol, writer}
	w.protocols = append(w.protocols, p)
	w.projectId = projectId
	w.appId = appId

	events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: w.run.ID, Protocol: protocol})

	return p, nil
}

func (w *ProtocolWriter) RunID() uint {
	return w.run.ID
}

func (w *ProtocolWriter) SessionID() string {
	return w.run.SessionID
}

func (w *ProtocolWriter) Close() {
	var (
		failedCount   int
		successCount  int
		unstableCount int
	)

	for _, p := range w.protocols {
		switch p.p.TestResult {
		case models.TestResultSuccess:
			successCount++
		case models.TestResultUnstable:
			unstableCount++
		case models.TestResultFailed:
			failedCount++
		}
	}

	events.TestRunFinished.Trigger(events.TestRunFinishedPayload{
		TestRunID: w.run.ID,
		TestRun:   w.run,
		TestName:  w.testName,
		ProjectID: w.projectId,
		AppID:     w.appId,
		Success:   failedCount == 0 && unstableCount == 0,
		Succeeded: successCount,
		Unstable:  unstableCount,
		Failed:    failedCount,
	})
}

func (w *ProtocolWriter) TrackStartupTime(deviceID uint, milliseconds int64) {
	entry := models.TestRunDeviceStatus{
		TestRunID:   w.RunID(),
		DeviceID:    deviceID,
		StartupTime: uint(milliseconds),
	}
	w.db.Create(&entry)
	w.run.DeviceStatus = append(w.run.DeviceStatus, entry)
}
