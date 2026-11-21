package base

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	hubsync "github.com/fsuhrau/automationhub/utils/sync"
)

// installAttempts is how often an installation is tried; a failed attempt
// uninstalls the app first.
const installAttempts = 2

// installApps installs the app on the devices that don't have it.
func (r *Runner) installApps(ctx context.Context, params app.Parameter, devices []DeviceMap) {
	forEachDevice(ctx, devices, func(d DeviceMap) {
		p := params
		installed, err := d.Device.IsAppInstalled(&p)
		if err != nil {
			r.LogError("check installation failed: %v", err)
			return
		}
		if !installed {
			r.installWithRetry(d.Device, &p)
		}
	})
}

func (r *Runner) installWithRetry(d device.Device, p *app.Parameter) {
	var err error
	for range installAttempts {
		if err = d.InstallApp(p); err == nil {
			return
		}
		r.LogInfo("installation failed try delete first: %v", err)
		_ = d.UninstallApp(p)
	}
	r.LogError("unable to install app: %v", err)
}

// stopApps stops the app on the devices and clears its data.
func (r *Runner) stopApps(ctx context.Context, params app.Parameter, devices []DeviceMap) {
	forEachDevice(ctx, devices, func(d DeviceMap) {
		p := params
		if err := StopAndClearApp(d.Device, &p); err != nil {
			r.LogError("unable to stop app: %v", err)
		}
	})
}

// appStart is what starting the app on one device needs.
type appStart struct {
	params   app.Parameter
	observer StartObserver
	out      chan<- DeviceMap
}

// startApps starts the app on the devices and returns those whose app
// connected in time. Apps that did not connect are stopped before it returns.
func (r *Runner) startApps(ctx context.Context, s *runState) ([]DeviceMap, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	observer, _ := r.strategy.(StartObserver)
	var inflight sync.WaitGroup
	// buffered for every device, so late devices never block
	connected := make(chan DeviceMap, len(s.devices))
	for _, d := range s.devices {
		inflight.Add(1)
		go func() {
			defer inflight.Done()
			r.startApp(ctx, d, appStart{params: s.params, observer: observer, out: connected})
		}()
	}
	err := r.waitForStart(ctx, &inflight)
	cancel()
	inflight.Wait()
	return drain(connected), err
}

// waitForStart waits until all apps are started or connected, the timeout
// ends or the context does.
func (r *Runner) waitForStart(ctx context.Context, inflight *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		inflight.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-r.Clock.After(r.Timeouts.Start):
		return hubsync.ErrTimeout
	}
}

func drain(connected chan DeviceMap) []DeviceMap {
	var out []DeviceMap
	for {
		select {
		case d := <-connected:
			out = append(out, d)
		default:
			return out
		}
	}
}

func (r *Runner) startApp(ctx context.Context, d DeviceMap, st appStart) {
	began := r.Clock.Now()
	session := r.ProtocolWriter.SessionID()
	r.LogInfo("Start App '%s' on Device '%s' with Session '%s'", st.params.Identifier, d.Device.DeviceID(), session)
	if err := d.Device.StartApp(nil, &st.params, session, r.NodeURL); err != nil {
		r.LogError("unable to start app: %v", err)
		return
	}
	if st.observer != nil {
		st.observer.AppStarted(ctx, d.Device)
	}
	if !r.awaitConnection(ctx, d.Device, &st.params) {
		return
	}
	r.trackStartupTime(d.Model.ID, r.Clock.Now().Sub(began))
	st.out <- d
}

// trackStartupTime is serialized: the devices connect in parallel, and the
// protocol writer appends to the run.
func (r *Runner) trackStartupTime(deviceID uint, took time.Duration) {
	r.startupMu.Lock()
	defer r.startupMu.Unlock()
	r.ProtocolWriter.TrackStartupTime(deviceID, took.Milliseconds())
}

// awaitConnection waits until the app of the device is connected; it stops
// the app when the context ends first.
func (r *Runner) awaitConnection(ctx context.Context, d device.Device, p *app.Parameter) bool {
	for !d.IsAppConnected() {
		select {
		case <-ctx.Done():
			_ = StopAndClearApp(d, p)
			r.LogInfo("Context cancelled, stopping app start process")
			return false
		case <-r.Clock.After(r.Timeouts.Poll):
		}
	}
	return true
}

// StopAndClearApp stops the app and, where the platform supports it, wipes its
// data so the next run starts from a clean state. Test runs rely on this;
// interactive control uses plain StopApp.
func StopAndClearApp(d device.Device, params *app.Parameter) error {
	if err := d.StopApp(params); err != nil {
		return err
	}
	// the stopped app's connection is dead, even if its socket looks open
	device.DropAppConnection(d)
	if clearer, ok := d.(device.AppDataClearer); ok {
		if err := clearer.ClearAppData(params); err != nil && !errors.Is(err, device.ErrUnsupported) {
			return err
		}
	}
	return nil
}
