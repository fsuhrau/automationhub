package api

import (
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication/token"
	settingsapi "github.com/fsuhrau/automationhub/endpoints/api/settings"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/access"
	"github.com/fsuhrau/automationhub/storage/accessstore"
	"github.com/gin-gonic/gin"
)

// requireSession keeps access tokens from managing tokens, so a leaked
// token can't create new ones or revoke the owner's.
func requireSession(c *gin.Context) {
	if _, viaToken := c.Get(token.AuthTokenKey); viaToken {
		httpx.Fail(c, apperr.Forbidden("tokens can only be managed from a browser session"))
	}
}

// accessService builds the use cases of users, tokens and nodes.
func (s *Service) accessService() *access.Service {
	return access.New(access.Deps{Repo: accessstore.New(s.db), Status: s.nodeManager})
}

// registerSettingsRoutes adds the project settings. They hand out and revoke
// tokens, so only a browser session may use them: a project token must not
// mint a node token, or tokens for other projects.
func (s *Service) registerSettingsRoutes(api *gin.RouterGroup) {
	settings := api.Group("/:project_id/settings", requireSession, s.ResolveProject)
	settingsapi.New(s.accessService()).Register(settings)
}
