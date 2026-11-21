package testkit

import (
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

func rows(t *testing.T, db *gorm.DB, m any) int64 {
	t.Helper()
	var n int64
	if err := db.Model(m).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestBuildersCreateMissingParents(t *testing.T) {
	tests := []struct {
		name  string
		build func(db *gorm.DB) uint
		want  []any // one row each
	}{
		{"project", func(db *gorm.DB) uint { return Project(db).ID }, []any{&models.Project{}}},
		{"app", func(db *gorm.DB) uint { return App(db).ID }, []any{&models.App{}, &models.Project{}}},
		{"test", func(db *gorm.DB) uint { return Test(db).ID }, []any{&models.Test{}, &models.App{}, &models.Project{}}},
		{"run", func(db *gorm.DB) uint { return TestRun(db).ID }, []any{&models.TestRun{}, &models.Test{}, &models.App{}, &models.Project{}}},
		{"protocol", func(db *gorm.DB) uint { return Protocol(db).ID }, []any{&models.TestProtocol{}, &models.TestRun{}, &models.Test{}}},
		{"device", func(db *gorm.DB) uint { return Device(db).ID }, []any{&models.Device{}}},
		{"user", func(db *gorm.DB) uint { return User(db).ID }, []any{&models.User{}}},
		{"token", func(db *gorm.DB) uint { return Token(db).ID }, []any{&models.AccessToken{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := DB(t)
			if tt.build(db) == 0 {
				t.Fatal("no id")
			}
			for _, m := range tt.want {
				if n := rows(t, db, m); n != 1 {
					t.Errorf("%T rows = %d, want 1", m, n)
				}
			}
		})
	}
}

func TestBuildersUseGivenParents(t *testing.T) {
	db := DB(t)
	app := App(db)
	test := Test(db, func(m *models.Test) { m.AppID = app.ID })
	run := TestRun(db, func(m *models.TestRun) { m.TestID = test.ID })
	Protocol(db, func(m *models.TestProtocol) { m.TestRunID = run.ID })
	for _, m := range []any{&models.Project{}, &models.App{}, &models.Test{}, &models.TestRun{}, &models.TestProtocol{}} {
		if n := rows(t, db, m); n != 1 {
			t.Errorf("%T rows = %d, want 1", m, n)
		}
	}
}

func TestUniqueDefaults(t *testing.T) {
	db := DB(t)
	u1, u2 := User(db), User(db)
	if u1.Name == u2.Name {
		t.Error("user names collide")
	}
	t1, t2 := Token(db), Token(db)
	if t1.Token == t2.Token {
		t.Error("tokens collide")
	}
}

func TestWithName(t *testing.T) {
	db := DB(t)
	if got := Project(db, WithName[models.Project]("x")).Name; got != "x" {
		t.Errorf("project name = %q", got)
	}
	if got := Test(db, WithName[models.Test]("t")).Name; got != "t" {
		t.Errorf("test name = %q", got)
	}
}

func TestWithNameWithoutNamePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	WithName[models.TestRun]("x")(&models.TestRun{})
}

func TestSeedPanicsOnStoreError(t *testing.T) {
	db := DB(t)
	User(db, WithName[models.User]("same"))
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for duplicate user name")
		}
	}()
	User(db, WithName[models.User]("same"))
}
