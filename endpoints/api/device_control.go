package api

import (
	"fmt"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/gin-gonic/gin"
)

const anonymousOwner = "anonymous"

// requestOwner names the authenticated caller for device leases. Leases are
// compared by this string, so it ends in the hub user id: users that share a
// display name are different owners. The name in front is for display only.
func requestOwner(c *gin.Context) string {
	if u, ok := httpx.CurrentUser(c); ok {
		return userOwner(u)
	}
	value, _ := c.Get(authentication.UserKey)
	if name, ok := value.(string); ok && name != "" {
		return "token:" + name
	}
	return anonymousOwner
}

// userOwner is the lease owner of a signed-in user, "Name (user:7)".
func userOwner(u authentication.User) string {
	switch {
	case u.ID == 0:
		return anonymousOwner
	case u.Name == "":
		return fmt.Sprintf("user:%d", u.ID)
	}
	return fmt.Sprintf("%s (user:%d)", u.Name, u.ID)
}
