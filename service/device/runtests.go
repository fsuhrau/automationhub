package device

import (
	"context"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/events"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
)

// testTimeout is how long a test started from the device page may run.
const testTimeout = 5 * time.Minute

// TestRequest starts one test of the app on the device with DeviceID:
// TestName is "Class Method", Env the run parameters.
type TestRequest struct {
	DeviceID string `json:"-"`
	TestName string `json:"testName"`
	Env      string `json:"env"`
}

// RunTests runs the test on the device and returns when it has finished.
// The device is locked for the run and unlocked afterwards. A client that
// goes away doesn't stop the run.
func (s *Service) RunTests(ctx context.Context, project *models.Project, req TestRequest) error {
	test, err := s.parseTest(req)
	if err != nil {
		return err
	}
	d, err := s.deps.Repo.Find(ctx, project, req.DeviceID)
	if err != nil {
		return err
	}
	dev, _ := s.deps.Devices.GetDevice(d.DeviceIdentifier)
	if dev == nil {
		return apperr.NotFound("real device not found")
	}
	return s.runLocked(context.WithoutCancel(ctx), dev, test)
}

// parseTest reads "Class Method" and the parameters of a request.
func (s *Service) parseTest(req TestRequest) (action.TestStart, error) {
	parts := strings.Split(req.TestName, " ")
	if len(parts) != 2 {
		return action.TestStart{}, apperr.Invalid("invalid method signature")
	}
	return action.TestStart{Class: parts[0], Method: parts[1], Env: s.deps.ExtractParams(req.Env)}, nil
}

// runLocked runs the test with the device locked, and only when the app
// of the device is connected.
func (s *Service) runLocked(ctx context.Context, dev device.Device, test action.TestStart) error {
	if err := events.LockDevice(dev); err != nil {
		return apperr.NotFound("%v", err)
	}
	defer func() { _ = events.UnlockDevice(dev) }()

	if !dev.IsAppConnected() {
		return apperr.NotFound("no app connected")
	}
	// a failed reset doesn't stop the run, as before
	_ = s.deps.Devices.SendAction(dev, &action.UnityReset{})
	if err := s.deps.NewExecutor().Execute(ctx, dev, test, testTimeout); err != nil {
		return apperr.Internal("%v", err)
	}
	return nil
}
