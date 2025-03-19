package protocol

import (
	"github.com/fsuhrau/automationhub/device"
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
	p.p.AvgCPU, p.p.AvgFPS, p.p.AvgMEM, p.p.AvgVertexCount, p.p.AvgTriangles = p.Writer.GetAvgPerformanceMetrics()
	p.db.Updates(&p.p)
	events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: p.p.TestRunID, Protocol: p.p})
}

func (p *logProtocol) Errors() []error {
	return p.Writer.errs
}

func (p *logProtocol) TestProtocolId() *uint {
	return &p.p.ID
}

type ProtocolWriter struct {
	db        *gorm.DB
	run       *models.TestRun
	testName  string
	protocols []*logProtocol
	projectId string
	appId     uint
}

func NewProtocolWriter(db *gorm.DB, projectId string, appId uint, testName string, run *models.TestRun) *ProtocolWriter {
	return &ProtocolWriter{db: db, projectId: projectId, appId: appId, testName: testName, run: run}
}

func (w *ProtocolWriter) NewProtocol(dev models.Device, testname string) (*logProtocol, error) {
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

	writer := NewLogWriter(w.db, protocol.ID, &dev, nil)

	p := &logProtocol{w.db, protocol, writer}
	w.protocols = append(w.protocols, p)

	events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: w.run.ID, Protocol: protocol})

	return p, nil
}

func (w *ProtocolWriter) NewSubProtocol(testname string, pw device.LogWriter) (*logProtocol, error) {

	dev := pw.Device().(*models.Device)

	protocol := &models.TestProtocol{
		TestRunID:            w.run.ID,
		ParentTestProtocolID: pw.TestProtocolId(),
		DeviceID:             &dev.ID,
		TestName:             testname,
		StartedAt:            time.Now(),
	}

	if err := w.db.Create(protocol).Error; err != nil {
		return nil, err
	}
	protocol.Device = dev

	writer := NewLogWriter(w.db, protocol.ID, dev, pw)

	p := &logProtocol{w.db, protocol, writer}
	w.protocols = append(w.protocols, p)

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

	for i := len(w.protocols) - 1; i > 0; i-- {
		w.protocols[i].Close()

		switch w.protocols[i].p.TestResult {
		case models.TestResultOpen:
			fallthrough
		case models.TestResultSuccess:
			successCount++
		case models.TestResultUnstable:
			unstableCount++
		case models.TestResultFailed:
			failedCount++
		default:
			panic("unhandled default case")
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
