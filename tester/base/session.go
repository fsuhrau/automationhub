package base

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"github.com/gofrs/uuid"
)

func (tr *TestRunner) NewSessionID() string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s", u)
}

func (tr *TestRunner) InitNewTestSession(appId uint, params string) error {
	sessionID := tr.NewSessionID()
	tr.TestRun = models.TestRun{
		TestID:    tr.Test.ID,
		AppID:     appId,
		SessionID: sessionID,
		Parameter: params,
	}
	if err := tr.DB.Create(&tr.TestRun).Error; err != nil {
		return err
	}

	tr.ProtocolWriter = protocol.NewProtocolWriter(tr.DB, tr.Test.Name, &tr.TestRun)
	return nil
}

func (tr *TestRunner) TestSessionFinished() {
	tr.ProtocolWriter.Close()
}
