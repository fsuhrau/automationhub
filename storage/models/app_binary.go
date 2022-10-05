package models

import (
	"github.com/fsuhrau/automationhub/events"
	"gorm.io/gorm"
)

type AppBinary struct {
	gorm.Model
	AppID          uint
	App            *App
	Name           string
	Platform       string
	Version        string
	AppPath        string
	Identifier     string
	LaunchActivity string
	Additional     string
	Hash           string
	Size           int
	Tags           string
}

func (a *AppBinary) AfterCreate(tx *gorm.DB) (err error) {
	events.AppCreated.Trigger(events.AppCreatedPayload{
		AppID: a.ID,
		App:   a,
	})
	return nil
}
