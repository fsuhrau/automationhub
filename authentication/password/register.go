package password

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// registration is the request to create a password account.
type registration struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

func (r *registration) trim() {
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
	r.Name = strings.TrimSpace(r.Name)
}

// problem says what is wrong with the registration, "" when it is fine.
func (r registration) problem() string {
	switch {
	case r.Email == "" || r.Password == "" || r.Name == "":
		return "name, email and password are required"
	case !strings.Contains(r.Email, "@") || strings.ContainsAny(r.Email, " \t"):
		return "that doesn't look like an email address"
	case len([]rune(r.Password)) < MinPasswordLength:
		return fmt.Sprintf("the password needs at least %d characters", MinPasswordLength)
	}
	return ""
}

// conflict says why the account can't be created, "" when it can.
func (p *Provider) conflict(r registration) string {
	var existing models.UserAuth
	if err := p.db.Where("email = ? and provider = ?", r.Email, ProviderName).First(&existing).Error; err == nil {
		return "an account with this email exists already, please sign in"
	}
	// names are unique (sessions and tokens find the account by it); a taken
	// name used to fail as "Failed to create user"
	var named int64
	if err := p.db.Model(&models.User{}).Where("name = ?", r.Name).Count(&named).Error; err == nil && named > 0 {
		return "someone has this name already; add your last name or an initial"
	}
	return ""
}

// create stores the account with the hash of its password.
func (p *Provider) create(r registration, hash []byte) (models.User, error) {
	user := models.User{
		Name: r.Name,
		Auth: []models.UserAuth{{Email: r.Email, Provider: ProviderName, PasswordHash: string(hash)}},
	}
	if err := p.db.Create(&user).Error; err != nil {
		return user, fmt.Errorf("create user %s: %w", r.Email, err)
	}
	return user, nil
}

// register creates an account and signs it in.
func (p *Provider) register(ctx *gin.Context) {
	var req registration
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.trim()
	if msg := req.problem(); msg != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	if msg := p.conflict(req); msg != "" {
		ctx.JSON(http.StatusConflict, gin.H{"error": msg})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user, err := p.create(req, hash)
	if err != nil {
		logrus.Errorf("register: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	authUser := authentication.User{Login: user.Name, Name: user.Name, Email: req.Email}
	p.guard.SignIn(ctx, authUser)
	ctx.JSON(http.StatusOK, gin.H{"user": authUser})
}
