package teststore_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

func newTestDefinition(appID, deviceID uint) models.Test {
	return models.Test{AppID: appID, Name: "new", TestConfig: models.TestConfig{
		Type: models.TestTypeUnity, AllDevices: false, ExecutionType: models.SimultaneouslyExecutionType, RecordSession: true,
		Devices: []models.TestConfigDevice{{DeviceID: deviceID}},
		Unity: &models.TestConfigUnity{UnityTestCategoryType: models.SelectedTestsOnly, Categories: "a", UnityTestFunctions: []models.UnityTestFunction{
			{Class: "S", Method: "One"}, {Class: "S", Method: "Two"},
		}},
	}}
}

func TestCreateTestStoresTheWholeConfig(t *testing.T) {
	w := newWorld(t)

	created, err := w.store.CreateTest(bg(), newTestDefinition(w.scope.AppID, w.device.ID))

	if err != nil || created.ID == 0 || created.TestConfig.ID != 0 {
		t.Fatalf("created %+v, err %v; want the bare test row", created, err)
	}
	stored, err := w.store.Test(bg(), w.ref(created.ID))
	if err != nil {
		t.Fatal(err)
	}
	cfg := stored.TestConfig
	if !cfg.RecordSession || len(cfg.Devices) != 1 || cfg.Unity.Categories != "a" || len(cfg.Unity.UnityTestFunctions) != 2 {
		t.Fatalf("stored config %+v", cfg)
	}
}

func TestCreateTestWithoutUnityStoresNoUnitySettings(t *testing.T) {
	w := newWorld(t)
	def := newTestDefinition(w.scope.AppID, w.device.ID)
	def.TestConfig.Unity = nil

	created, err := w.store.CreateTest(bg(), def)

	stored, _ := w.store.Test(bg(), w.ref(created.ID))
	if err != nil || stored.TestConfig.Unity != nil {
		t.Fatalf("stored %+v, err %v", stored, err)
	}
}

func TestCreateTestRollsBackWhenAStepFails(t *testing.T) {
	steps := []string{"test_configs", "test_config_devices", "test_config_unities", "unity_test_functions"}
	for _, table := range steps {
		t.Run(table, func(t *testing.T) {
			w := newWorld(t)
			w.drop(t, table)

			_, err := w.store.CreateTest(bg(), newTestDefinition(w.scope.AppID, w.device.ID))

			wantKind(t, err, apperr.KindInternal)
			var n int64
			w.db.Model(&models.Test{}).Where("name = ?", "new").Count(&n)
			if n != 0 {
				t.Fatalf("%d tests stored by a failed create", n)
			}
		})
	}
}

func TestCreateTestMapsFailingTestInsert(t *testing.T) {
	w := newWorld(t)
	w.drop(t, "tests")

	_, err := w.store.CreateTest(bg(), newTestDefinition(w.scope.AppID, w.device.ID))

	wantKind(t, err, apperr.KindInternal)
}

func desired(deviceID uint) models.Test {
	return models.Test{Name: "renamed", TestConfig: models.TestConfig{
		AllDevices: true, ExecutionType: models.SimultaneouslyExecutionType, RecordSession: true,
		Devices: []models.TestConfigDevice{{DeviceID: deviceID}, {DeviceID: deviceID + 1}},
		Unity: &models.TestConfigUnity{UnityTestCategoryType: models.AllOfCategory, Categories: "smoke", UnityTestFunctions: []models.UnityTestFunction{
			{Class: "N", Method: "New"},
		}},
	}}
}

func TestUpdateTestReplacesDevicesAndFunctions(t *testing.T) {
	w := newWorld(t)
	current, _ := w.store.Test(bg(), w.ref(w.mine.ID))

	err := w.store.UpdateTest(bg(), current, desired(w.device.ID))

	saved, _ := w.store.Test(bg(), w.ref(w.mine.ID))
	cfg := saved.TestConfig
	if err != nil || saved.Name != "renamed" || !cfg.AllDevices || !cfg.RecordSession || len(cfg.Devices) != 2 {
		t.Fatalf("saved %+v, err %v", saved, err)
	}
	if fns := cfg.Unity.UnityTestFunctions; cfg.Unity.Categories != "smoke" || len(fns) != 1 || fns[0].Method != "New" {
		t.Fatalf("unity %+v", cfg.Unity)
	}
}

func TestUpdateTestWithoutUnitySettingsLeavesThemAlone(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(cur *models.Test, want *models.Test)
	}{
		{"nothing desired", func(_, want *models.Test) { want.TestConfig.Unity = nil }},
		{"nothing stored", func(cur, _ *models.Test) { cur.TestConfig.Unity = nil }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newWorld(t)
			current, _ := w.store.Test(bg(), w.ref(w.mine.ID))
			want := desired(w.device.ID)
			tc.prepare(current, &want)

			err := w.store.UpdateTest(bg(), current, want)

			saved, _ := w.store.Test(bg(), w.ref(w.mine.ID))
			fns := saved.TestConfig.Unity.UnityTestFunctions
			if err != nil || saved.Name != "renamed" || len(fns) != 1 || fns[0].Method != "Old" {
				t.Fatalf("saved %+v, functions %+v, err %v", saved, fns, err)
			}
		})
	}
}

func TestUpdateTestRollsBackWhenAStepFails(t *testing.T) {
	steps := []string{"tests", "test_configs", "test_config_devices", "test_config_unities", "unity_test_functions"}
	for _, table := range steps {
		t.Run(table, func(t *testing.T) {
			w := newWorld(t)
			current, _ := w.store.Test(bg(), w.ref(w.mine.ID))
			w.drop(t, table)

			err := w.store.UpdateTest(bg(), current, desired(w.device.ID))

			wantKind(t, err, apperr.KindInternal)
			var name string
			w.db.Raw("SELECT name FROM tests WHERE id = ?", w.mine.ID).Scan(&name)
			if table != "tests" && name != "mine" {
				t.Fatalf("name %q after a failed update, want it rolled back to mine", name)
			}
		})
	}
}

func TestUpdateTestKeepsTheOldDevicesWhenAnInsertFails(t *testing.T) {
	w := newWorld(t)
	current, _ := w.store.Test(bg(), w.ref(w.mine.ID))
	want := desired(w.device.ID)
	w.db.Exec("CREATE UNIQUE INDEX one_device ON test_config_devices (test_config_id, device_id)")
	want.TestConfig.Devices = []models.TestConfigDevice{{DeviceID: 5}, {DeviceID: 5}}

	err := w.store.UpdateTest(bg(), current, want)

	saved, _ := w.store.Test(bg(), w.ref(w.mine.ID))
	// a duplicate key is a conflict since the database translates it
	wantKind(t, err, apperr.KindConflict)
	if len(saved.TestConfig.Devices) != 1 || saved.TestConfig.Devices[0].DeviceID != w.device.ID {
		t.Fatalf("devices %+v after a failed update", saved.TestConfig.Devices)
	}
}
