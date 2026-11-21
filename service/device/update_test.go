package device_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/models"
)

func withParams(params ...models.CustomParameter) models.Device {
	d := stored(1, "dev-1")
	d.CustomParameter = params
	return d
}

func cp(id uint, key string) models.CustomParameter {
	return models.CustomParameter{Model: models.Model{ID: id}, DeviceID: 1, Key: key}
}

func TestUpdatePlansTheParameterChange(t *testing.T) {
	e := newEnv(withParams(cp(10, "PIN"), cp(11, "LANG"), cp(12, "GONE")))

	_, err := e.svc.Update(context.Background(), e.project, svc.Update{DeviceID: "1",
		Alias: "  Lab  ",
		CustomParameter: []svc.Param{
			{ID: 10, Key: "PIN", Value: "9"},
			{ID: 11, Key: " LOCALE ", Value: "en"},
			{Key: "NEW", Value: "n"},
			{ID: 99, Key: "STOLEN", Value: "s"}, // another device's parameter
		},
	})

	if err != nil {
		t.Fatal(err)
	}
	want := svc.Change{
		Alias:  "Lab",
		Update: []svc.Param{{ID: 10, Key: "PIN", Value: "9"}, {ID: 11, Key: "LOCALE", Value: "en"}},
		Add:    []svc.Param{{Key: "NEW", Value: "n"}, {ID: 99, Key: "STOLEN", Value: "s"}},
		Remove: []uint{12},
	}
	if len(e.repo.applied) != 1 || !reflect.DeepEqual(e.repo.applied[0], want) {
		t.Fatalf("applied %+v, want %+v", e.repo.applied, want)
	}
}

func TestUpdateRejectsBadParametersBeforeTouchingStorage(t *testing.T) {
	tests := []struct {
		name   string
		params []svc.Param
		want   string
	}{
		{"empty key", []svc.Param{{Key: " "}}, "custom parameter 1 has no key"},
		{"second has no key", []svc.Param{{Key: "A"}, {}}, "custom parameter 2 has no key"},
		{"duplicate key", []svc.Param{{Key: "A"}, {Key: " A "}}, `custom parameter "A" is there twice`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(stored(1, "dev-1"))

			_, err := e.svc.Update(context.Background(), e.project, svc.Update{DeviceID: "1", CustomParameter: tc.params})

			assertKind(t, err, apperr.KindInvalid)
			if err.Error() != tc.want || e.repo.finds != 0 || len(e.repo.applied) != 0 {
				t.Fatalf("err %q finds %d applied %d", err, e.repo.finds, len(e.repo.applied))
			}
		})
	}
}

func TestUpdateStorageErrors(t *testing.T) {
	tests := []struct {
		name string
		id   string
		fail map[string]error
		want apperr.Kind
	}{
		{"unknown device", "2", nil, apperr.KindNotFound},
		{"apply fails", "1", map[string]error{"Apply": apperr.Internal("x")}, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv(stored(1, "dev-1"))
			e.repo.fail = tc.fail

			_, err := e.svc.Update(context.Background(), e.project, svc.Update{DeviceID: tc.id})

			assertKind(t, err, tc.want)
		})
	}
}
