package models

type Project struct {
	Model
	Identifier   string         `json:"identifier"`
	Name         string         `json:"name"`
	CompanyID    uint           `json:"companyId"`
	Company      *Company       `json:"-"`
	AccessTokens []*AccessToken `json:"accessTokens"`
	Users        []*User        `json:"users" gorm:"many2many:user_projects;"`
	Apps         []*App         `json:"apps"`
}
