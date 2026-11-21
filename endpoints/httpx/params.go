package httpx

import (
	"strconv"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/gin-gonic/gin"
)

// ParseID reads the numeric path parameter name. Anything else, including a
// missing value, is an Invalid error.
func ParseID(c *gin.Context, name string) (uint, error) {
	value := c.Param(name)
	id, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		return 0, apperr.Invalid("invalid %s %q", name, value)
	}
	return uint(id), nil
}

// Limit reads the limit query parameter: def when it is missing, not a
// number or not positive, and at most max.
func Limit(c *gin.Context, def, limitMax int) int {
	v, err := strconv.Atoi(c.Query("limit"))
	if err != nil || v <= 0 {
		return def
	}
	return min(v, limitMax)
}
