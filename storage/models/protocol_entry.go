package models

import (
	"time"
)

type ProtocolEntry struct {
	Model
	TestProtocolID uint      `json:"testProtocolId"`
	Timestamp      time.Time `json:"timestamp"`
	Source         string    `json:"source"`
	Level          string    `json:"level"`
	Message        string    `json:"message"`
	Data           string    `json:"data"`
	Runtime        float64   `json:"runtime"`
}
