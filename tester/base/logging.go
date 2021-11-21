package base

import (
	"fmt"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/storage/models"
)

func (tr *TestRunner) LogInfo(format string, params ...interface{}) {
	logEntry := &models.TestRunLogEntry{
		TestRunID: tr.ProtocolWriter.RunID(),
		Level:     "log",
		Log:       fmt.Sprintf(format, params...),
	}
	tr.DB.Create(logEntry)
	events.NewTestLogEntry.Trigger(events.NewTestLogEntryPayload{
		logEntry.TestRunID,
		logEntry,
	})
}

func (tr *TestRunner) LogError(format string, params ...interface{}) {
	tr.Err = fmt.Errorf(format, params)
	logEntry := &models.TestRunLogEntry{
		TestRunID: tr.ProtocolWriter.RunID(),
		Level:     "error",
		Log:       fmt.Sprintf(format, params...),
	}
	tr.DB.Create(logEntry)

	events.NewTestLogEntry.Trigger(events.NewTestLogEntryPayload{
		logEntry.TestRunID,
		logEntry,
	})
}