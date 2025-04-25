package models

type UserAuth struct {
	Model
	UserID         uint   `json:"user_id" gorm:"uniqueIndex:idx_userid_provider;not null"`
	User           *User  `json:"user"`
	Email          string `json:"email" gorm:"uniqueIndex:idx_email_provider;not null"`
	Provider       string `json:"provider" gorm:"uniqueIndex:idx_email_provider;uniqueIndex:idx_userid_provider;not null"`
	ProviderUserID string `json:"providerUserId"`
	Token          string `json:"token"`
	Username       string `json:"username"`
	PasswordHash   string `json:"password_hash"`
}
