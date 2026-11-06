package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

var errNoUser = errors.New("no signed in user; personal access tokens need a login (configure authentication on the master)")

// ProfileResponse is the signed in user.
type ProfileResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
	Email string `json:"email"`
}

// PersonalToken is a personal access token without its value.
type PersonalToken struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Hint       string     `json:"hint"`
	CreatedAt  time.Time  `json:"createdAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
}

// CreatedPersonalToken carries the token value, which is only returned once.
type CreatedPersonalToken struct {
	PersonalToken
	Token string `json:"token"`
}

func personalToken(t models.AccessToken) PersonalToken {
	return PersonalToken{ID: t.ID, Name: t.Name, Hint: t.Hint, CreatedAt: t.CreatedAt, ExpiresAt: t.ExpiresAt, LastUsedAt: t.LastUsedAt}
}

func (s *Service) registerProfile(api *gin.RouterGroup) {
	me := api.Group("/me")
	me.GET("", s.getProfile)
	me.GET("/tokens", s.getPersonalTokens)
	me.POST("/tokens", s.createPersonalToken)
	me.DELETE("/tokens/:token_id", s.deletePersonalToken)
}

// currentUser resolves the stored user of the request. Sessions only carry
// name, login and email: the name is unique, the login is the provider user
// id for OAuth2 logins.
func (s *Service) currentUser(c *gin.Context) (*models.User, authentication.User, error) {
	var authUser authentication.User
	value, _ := c.Get(authentication.UserKey)
	switch u := value.(type) {
	case authentication.User:
		authUser = u
	case *authentication.User:
		if u != nil {
			authUser = *u
		}
	}

	var user models.User
	if authUser.Name != "" {
		if err := s.db.First(&user, "name = ?", authUser.Name).Error; err == nil {
			return &user, authUser, nil
		}
	}
	for _, q := range []struct{ query, value string }{
		{"provider_user_id = ?", authUser.Login},
		{"email = ?", authUser.Email},
	} {
		if q.value == "" {
			continue
		}
		var auth models.UserAuth
		if err := s.db.Preload("User").First(&auth, q.query, q.value).Error; err == nil && auth.User != nil {
			return auth.User, authUser, nil
		}
	}
	return nil, authUser, errNoUser
}

func (s *Service) getProfile(c *gin.Context) {
	user, authUser, err := s.currentUser(c)
	if err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, ProfileResponse{ID: user.ID, Name: user.Name, Login: authUser.Login, Email: authUser.Email})
}

func (s *Service) getPersonalTokens(c *gin.Context) {
	user, _, err := s.currentUser(c)
	if err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}
	var tokens []models.AccessToken
	if err := s.db.Order("id desc").Find(&tokens, "user_id = ?", user.ID).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	result := make([]PersonalToken, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, personalToken(t))
	}
	c.JSON(http.StatusOK, result)
}

// sessionOnly keeps access tokens from managing tokens, so a leaked token
// can't create new ones or revoke the owner's.
func (s *Service) sessionOnly(c *gin.Context) bool {
	if _, viaToken := c.Get(token.AuthTokenKey); viaToken {
		s.error(c, http.StatusForbidden, fmt.Errorf("personal tokens can only be managed from a browser session"))
		return false
	}
	return true
}

func (s *Service) createPersonalToken(c *gin.Context) {
	if !s.sessionOnly(c) {
		return
	}
	user, _, err := s.currentUser(c)
	if err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}
	var req struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		s.error(c, http.StatusBadRequest, fmt.Errorf("name is required"))
		return
	}
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		s.error(c, http.StatusBadRequest, fmt.Errorf("expiry is in the past"))
		return
	}

	plain, stored, hint, err := token.NewPersonal()
	if err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	t := models.AccessToken{UserID: &user.ID, Name: req.Name, Token: stored, Hint: hint, ExpiresAt: req.ExpiresAt}
	if err := s.db.Create(&t).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, CreatedPersonalToken{PersonalToken: personalToken(t), Token: plain})
}

func (s *Service) deletePersonalToken(c *gin.Context) {
	if !s.sessionOnly(c) {
		return
	}
	user, _, err := s.currentUser(c)
	if err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}
	result := s.db.Where("id = ? AND user_id = ?", c.Param("token_id"), user.ID).Delete(&models.AccessToken{})
	if result.Error != nil {
		s.error(c, http.StatusInternalServerError, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		s.error(c, http.StatusNotFound, fmt.Errorf("token not found"))
		return
	}
	c.Status(http.StatusNoContent)
}
