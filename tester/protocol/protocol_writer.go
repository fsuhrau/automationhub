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
	p.db.Updates(&p.p)
}

type ProtocolWriter struct {
	db        *gorm.DB
	run       *models.TestRun
	protocols []*logProtocol
}

func NewProtocolWriter(db *gorm.DB, run *models.TestRun) *ProtocolWriter {
	return &ProtocolWriter{db: db, run: run}
}

func (w *ProtocolWriter) NewProtocol(deviceID uint, testname string) (*logProtocol, error) {
	protocol := &models.TestProtocol{
		TestRunID: w.run.ID,
		DeviceID:  &deviceID,
		TestName:  testname,
		StartedAt: time.Now(),
	}

	if err := w.db.Create(protocol).Error; err != nil {
		return nil, err
	}

	writer := NewLogWriter(w.db, protocol.ID)

	p := &logProtocol{w.db,protocol, writer}
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
	failedCount := 0
	successCount := 0
	unstableCount := 0
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
		TestRun: w.run,
		Success: failedCount == 0 && unstableCount == 0,
		Succeeded: successCount,
		Unstable: unstableCount,
		Failed: failedCount,
	})
}
