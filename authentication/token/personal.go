package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PersonalPrefix starts every personal access token. Project and node tokens
// are UUIDs, so the prefix tells the two kinds apart.
const PersonalPrefix = "ahp_"

// storedPrefix marks the hash of a personal token in AccessToken.Token. The
// plain token is only shown once, when it is created.
const storedPrefix = "sha256:"

// AuthTokenKey is set in the gin context when a request was authenticated
// with an access token instead of a browser session.
const AuthTokenKey = "auth_token_id"

// lastUsedInterval limits the writes of AccessToken.LastUsedAt.
const lastUsedInterval = time.Minute

// NewPersonal returns a new personal token, the value to store in
// AccessToken.Token and a hint (the last characters) to show in lists.
func NewPersonal() (plain, stored, hint string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", err
	}
	plain = PersonalPrefix + base64.RawURLEncoding.EncodeToString(b)
	return plain, StoredPersonal(plain), plain[len(plain)-4:], nil
}

// StoredPersonal returns the stored form (hash) of a personal token.
func StoredPersonal(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return storedPrefix + hex.EncodeToString(sum[:])
}

// IsPersonal tells whether the header value is a personal token.
func IsPersonal(value string) bool {
	return strings.HasPrefix(value, PersonalPrefix)
}

// lookup finds the token of an X-Auth-Token header. Personal tokens are only
// found by their hash among user tokens, other tokens only by their plain
// value among non-user tokens, so a leaked hash can't be used as a token.
func lookup(db *gorm.DB, value string, token *models.AccessToken) error {
	if IsPersonal(value) {
		return db.First(token, "token = ? AND user_id IS NOT NULL", StoredPersonal(value)).Error
	}
	return db.First(token, "token = ? AND user_id IS NULL", value).Error
}

// authenticatePersonal signs the request in as the token's user.
func authenticatePersonal(db *gorm.DB, ctx *gin.Context, token *models.AccessToken) bool {
	var user models.User
	if err := db.Preload("Auth").First(&user, *token.UserID).Error; err != nil {
		return false
	}
	authUser := authentication.User{Login: user.Name, Name: user.Name, Role: user.Role}
	for _, a := range user.Auth {
		if a.Email != "" {
			authUser.Email = a.Email
			break
		}
	}
	ctx.Set(authentication.UserKey, authUser)
	ctx.Set(AuthTokenKey, token.ID)

	now := time.Now()
	if token.LastUsedAt == nil || now.Sub(*token.LastUsedAt) > lastUsedInterval {
		db.Model(token).UpdateColumn("last_used_at", now)
	}
	return true
}
