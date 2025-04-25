package models

type Company struct {
	Model
	Token    string    `json:"token" gorm:"unique"`
	Name     string    `json:"name" gorm:"unique"`
	Users    []*User   `json:"users" gorm:"many2many:user_companies;"`
	Projects []Project `json:"projects" gorm:"many2many:project_companies;"`
}
