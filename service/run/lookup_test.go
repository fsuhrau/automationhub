package run_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestDevicesListsEveryDeviceOfTheProject(t *testing.T) {
	e := newEnv()
	e.repo.devices = []models.Device{{DeviceIdentifier: "a"}, {DeviceIdentifier: "b"}}

	got, err := e.svc.Devices(bg(), e.project)

	if err != nil || len(got) != 2 || !e.repo.deviceQ.All || e.repo.deviceQ.Project != e.project {
		t.Fatalf("got %+v, err %v, query %+v", got, err, e.repo.deviceQ)
	}
}

func TestDeviceReturnsOneDevice(t *testing.T) {
	e := newEnv()
	e.repo.devices = []models.Device{{DeviceIdentifier: "a"}}

	got, err := e.svc.Device(bg(), e.project, 1)

	if err != nil || got.DeviceIdentifier != "a" {
		t.Fatalf("got %+v, err %v", got, err)
	}
	e.repo.fail["Device"] = apperr.NotFound("device")
	_, err = e.svc.Device(bg(), e.project, 2)
	assertKind(t, err, apperr.KindNotFound)
}

func TestProtocolReturnsAProtocolOfTheProject(t *testing.T) {
	e := newEnv()
	e.repo.protocol = &models.TestProtocol{TestName: "S/A"}

	got, err := e.svc.Protocol(bg(), 1, 5)

	if err != nil || got.TestName != "S/A" {
		t.Fatalf("got %+v, err %v", got, err)
	}
	e.repo.fail["ProjectProtocol"] = apperr.NotFound("protocol")
	_, err = e.svc.Protocol(bg(), 2, 5)
	assertKind(t, err, apperr.KindNotFound)
}

func TestLogsClampTheLimit(t *testing.T) {
	tests := []struct {
		name        string
		asked, want int
	}{
		{"default", 0, run.LogsDefault},
		{"negative", -3, run.LogsDefault},
		{"in range", 20, 20},
		{"too many", 5000, run.LogsMax},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := newEnv()
			e.repo.protocol = &models.TestProtocol{}
			e.repo.logs = []models.ProtocolEntry{{Message: "a"}}
			e.repo.logsTotal = 42

			page, err := e.svc.Logs(bg(), run.LogQuery{ProjectID: 1, ProtocolID: 5, Offset: 7, Limit: tc.asked, Level: "error"})

			if err != nil || page.Total != 42 || page.Offset != 7 || len(page.Entries) != 1 {
				t.Fatalf("page %+v, err %v", page, err)
			}
			if q := e.repo.logQuery; q.Limit != tc.want || q.Level != "error" || q.ProtocolID != 5 {
				t.Errorf("query %+v, want limit %d", q, tc.want)
			}
		})
	}
}

func TestLogsErrors(t *testing.T) {
	for _, fail := range []string{"ProjectProtocol", "ProtocolLogs"} {
		t.Run(fail, func(t *testing.T) {
			e := newEnv()
			e.repo.fail[fail] = apperr.NotFound("x")

			page, err := e.svc.Logs(bg(), run.LogQuery{ProjectID: 1, ProtocolID: 5})

			assertKind(t, err, apperr.KindNotFound)
			if page != nil {
				t.Fatalf("page %+v with an error", page)
			}
		})
	}
}
