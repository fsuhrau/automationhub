package storeerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
	"gorm.io/gorm"
)

func TestMap(t *testing.T) {
	boom := errors.New("disk full")
	tests := []struct {
		name     string
		err      error
		wantKind apperr.Kind
		wantMsg  string
	}{
		{"not found", gorm.ErrRecordNotFound, apperr.KindNotFound, "test run 3 not found: record not found"},
		{"wrapped not found", fmt.Errorf("load: %w", gorm.ErrRecordNotFound), apperr.KindNotFound, "test run 3 not found: load: record not found"},
		{"duplicate", gorm.ErrDuplicatedKey, apperr.KindConflict, "test run 3 exists already: duplicated key not allowed"},
		{"reference", gorm.ErrForeignKeyViolated, apperr.KindConflict, "test run 3 is still in use: violates foreign key constraint"},
		{"invalid data", gorm.ErrInvalidData, apperr.KindInvalid, "test run 3 has invalid data: unsupported data"},
		{"invalid value", gorm.ErrInvalidValue, apperr.KindInvalid, "test run 3 has an invalid value: invalid value, should be pointer to struct or slice"},
		{"other", boom, apperr.KindInternal, "test run 3: disk full"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := storeerr.Map(tc.err, "test run 3")
			if apperr.KindOf(got) != tc.wantKind || got.Error() != tc.wantMsg {
				t.Fatalf("got %v (%v), want %v %q", got, apperr.KindOf(got), tc.wantKind, tc.wantMsg)
			}
			if !errors.Is(got, tc.err) {
				t.Fatal("cause is not reachable")
			}
		})
	}
}

func TestMapKeepsNilAndKnownErrors(t *testing.T) {
	if storeerr.Map(nil, "x") != nil {
		t.Fatal("nil was mapped")
	}
	known := apperr.Busy("device 1")
	if got := storeerr.Map(known, "x"); got != known {
		t.Fatalf("known error was changed: %v", got)
	}
}

func TestMapClassifiesARealMissingRow(t *testing.T) {
	db := testkit.DB(t)
	var run models.TestRun

	err := storeerr.Map(db.First(&run, 99).Error, "test run 99")

	if apperr.KindOf(err) != apperr.KindNotFound {
		t.Fatalf("kind = %v, err %v", apperr.KindOf(err), err)
	}
}
