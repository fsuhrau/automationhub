package models

import (
	"time"
)

type AccessToken struct {
	Model
	ProjectID *uint      `json:"projectId" gorm:"default:NULL"`
	Project   *Project   `json:"project"`
	NodeID    *uint      `json:"nodeId" gorm:"default:NULL"`
	Node      *Node      `json:"node"`
	Name      string     `json:"name"`
	Token     string     `json:"token" gorm:"unique"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
