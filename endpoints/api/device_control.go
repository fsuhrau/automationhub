package api

import (
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/gin"
)

// requestOwner names the authenticated caller for device leases.
func requestOwner(c *gin.Context) string {
	value, _ := c.Get(authentication.UserKey)
	switch u := value.(type) {
	case string:
		if u != "" {
			return "token:" + u
		}
	case authentication.User:
		if u.Name != "" {
			return u.Name
		}
		if u.Login != "" {
			return u.Login
		}
	case *authentication.User:
		if u != nil && u.Name != "" {
			return u.Name
		}
		if u != nil && u.Login != "" {
			return u.Login
		}
	}
	return "anonymous"
}
