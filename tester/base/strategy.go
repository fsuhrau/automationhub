package base

import (
	"context"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Strategy is what a type of runner adds to the pipeline: which tests to run
// and how to run one of them on a device.
type Strategy interface {
	// TestList returns the tests of the run, asked for once the apps are
	// connected.
	TestList(ctx context.Context, connected []DeviceMap) ([]models.UnityTestFunction, error)
	// RunTest runs one test on one device and reports its result.
	RunTest(ctx context.Context, dev DeviceMap, task action.TestStart)
}

// Preparer is implemented by strategies that prepare the devices before the
// app is installed, e.g. by uploading it.
type Preparer interface {
	Prepare(ctx context.Context, params app.Parameter, devices []DeviceMap)
}

// StartObserver is implemented by strategies that react once the app of a
// device is started, before it is connected.
type StartObserver interface {
	AppStarted(ctx context.Context, d device.Device)
}
