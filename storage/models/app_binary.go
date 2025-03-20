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
	gorm.Model
	AppID      uint
	App        *App
	Name       string
	Platform   string
	Version    string
	AppPath    string
	Identifier string
	Additional string
	Hash       string
	Size       int
	Tags       string
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
