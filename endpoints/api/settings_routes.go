package api

import (
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/gin-gonic/gin"
)

// requireSession keeps access tokens from managing tokens, so a leaked
// token can't create new ones or revoke the owner's.
func requireSession(c *gin.Context) {
	if _, viaToken := c.Get(token.AuthTokenKey); viaToken {
		httpx.Fail(c, apperr.Forbidden("tokens can only be managed from a browser session"))
	}
}
