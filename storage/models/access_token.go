package models

import (
	"time"
)

type AccessToken struct {
	Model
	ProjectID *uint    `json:"projectId" gorm:"default:NULL"`
	Project   *Project `json:"project"`
	NodeID    *uint    `json:"nodeId" gorm:"default:NULL"`
	Node      *Node    `json:"node"`
	// UserID is set for personal tokens; Token then holds a hash
	UserID     *uint      `json:"userId" gorm:"default:NULL;index"`
	User       *User      `json:"-"`
	Hint       string     `json:"hint"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	Name       string     `json:"name"`
	Token      string     `json:"token" gorm:"unique"`
	ExpiresAt  *time.Time `json:"expiresAt"`
}
