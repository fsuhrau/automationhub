package devicestore_test

import (
	"strconv"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/internal/testkit"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/fsuhrau/automationhub/storage/devicestore"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

func inCompany(id uint) testkit.Option[models.Device] {
	return func(d *models.Device) { d.CompanyID = id }
}

func onPlatform(p models.PlatformType) testkit.Option[models.Device] {
	return func(d *models.Device) { d.PlatformType = p }
}

func idOf(d *models.Device) string { return strconv.Itoa(int(d.ID)) }

func companyProject(db *gorm.DB, company uint) *models.Project {
	return testkit.Project(db, func(p *models.Project) { p.CompanyID = company })
}

func assertKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if got := apperr.KindOf(err); got != want {
		t.Fatalf("kind = %v (%v), want %v", got, err, want)
	}
}

func TestListScopesToTheProjectsCompany(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	shared := testkit.Device(db)
	mine := testkit.Device(db, inCompany(1))
	testkit.Device(db, inCompany(2))

	got, err := devicestore.New(db).List(ctx, companyProject(db, 1), nil)

	if err != nil || len(got) != 2 || got[0].ID != shared.ID || got[1].ID != mine.ID {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestListFiltersByPlatformAndLoadsTheNode(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	node := models.Node{Name: "lab"}
	db.Create(&node)
	android := testkit.Device(db, onPlatform(models.PlatformTypeAndroid), func(d *models.Device) { d.NodeID = node.ID })
	testkit.Device(db, onPlatform(models.PlatformTypeiOS))
	platform := models.PlatformTypeAndroid

	got, err := devicestore.New(db).List(ctx, companyProject(db, 0), &platform)

	if err != nil || len(got) != 1 || got[0].ID != android.ID || got[0].Node == nil || got[0].Node.Name != "lab" {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestListFailureIsInternal(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	dropTable(t, db, &models.Device{})

	_, err := devicestore.New(db).List(ctx, companyProject(db, 0), nil)

	assertKind(t, err, apperr.KindInternal)
}

func TestFindIsScopedAndFindDetailLoadsParameters(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	mine := testkit.Device(db, inCompany(1), func(d *models.Device) {
		d.CustomParameter = []models.CustomParameter{{Key: "PIN", Value: "1"}}
	})
	foreign := testkit.Device(db, inCompany(2))
	store, project := devicestore.New(db), companyProject(db, 1)

	plain, err := store.Find(ctx, project, idOf(mine))
	if err != nil || plain.CustomParameter != nil {
		t.Fatalf("Find %+v err %v", plain, err)
	}
	detail, err := store.FindDetail(ctx, project, idOf(mine))
	if err != nil || len(detail.CustomParameter) != 1 {
		t.Fatalf("FindDetail %+v err %v", detail, err)
	}
	_, err = store.Find(ctx, project, idOf(foreign))
	assertKind(t, err, apperr.KindNotFound)
	if err.Error() != "device "+idOf(foreign)+" not found" {
		t.Fatalf("message %q", err)
	}
}

func TestDelete(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	d := testkit.Device(db)
	store := devicestore.New(db)

	if err := store.Delete(ctx, d); err != nil {
		t.Fatal(err)
	}
	_, err := store.Find(ctx, companyProject(db, 0), idOf(d))
	assertKind(t, err, apperr.KindNotFound)

	dropTable(t, db, &models.Device{})
	assertKind(t, store.Delete(ctx, d), apperr.KindInternal)
}

func TestApplyWritesAliasAndParameters(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	d := testkit.Device(db, func(d *models.Device) {
		d.CustomParameter = []models.CustomParameter{{Key: "PIN", Value: "1"}, {Key: "GONE", Value: "x"}}
	})
	other := testkit.Device(db, func(d *models.Device) {
		d.CustomParameter = []models.CustomParameter{{Key: "OTHER", Value: "o"}}
	})
	pin, gone, foreign := d.CustomParameter[0].ID, d.CustomParameter[1].ID, other.CustomParameter[0].ID

	err := devicestore.New(db).Apply(ctx, d.ID, svc.Change{
		Alias:  "Lab",
		Update: []svc.Param{{ID: pin, Key: "PIN", Value: "9"}, {ID: foreign, Key: "HACK", Value: "h"}},
		Add:    []svc.Param{{Key: "NEW", Value: "n"}},
		Remove: []uint{gone, foreign},
	})

	if err != nil {
		t.Fatal(err)
	}
	if got := paramsOf(db, d.ID); got != "NEW=n PIN=9" {
		t.Errorf("parameters %q", got)
	}
	if got := paramsOf(db, other.ID); got != "OTHER=o" {
		t.Errorf("another device's parameters changed: %q", got)
	}
	var saved models.Device
	db.First(&saved, d.ID)
	if saved.Alias != "Lab" {
		t.Errorf("alias %q", saved.Alias)
	}
}

func paramsOf(db *gorm.DB, deviceID uint) string {
	var list []models.CustomParameter
	db.Where("device_id = ?", deviceID).Order("key").Find(&list)
	out := ""
	for i, p := range list {
		if i > 0 {
			out += " "
		}
		out += p.Key + "=" + p.Value
	}
	return out
}

func TestApplyFailuresAreInternal(t *testing.T) {
	ctx := t.Context()
	changes := map[string]svc.Change{
		"alias":  {},
		"update": {Update: []svc.Param{{ID: 1}}},
		"add":    {Add: []svc.Param{{Key: "k"}}},
		"remove": {Remove: []uint{1}},
	}
	drop := map[string]any{
		"alias":  &models.Device{},
		"update": &models.CustomParameter{},
		"add":    &models.CustomParameter{},
		"remove": &models.CustomParameter{},
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			db := testkit.DB(t)
			d := testkit.Device(db)
			dropTable(t, db, drop[name])

			assertKind(t, devicestore.New(db).Apply(ctx, d.ID, change), apperr.KindInternal)
		})
	}
}

func TestProtocolsOfTheProjectsTestsOnTheDevice(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	mine, other := testkit.Project(db), testkit.Project(db)
	dev, elsewhere := testkit.Device(db), testkit.Device(db)
	run := func(p *models.Project) uint {
		return testkit.TestRun(db, func(r *models.TestRun) {
			r.TestID = testkit.Test(db, func(x *models.Test) { x.AppID = testkit.App(db, func(a *models.App) { a.ProjectID = p.ID }).ID }).ID
		}).ID
	}
	protocol := func(runID, deviceID uint, parent *uint) uint {
		return testkit.Protocol(db, func(p *models.TestProtocol) {
			p.TestRunID, p.DeviceID, p.ParentTestProtocolID = runID, &deviceID, parent
		}).ID
	}
	first, second, foreign := run(mine), run(mine), run(other)
	runner := protocol(first, dev.ID, nil)
	wrapped := protocol(first, dev.ID, &runner)
	solo := protocol(second, dev.ID, nil)
	protocol(second, elsewhere.ID, nil)
	protocol(foreign, dev.ID, nil)
	store := devicestore.New(db)

	got, err := store.Protocols(ctx, svc.ProtocolScope{ProjectID: mine.ID, DeviceID: dev.ID, Limit: 10})

	if err != nil || len(got) != 2 || got[0].ID != solo || got[1].ID != wrapped {
		t.Fatalf("got %+v err %v", got, err)
	}
	if got[0].TestRun == nil || got[0].TestRun.Test == nil {
		t.Fatalf("test run and test not loaded: %+v", got[0].TestRun)
	}
	limited, _ := store.Protocols(ctx, svc.ProtocolScope{ProjectID: mine.ID, DeviceID: dev.ID, Limit: 1})
	if len(limited) != 1 || limited[0].ID != solo {
		t.Fatalf("limited %+v", limited)
	}

	dropTable(t, db, &models.TestProtocol{})
	_, err = store.Protocols(ctx, svc.ProtocolScope{ProjectID: mine.ID, DeviceID: dev.ID, Limit: 10})
	assertKind(t, err, apperr.KindInternal)
}

func TestBundleIsScopedToTheProject(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	app := testkit.App(db)
	binary := models.AppBinary{AppID: app.ID, Name: "b"}
	db.Create(&binary)
	store := devicestore.New(db)

	got, err := store.Bundle(ctx, &models.Project{Model: models.Model{ID: app.ProjectID}}, binary.ID)
	if err != nil || got.ID != binary.ID || got.App == nil {
		t.Fatalf("got %+v err %v", got, err)
	}
	_, err = store.Bundle(ctx, &models.Project{Model: models.Model{ID: app.ProjectID + 1}}, binary.ID)
	assertKind(t, err, apperr.KindNotFound)
	if err.Error() != "app bundle "+strconv.Itoa(int(binary.ID))+" not found" {
		t.Fatalf("message %q", err)
	}
}

func TestApplyWithoutParameterChangesOnlySetsTheAlias(t *testing.T) {
	ctx := t.Context()
	db := testkit.DB(t)
	d := testkit.Device(db)

	if err := devicestore.New(db).Apply(ctx, d.ID, svc.Change{Alias: "Lab"}); err != nil {
		t.Fatal(err)
	}
	var saved models.Device
	db.First(&saved, d.ID)
	if saved.Alias != "Lab" {
		t.Fatalf("alias %q", saved.Alias)
	}
}

func dropTable(t *testing.T, db *gorm.DB, model any) {
	t.Helper()
	if err := db.Migrator().DropTable(model); err != nil {
		t.Fatal(err)
	}
}
