package models

type Node struct {
	Model
	Identifier string `json:"identifier" gorm:"unique"`
	Name       string `json:"name"`
	Status     int    `json:"status" db:"-"`
}
