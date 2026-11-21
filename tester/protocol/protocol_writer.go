package protocol

import (
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"sync"
	"time"
)

type logProtocol struct {
	db     *gorm.DB
	p      *models.TestProtocol
	Writer *LogWriter
}

func (p *logProtocol) Close() {
	var state models.TestResultState

	if p.Writer.passed {
		state = models.TestResultSuccess
	} else {
		state = models.TestResultFailed
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
	// failed marks a run that couldn't run its tests, e.g. without devices
	failed   bool
	db       *gorm.DB
	run      *models.TestRun
	testName string
	// protocols grows from several devices of a run at once
	mu        sync.Mutex
	protocols []*logProtocol
	projectId string
	appId     uint
}

// RunTarget names the run a ProtocolWriter writes the protocols of: its test
// and the app and project the test belongs to.
type RunTarget struct {
	ProjectID string
	AppID     uint
	TestName  string
	Run       *models.TestRun
}

func NewProtocolWriter(db *gorm.DB, target RunTarget) *ProtocolWriter {
	return &ProtocolWriter{db: db, projectId: target.ProjectID, appId: target.AppID, testName: target.TestName, run: target.Run}
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

	writer := NewLogWriter(w.db, LogTarget{RunID: w.run.ID, ProtocolID: protocol.ID, Device: &dev})

	p := &logProtocol{w.db, protocol, writer}
	w.track(p)

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

	writer := NewLogWriter(w.db, LogTarget{RunID: w.run.ID, ProtocolID: protocol.ID, Device: dev, Parent: pw})

	p := &logProtocol{w.db, protocol, writer}
	w.track(p)

	events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: w.run.ID, Protocol: protocol})

	return p, nil
}

// Fail marks the run as failed even when no test failed, e.g. because no
// device could run the tests.
func (w *ProtocolWriter) Fail() {
	w.failed = true
}

func (w *ProtocolWriter) RunID() uint {
	return w.run.ID
}

func (w *ProtocolWriter) SessionID() string {
	return w.run.SessionID
}

// resultCounts tallies the results of a run's protocols.
type resultCounts struct {
	failed, succeeded, unstable int
}

func (c *resultCounts) add(state models.TestResultState) {
	switch state {
	case models.TestResultSuccess:
		c.succeeded++
	case models.TestResultUnstable:
		c.unstable++
	default:
		// a failed or still open result is no success
		c.failed++
	}
}

// track remembers the protocol, to close it with the run.
func (w *ProtocolWriter) track(p *logProtocol) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.protocols = append(w.protocols, p)
}

// tracked returns the protocols of the run so far.
func (w *ProtocolWriter) tracked() []*logProtocol {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]*logProtocol(nil), w.protocols...)
}

func (w *ProtocolWriter) Close() {
	var c resultCounts

	protocols := w.tracked()
	for i := len(protocols) - 1; i >= 0; i-- {
		protocols[i].Close()
		c.add(protocols[i].p.TestResult)
	}

	events.TestRunFinished.Trigger(events.TestRunFinishedPayload{
		TestRunID: w.run.ID,
		TestRun:   w.run,
		TestName:  w.testName,
		ProjectID: w.projectId,
		AppID:     w.appId,
		Success:   !w.failed && c.failed == 0 && c.unstable == 0,
		Succeeded: c.succeeded,
		Unstable:  c.unstable,
		Failed:    c.failed,
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
