package apps

import (
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// invalid makes a request that could not be read an Invalid error.
func invalid(err error) error { return apperr.Invalid("%s", err) }

func (h *Handler) addParameter(c *gin.Context) {
	in, err := parameterInput(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	added, err := h.svc.AddParameter(c.Request.Context(), scope(c), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, added)
}

func (h *Handler) updateParameter(c *gin.Context) {
	ref, err := parameterRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	in, err := parameterInput(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	updated, err := h.svc.UpdateParameter(c.Request.Context(), ref, in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) removeParameter(c *gin.Context) {
	ref, err := parameterRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	if err := h.svc.RemoveParameter(c.Request.Context(), ref); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func parameterRef(c *gin.Context) (appsvc.ParameterRef, error) {
	id, err := httpx.ParseID(c, "parameter_id")
	return appsvc.ParameterRef{Scope: scope(c), ID: id}, err
}

// parameterInput reads the body as the parameter model does, which takes the
// type as a string or as an object.
func parameterInput(c *gin.Context) (appsvc.ParameterInput, error) {
	var body models.AppParameter
	if err := c.ShouldBindJSON(&body); err != nil {
		return appsvc.ParameterInput{}, invalid(err)
	}
	return appsvc.ParameterInput{Name: body.Name, Type: body.Type}, nil
}
