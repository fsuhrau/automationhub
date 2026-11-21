package token

import (
	"errors"
	"net/http"
	"time"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const projectDependent = "X-Auth-Token is project dependent"

// denial is the answer to a request whose token is not accepted.
type denial struct {
	status  int
	message string
}

func unauthorized() *denial { return &denial{http.StatusUnauthorized, "unauthorized"} }

// guard checks the access token of a request.
type guard struct {
	db *gorm.DB
}

// check returns nil and signs the request in when the token is accepted. It
// fails closed: only a token that was found signs the request in.
func (g guard) check(ctx *gin.Context, value string) *denial {
	var token models.AccessToken
	if d := g.find(value, &token); d != nil {
		return d
	}
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		return &denial{http.StatusUnauthorized, "unauthorized token expired"}
	}
	// a personal token acts as its user, in every project
	if token.UserID != nil {
		return personalDenial(authenticatePersonal(g.db, ctx, &token))
	}
	if d := g.checkProject(ctx, &token); d != nil {
		return d
	}
	if d := g.checkNode(ctx, &token); d != nil {
		return d
	}
	ctx.Set(authentication.UserKey, token.Name)
	ctx.Set(AuthTokenKey, token.ID)
	return nil
}

func personalDenial(ok bool) *denial {
	if ok {
		return nil
	}
	return unauthorized()
}

// find loads the token of the header value.
func (g guard) find(value string, token *models.AccessToken) *denial {
	err := lookup(g.db, value, token)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return unauthorized()
	}
	if err != nil {
		return &denial{http.StatusInternalServerError, "token lookup failed"}
	}
	return nil
}

// checkProject makes a project token work only for its own project.
func (g guard) checkProject(ctx *gin.Context, token *models.AccessToken) *denial {
	if token.ProjectID == nil {
		return nil
	}
	// tokens work only with params
	projectParam := ctx.Param("project_id")
	if projectParam == "" {
		return &denial{http.StatusBadRequest, projectDependent}
	}
	var project models.Project
	if err := g.db.First(&project, "id = ?", token.ProjectID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return unauthorized()
	}
	if projectParam != project.Identifier {
		return &denial{http.StatusBadRequest, projectDependent}
	}
	return nil
}

// checkNode makes a node token work only for the node endpoints. It is the
// node's credential, not an API key. FullPath is the matched route pattern,
// so ".." or doubled slashes in the request path can't sneak past the check.
func (g guard) checkNode(ctx *gin.Context, token *models.AccessToken) *denial {
	if token.NodeID == nil {
		return nil
	}
	if !NodeRoute(ctx.FullPath()) {
		return &denial{http.StatusForbidden, "node tokens only work for node endpoints"}
	}
	var node models.Node
	if err := g.db.First(&node, "id = ?", token.NodeID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return unauthorized()
	}
	return nil
}
