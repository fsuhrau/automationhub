package models

// UserAuth is one way a user signs in. An email only identifies an account when there is one: providers such as the
// generic OAuth2 one or GitHub with a private email give none, so the unique
// indexes skip the empty values. The index names are global in PostgreSQL.
type UserAuth struct {
	Model
	UserID         uint   `json:"user_id" gorm:"uniqueIndex:idx_userid_provider;not null"`
	User           *User  `json:"user"`
	Email          string `json:"email" gorm:"uniqueIndex:idx_user_auths_email,where:email <> '';not null"`
	Provider       string `json:"provider" gorm:"uniqueIndex:idx_user_auths_email,where:email <> '';uniqueIndex:idx_userid_provider;uniqueIndex:idx_user_auths_identity,where:provider_user_id <> '';not null"`
	ProviderUserID string `json:"providerUserId" gorm:"uniqueIndex:idx_user_auths_identity,where:provider_user_id <> ''"`
	Token          string `json:"token"`
	Username       string `json:"username"`
	PasswordHash   string `json:"password_hash"`
}
