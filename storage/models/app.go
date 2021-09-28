package models

import (
	"github.com/fsuhrau/automationhub/events"
	"gorm.io/gorm"
)

type App struct {
	gorm.Model
	AppID          string
	Name           string
	Platform       string
	Version        string
	AppPath        string
	Identifier     string
	LaunchActivity string
	Additional     string
	Hash           string
	Size           int
	AppFunctions   []AppFunction
}

func (a *App) AfterCreate(tx *gorm.DB) (err error) {
	events.AppCreated.Trigger(events.AppCreatedPayload{
		AppID: a.ID,
		App:   a,
	})
	return nil
}
