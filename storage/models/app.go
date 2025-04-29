package models

type App struct {
	Model
	ProjectID  uint           `json:"projectId"`
	Project    *Project       `json:"project" gorm:"foreignKey:ProjectID"`
	Name       string         `json:"name"`
	Identifier string         `json:"identifier"`
	Platform   PlatformType   `json:"platform"`
	Parameter  []AppParameter `json:"parameter"`
	//	DefaultParameter string       `json:"defaultParameter"`
}
