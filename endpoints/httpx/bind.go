package httpx

import (
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/gin-gonic/gin"
)

// BindJSON reads the JSON request body into v. A body that is malformed or
// doesn't fit v is an Invalid error carrying the binder's message.
func BindJSON(c *gin.Context, v any) error {
	if err := c.ShouldBindJSON(v); err != nil {
		return apperr.Invalid("%v", err)
	}
	return nil
}
