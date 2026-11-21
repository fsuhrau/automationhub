package base

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/protocol"
	"github.com/gofrs/uuid"
)

func newSessionID() string {
	u, _ := uuid.NewV4()
	return u.String()
}

// Session is what a run is started with.
type Session struct {
	AppBinaryID uint
	StartURL    string
	// Parameter are the environment lines of the run.
	Parameter string
}

// InitNewTestSession creates the run of the test and its protocol writer.
func (r *Runner) InitNewTestSession(ctx context.Context, session Session) error {
	r.TestRun = models.TestRun{
		TestID:      r.Test.ID,
		AppBinaryID: session.AppBinaryID,
		SessionID:   newSessionID(),
		StartURL:    session.StartURL,
		Parameter:   session.Parameter,
	}
	if err := r.Store.CreateRun(ctx, &r.TestRun); err != nil {
		return fmt.Errorf("init test session: %w", err)
	}
	r.ProtocolWriter = protocol.NewProtocolWriter(r.DB, r.ProjectID, r.AppID, r.Test.Name, &r.TestRun)
	return nil
}

// TestSessionFinished closes the protocols and announces the result.
func (r *Runner) TestSessionFinished() {
	r.ProtocolWriter.Close()
}
