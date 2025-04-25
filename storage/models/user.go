package models

type User struct {
	Model
	Name      string     `json:"name" gorm:"uniqueIndex;not null"`
	Role      string     `json:"role"`
	Auth      []UserAuth `json:"auth"`
	Companies []*Company `json:"companies" gorm:"many2many:user_companies;"`
	Projects  []*Project `json:"projects" gorm:"many2many:user_projects;"`
}
