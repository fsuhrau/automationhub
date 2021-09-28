package protocol

import (
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
	"time"
)

type logProtocol struct {
	p *models.TestProtocol
	l *LogWriter
}

type ProtocolWriter struct {
	db        *gorm.DB
	run       *models.TestRun
	protocols []logProtocol
}

func NewProtocolWriter(db *gorm.DB, run *models.TestRun) *ProtocolWriter {
	return &ProtocolWriter{db: db, run: run}
}

func (w *ProtocolWriter) NewProtocol(deviceID uint, testname string) (*LogWriter, error) {
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

	w.protocols = append(w.protocols, logProtocol{protocol, writer})

	events.NewTestProtocol.Trigger(events.NewTestProtocolPayload{TestRunID: w.run.ID, Protocol: protocol})

	return writer, nil
}

func (w *ProtocolWriter) RunID() uint {
	return w.run.ID
}

func (w *ProtocolWriter) SessionID() string {
	return w.run.SessionID
}

func (w *ProtocolWriter) Close() {
	for _, p := range w.protocols {
		var state models.TestResultState

		if len(p.l.errs) > 0 {
			state = models.TestResultFailed
		} else {
			state = models.TestResultSuccess
		}

		endTime := time.Now()
		p.p.TestResult = state
		p.p.EndedAt = &endTime
		w.db.Updates(&p.p)
	}
}
