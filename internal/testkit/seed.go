package testkit

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// Option changes a model before a builder stores it.
type Option[T any] func(*T)

// WithName sets the Name field of any model that has one. It panics for a
// model without a string Name field, which is a mistake in the test.
func WithName[T any](name string) Option[T] {
	return func(m *T) {
		f := reflect.ValueOf(m).Elem().FieldByName("Name")
		if !f.IsValid() || f.Kind() != reflect.String {
			panic(fmt.Sprintf("testkit: %T has no string Name field", *m))
		}
		f.SetString(name)
	}
}

func unique(prefix string) string { return prefix + "-" + rand.Text() }

// seed applies the options in order and stores m. The builders append an
// option that fills in missing parents, so it runs after the caller's.
// A failing insert panics: a seed that fails is a broken test setup.
func seed[T any](db *gorm.DB, m *T, opts []Option[T]) *T {
	for _, opt := range opts {
		opt(m)
	}
	if err := db.Create(m).Error; err != nil {
		panic(fmt.Sprintf("testkit: seed %T: %v", *m, err))
	}
	return m
}

// Project stores a project.
func Project(db *gorm.DB, opts ...Option[models.Project]) *models.Project {
	m := &models.Project{Name: "project", Identifier: unique("project")}
	return seed(db, m, opts)
}

// App stores an app; without a ProjectID it creates a project for it.
func App(db *gorm.DB, opts ...Option[models.App]) *models.App {
	m := &models.App{Name: "app", Identifier: unique("app"), Platform: models.PlatformTypeAndroid}
	return seed(db, m, append(opts, func(a *models.App) {
		if a.ProjectID == 0 {
			a.ProjectID = Project(db).ID
		}
	}))
}

// Test stores a test; without an AppID it creates an app (and project).
func Test(db *gorm.DB, opts ...Option[models.Test]) *models.Test {
	m := &models.Test{Name: "test"}
	return seed(db, m, append(opts, func(x *models.Test) {
		if x.AppID == 0 {
			x.AppID = App(db).ID
		}
	}))
}

// TestRun stores a test run; without a TestID it creates a test.
func TestRun(db *gorm.DB, opts ...Option[models.TestRun]) *models.TestRun {
	m := &models.TestRun{SessionID: unique("session")}
	return seed(db, m, append(opts, func(r *models.TestRun) {
		if r.TestID == 0 {
			r.TestID = Test(db).ID
		}
	}))
}

// Protocol stores a test protocol; without a TestRunID it creates a test run.
func Protocol(db *gorm.DB, opts ...Option[models.TestProtocol]) *models.TestProtocol {
	m := &models.TestProtocol{TestName: "protocol", StartedAt: time.Now()}
	return seed(db, m, append(opts, func(p *models.TestProtocol) {
		if p.TestRunID == 0 {
			p.TestRunID = TestRun(db).ID
		}
	}))
}

// Device stores a device.
func Device(db *gorm.DB, opts ...Option[models.Device]) *models.Device {
	m := &models.Device{Name: "device", DeviceIdentifier: unique("device")}
	return seed(db, m, opts)
}

// User stores a user with a unique name.
func User(db *gorm.DB, opts ...Option[models.User]) *models.User {
	m := &models.User{Name: unique("user"), Role: "user"}
	return seed(db, m, opts)
}

// Token stores an access token with a unique value.
func Token(db *gorm.DB, opts ...Option[models.AccessToken]) *models.AccessToken {
	m := &models.AccessToken{Name: "token", Token: unique("token")}
	return seed(db, m, opts)
}
