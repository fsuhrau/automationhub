package password

import (
	"net/http"
	"strings"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// credentials are the email and password of a login.
type credentials struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (c *credentials) trim() {
	c.Email = strings.TrimSpace(c.Email)
	c.Password = strings.TrimSpace(c.Password)
}

// login signs in the account whose email and password match.
func (p *Provider) login(ctx *gin.Context) {
	var req credentials
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.trim()
	account, ok := p.authenticate(req)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	user := authentication.User{
		Login: account.User.Name,
		Role:  account.User.Name,
		Name:  account.User.Name,
		Email: account.Email,
	}
	p.guard.SignIn(ctx, user)
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

// authenticate finds the password account of the email and checks the password.
func (p *Provider) authenticate(req credentials) (*models.UserAuth, bool) {
	var account models.UserAuth
	err := p.db.Where("email = ? and provider = ?", req.Email, ProviderName).Preload("User").First(&account).Error
	if err != nil {
		return nil, false
	}
	if bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)) != nil {
		return nil, false
	}
	return &account, true
}
