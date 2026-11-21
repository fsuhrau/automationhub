package runstore_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/run"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestStartTestLoadsTheTestWithItsConfigAndUnitySettings(t *testing.T) {
	w := newWorld(t)
	cfg := models.TestConfig{TestID: w.mine.ID, Type: models.TestTypeUnity, AllDevices: true}
	w.db.Create(&cfg)
	unity := models.TestConfigUnity{TestConfigID: cfg.ID}
	w.db.Create(&unity)
	w.db.Create(&models.UnityTestFunction{TestConfigUnityID: unity.ID, Method: "Smoke"})

	got, err := w.store.StartTest(bg(), w.mineProject.ID, w.mine.ID)

	if err != nil || got.App == nil || got.App.ID != w.mineApp.ID || got.TestConfig.ID != cfg.ID {
		t.Fatalf("test %+v, err %v", got, err)
	}
	if got.TestConfig.Unity == nil || len(got.TestConfig.Unity.UnityTestFunctions) != 1 {
		t.Fatalf("unity settings %+v", got.TestConfig.Unity)
	}
}

func TestStartTestOfAnotherProjectIsNotFound(t *testing.T) {
	w := newWorld(t)

	_, err := w.store.StartTest(bg(), w.mineProject.ID, w.foreign.ID)

	assertKind(t, err, apperr.KindNotFound, "not found: test "+itoa(w.foreign.ID))
}

func TestStartTestReportsADatabaseError(t *testing.T) {
	_, err := closed(t).StartTest(bg(), 1, 1)

	assertKind(t, err, apperr.KindInternal, "")
}

func TestBundleBelongsToTheApp(t *testing.T) {
	w := newWorld(t)
	own := models.AppBinary{AppID: w.mineApp.ID, Name: "Game"}
	other := models.AppBinary{AppID: w.foreApp.ID, Name: "Other"}
	w.db.Create(&own)
	w.db.Create(&other)

	got, err := w.store.Bundle(bg(), w.mineApp.ID, own.ID)
	if err != nil || got.ID != own.ID || got.App == nil {
		t.Fatalf("bundle %+v, err %v", got, err)
	}
	_, err = w.store.Bundle(bg(), w.mineApp.ID, other.ID)
	assertKind(t, err, apperr.KindNotFound, "not found: app bundle "+itoa(other.ID)+" for app "+itoa(w.mineApp.ID))
}

func TestDevicesListsWhatTheProjectCanSeeWithTheirNode(t *testing.T) {
	w := newWorld(t)
	node := models.Node{Identifier: "n", Name: "Node"}
	w.db.Create(&node)
	w.db.Model(w.device).Update("node_id", node.ID)
	company := &models.Project{CompanyID: 1}
	w.db.Create(&models.Device{CompanyID: 1, DeviceIdentifier: "mine", Name: "mine"})
	w.db.Create(&models.Device{CompanyID: 2, DeviceIdentifier: "theirs", Name: "theirs"})

	all, err := w.store.Devices(bg(), run.DeviceQuery{Project: company, All: true})

	if err != nil || len(all) != 2 || all[0].ID != w.device.ID || all[0].Node == nil || all[0].Node.Name != "Node" || all[1].DeviceIdentifier != "mine" {
		t.Fatalf("devices %+v, err %v; want the shared and the own device by id", all, err)
	}
	some, err := w.store.Devices(bg(), run.DeviceQuery{Project: company, IDs: []uint{all[1].ID, 999}})
	if err != nil || len(some) != 1 || some[0].ID != all[1].ID {
		t.Fatalf("listed devices %+v, err %v", some, err)
	}
	none, err := w.store.Devices(bg(), run.DeviceQuery{Project: company})
	if err != nil || len(none) != 0 {
		t.Fatalf("an empty list is no devices, got %+v, err %v", none, err)
	}
}

func TestDeviceIsOnlyFoundWhenTheProjectCanSeeIt(t *testing.T) {
	w := newWorld(t)
	theirs := models.Device{CompanyID: 2, DeviceIdentifier: "theirs"}
	w.db.Create(&theirs)

	got, err := w.store.Device(bg(), w.mineProject, w.device.ID)
	if err != nil || got.ID != w.device.ID {
		t.Fatalf("device %+v, err %v", got, err)
	}
	_, err = w.store.Device(bg(), w.mineProject, theirs.ID)
	assertKind(t, err, apperr.KindNotFound, "not found: device "+itoa(theirs.ID))
}

func TestRunInProjectChecksTheProject(t *testing.T) {
	w := newWorld(t)

	if err := w.store.RunInProject(bg(), w.mineProject.ID, w.run1.ID); err != nil {
		t.Fatal(err)
	}
	err := w.store.RunInProject(bg(), w.mineProject.ID, w.foreignRun.ID)
	assertKind(t, err, apperr.KindNotFound, "not found: test run "+itoa(w.foreignRun.ID))
}

func TestScopeChecksReportADatabaseError(t *testing.T) {
	s := closed(t)
	w := &world{}
	for name, err := range map[string]error{
		"run in project": s.RunInProject(bg(), 1, 1),
		"test in scope":  s.TestInScope(bg(), w.ref(&models.Test{}, &models.App{}, &models.Project{})),
		"run in test":    s.RunInTest(bg(), 1, 1),
	} {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("%s: %v", name, err)
		}
	}
}
