// Package profile is the HTTP adapter of the signed in user: the profile
// and the personal access tokens.
package profile

import (
	"context"
	"net/http"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/access"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers need from the access use cases.
type Service interface {
	Me(ctx context.Context, who authentication.User) (*models.User, error)
	PersonalTokens(ctx context.Context, who authentication.User) ([]models.AccessToken, error)
	CreatePersonalToken(ctx context.Context, who authentication.User, in access.NewToken) (*access.PersonalToken, error)
	DeletePersonalToken(ctx context.Context, who authentication.User, tokenID uint) error
}

// Handler serves the profile routes.
type Handler struct {
	svc Service
}

// New returns a Handler on the service.
func New(svc Service) *Handler { return &Handler{svc: svc} }

// Register adds the routes to the profile group. The guard runs before the
// routes that create or revoke tokens.
func (h *Handler) Register(me gin.IRoutes, guard gin.HandlerFunc) {
	me.GET("", h.profile)
	me.GET("/tokens", h.tokens)
	me.POST("/tokens", guard, h.createToken)
	me.DELETE("/tokens/:token_id", guard, h.deleteToken)
}

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

func (h *Handler) profile(c *gin.Context) {
	who, _ := httpx.CurrentUser(c)
	user, err := h.svc.Me(c.Request.Context(), who)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ProfileResponse{ID: user.ID, Name: user.Name, Login: who.Login, Email: who.Email})
}

func (h *Handler) tokens(c *gin.Context) {
	who, _ := httpx.CurrentUser(c)
	tokens, err := h.svc.PersonalTokens(c.Request.Context(), who)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	result := make([]PersonalToken, 0, len(tokens))
	for _, t := range tokens {
		result = append(result, personalToken(t))
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) createToken(c *gin.Context) {
	var request struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		httpx.Fail(c, apperr.Invalid("%v", err))
		return
	}
	who, _ := httpx.CurrentUser(c)
	created, err := h.svc.CreatePersonalToken(c.Request.Context(), who, access.NewToken(request))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, CreatedPersonalToken{PersonalToken: personalToken(created.AccessToken), Token: created.Plain})
}

func (h *Handler) deleteToken(c *gin.Context) {
	id, err := httpx.ParseID(c, "token_id")
	if err == nil {
		who, _ := httpx.CurrentUser(c)
		err = h.svc.DeletePersonalToken(c.Request.Context(), who, id)
	}
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
