package device_test

import (
	"context"
	"errors"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/fsuhrau/automationhub/hub/control"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

var errBoom = errors.New("boom")

// fakeRepo serves the devices it holds to the project they belong to
// (CompanyID 0 is visible to all) and records what it was asked to do.
type fakeRepo struct {
	devices   []models.Device
	protocols []models.TestProtocol
	bundle    *models.AppBinary
	fail      map[string]error

	projects []*models.Project
	applied  []svc.Change
	deleted  []uint
	limit    int
	platform *models.PlatformType
	finds    int
}

func (r *fakeRepo) err(method string) error { return r.fail[method] }

func (r *fakeRepo) List(_ context.Context, p *models.Project, platform *models.PlatformType) ([]models.Device, error) {
	r.projects, r.platform = append(r.projects, p), platform
	return append([]models.Device(nil), r.devices...), r.err("List")
}

func (r *fakeRepo) Find(_ context.Context, p *models.Project, id string) (*models.Device, error) {
	r.projects = append(r.projects, p)
	r.finds++
	if err := r.err("Find"); err != nil {
		return nil, err
	}
	for i := range r.devices {
		if r.devices[i].ID == uintID(id) {
			d := r.devices[i]
			return &d, nil
		}
	}
	return nil, apperr.NotFound("device %s not found", id)
}

func (r *fakeRepo) FindDetail(ctx context.Context, p *models.Project, id string) (*models.Device, error) {
	if err := r.err("FindDetail"); err != nil {
		return nil, err
	}
	return r.Find(ctx, p, id)
}

func (r *fakeRepo) Delete(_ context.Context, d *models.Device) error {
	r.deleted = append(r.deleted, d.ID)
	return r.err("Delete")
}

func (r *fakeRepo) Apply(_ context.Context, _ uint, c svc.Change) error {
	r.applied = append(r.applied, c)
	return r.err("Apply")
}

func (r *fakeRepo) Protocols(_ context.Context, scope svc.ProtocolScope) ([]models.TestProtocol, error) {
	r.limit = scope.Limit
	return r.protocols, r.err("Protocols")
}

func (r *fakeRepo) Bundle(context.Context, *models.Project, uint) (*models.AppBinary, error) {
	return r.bundle, r.err("Bundle")
}

func uintID(id string) uint {
	var n uint
	for _, c := range id {
		n = n*10 + uint(c-'0')
	}
	return n
}

type fakeLeases struct {
	holders  map[string]string
	released []string
}

func (l *fakeLeases) Holder(id string) (control.Lease, bool) {
	owner, ok := l.holders[id]
	return control.Lease{DeviceID: id, Owner: owner}, ok
}

func (l *fakeLeases) ForceRelease(id string) bool {
	l.released = append(l.released, id)
	return true
}

type fakeExecutor struct {
	err   error
	tests []action.TestStart
	dev   device.Device
	ctx   context.Context
}

func (e *fakeExecutor) Execute(ctx context.Context, dev device.Device, t action.TestStart, _ time.Duration) error {
	e.ctx, e.dev = ctx, dev
	e.tests = append(e.tests, t)
	return e.err
}
