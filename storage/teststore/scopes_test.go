package teststore_test

import (
	"reflect"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/teststore"
	"gorm.io/gorm"
)

// protocolFixture is a run with a runner protocol around one test protocol
// and a second run with a lone test protocol.
type protocolFixture struct {
	db                *gorm.DB
	runner, inner, lo *models.TestProtocol
}

func newProtocolFixture(t *testing.T) *protocolFixture {
	t.Helper()
	db := testkit.DB(t)
	run1, run2 := testkit.TestRun(db), testkit.TestRun(db)
	f := &protocolFixture{db: db}
	f.runner = testkit.Protocol(db, func(p *models.TestProtocol) { p.TestRunID = run1.ID })
	f.inner = testkit.Protocol(db, func(p *models.TestProtocol) { p.TestRunID, p.ParentTestProtocolID = run1.ID, &f.runner.ID })
	f.lo = testkit.Protocol(db, func(p *models.TestProtocol) { p.TestRunID = run2.ID })
	return f
}

func (f *protocolFixture) ids(scopes ...func(*gorm.DB) *gorm.DB) []uint {
	var got []uint
	f.db.Model(&models.TestProtocol{}).Scopes(scopes...).Order("id").Pluck("id", &got)
	return got
}

func TestLeafProtocols(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*protocolFixture)
		want  func(*protocolFixture) []uint
	}{
		{"runner protocols are left out", func(*protocolFixture) {},
			func(f *protocolFixture) []uint { return []uint{f.inner.ID, f.lo.ID} }},
		{"a runner whose children were deleted is a leaf again",
			func(f *protocolFixture) { f.db.Delete(f.inner) },
			func(f *protocolFixture) []uint { return []uint{f.runner.ID, f.lo.ID} }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newProtocolFixture(t)
			tc.setup(f)

			if got, want := f.ids(teststore.LeafProtocols), tc.want(f); !reflect.DeepEqual(got, want) {
				t.Fatalf("leaf protocols %v, want %v", got, want)
			}
		})
	}
}

func TestInRunsLimitsToTheRuns(t *testing.T) {
	f := newProtocolFixture(t)

	got := f.ids(teststore.InRuns([]uint{f.lo.TestRunID}))

	if !reflect.DeepEqual(got, []uint{f.lo.ID}) {
		t.Fatalf("protocols of the run %v, want [%d]", got, f.lo.ID)
	}
}

func TestScopesCombine(t *testing.T) {
	f := newProtocolFixture(t)

	got := f.ids(teststore.LeafProtocols, teststore.InRuns([]uint{f.inner.TestRunID}))

	if !reflect.DeepEqual(got, []uint{f.inner.ID}) {
		t.Fatalf("leaf protocols of run 1 %v, want [%d]", got, f.inner.ID)
	}
}
