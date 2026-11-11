package models

import (
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/events"
	"gorm.io/gorm"
)

type Android struct {
	LaunchActivity string
}
type Executable struct {
	Executable string
}

type AppBinary struct {
	Model
	AppID      uint       `json:"appId"`
	App        *App       `json:"app"`
	Name       string     `json:"name"`
	Platform   string     `json:"platform"`
	Version    string     `json:"version"`
	AppPath    string     `json:"appPath"`
	Identifier string     `json:"identifier"`
	Additional string     `json:"additional"`
	Hash       string     `json:"hash"`
	Size       int        `json:"size"`
	Tags       string     `json:"tags"`
	Android    Android    `json:"android" db:"android" gorm:"embedded"`
	Executable Executable `json:"executable" db:"executable" gorm:"embedded"`
	// Info is all the master read from the file; nil for bundles uploaded
	// before it did
	Info *app.BundleInfo `json:"info,omitempty" gorm:"serializer:json"`
	// ArchivedAt is when the file of the bundle was removed. An archived
	// bundle keeps all read from its file, so its runs stay comparable, but
	// it can't be installed or run anymore.
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}

// Archived says whether the file of the bundle is gone
func (a *AppBinary) Archived() bool {
	return a.ArchivedAt != nil
}

func (a *AppBinary) AfterCreate(tx *gorm.DB) (err error) {
	events.AppCreated.Trigger(events.AppCreatedPayload{
		AppID: a.ID,
		App:   a,
	})
	return nil
}
