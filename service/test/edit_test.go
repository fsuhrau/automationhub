package test_test

import (
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	service "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Every rule of the edit page is checked through Create, which validates
// like Update does for a Unity test.
func TestEditValidation(t *testing.T) {
	tests := []struct {
		name   string
		change func(*service.Edit)
		kind   apperr.Kind
		msg    string
	}{
		{"valid", func(*service.Edit) {}, apperr.KindNone, ""},
		{"no name", func(e *service.Edit) { e.Name = " " }, apperr.KindInvalid, "the test needs a name"},
		{"unknown execution", func(e *service.Edit) { e.ExecutionType = 7 }, apperr.KindInvalid, "unknown execution type 7"},
		{"no devices", func(e *service.Edit) { e.Devices = nil }, apperr.KindInvalid, "pick at least one device, or run on all devices"},
		{"all devices", func(e *service.Edit) { e.AllDevices, e.Devices = true, nil }, apperr.KindNone, ""},
		{"no category", func(e *service.Edit) { e.Categories = " , " }, apperr.KindInvalid, "name at least one category"},
		{"no tests picked", func(e *service.Edit) {
			e.UnityTestCategoryType, e.TestFunctions = models.SelectedTestsOnly, nil
		}, apperr.KindInvalid, "pick at least one test"},
		{"tests picked", func(e *service.Edit) {
			e.UnityTestCategoryType = models.SelectedTestsOnly
		}, apperr.KindNone, ""},
		{"all tests", func(e *service.Edit) { e.UnityTestCategoryType = models.AllTest }, apperr.KindNone, ""},
		{"unknown selection", func(e *service.Edit) { e.UnityTestCategoryType = 9 }, apperr.KindInvalid, "unknown Unity test selection 9"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFake()
			repo.visible = 2
			edit := validEdit()
			tc.change(&edit)

			_, err := service.New(service.Deps{Repo: repo}).Create(bg(), mine(), service.NewTest{Type: models.TestTypeUnity, Edit: edit})

			wantKind(t, err, tc.kind)
			if tc.msg != "" && err.Error() != tc.msg {
				t.Fatalf("message %q, want %q", err, tc.msg)
			}
		})
	}
}
