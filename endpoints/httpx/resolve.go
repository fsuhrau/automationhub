package httpx

import (
	"context"
	"net/http"
	"strconv"

	"github.com/fsuhrau/automationhub/apperr"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// notFoundMessage is the text clients get for an unknown project or app.
// It is part of the REST contract and was once the text of the database
// error.
const notFoundMessage = "record not found"

// ProjectLookup finds the project with the identifier named in the URL.
type ProjectLookup interface {
	Get(ctx context.Context, identifier string) (*models.Project, error)
}

// AppLookup finds an app of a project, with its parameters.
type AppLookup interface {
	App(ctx context.Context, sc appsvc.Scope) (*models.App, error)
}

// ResolveProject is the middleware that loads the project of the
// :project_id parameter and stores it for Project.
func ResolveProject(projects ProjectLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		project, err := projects.Get(lookupContext(c), c.Param("project_id"))
		if err != nil {
			failResolve(c, err)
			return
		}
		c.Set(ProjectKey, project)
	}
}

// ResolveApp is the middleware that loads the app of the :app_id parameter
// within the resolved project and stores it for App. It must run after
// ResolveProject.
func ResolveApp(apps AppLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("app_id"), 10, 0)
		if err != nil {
			failResolve(c, apperr.NotFound("app %s not found", c.Param("app_id")))
			return
		}
		app, err := apps.App(lookupContext(c), appsvc.Scope{ProjectID: Project(c).ID, AppID: uint(id)})
		if err != nil {
			failResolve(c, err)
			return
		}
		c.Set(AppKey, app)
	}
}

// failResolve answers an unknown project or app with the stable 404 text and
// any other failure like every other handler.
func failResolve(c *gin.Context, err error) {
	if apperr.KindOf(err) == apperr.KindNotFound {
		c.AbortWithStatusJSON(http.StatusNotFound, ErrorResponse{Message: notFoundMessage})
		return
	}
	Fail(c, err)
}

// lookupContext is the context of the lookup of a project or app: the values
// of the request, but not its cancellation, so a client that is already gone
// still gets the status of the route (the streaming routes end themselves).
func lookupContext(c *gin.Context) context.Context {
	return context.WithoutCancel(c.Request.Context())
}
