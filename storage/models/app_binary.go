package models

import (
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
}

func (a *AppBinary) AfterCreate(tx *gorm.DB) (err error) {
	events.AppCreated.Trigger(events.AppCreatedPayload{
		AppID: a.ID,
		App:   a,
	})
	return nil
}
