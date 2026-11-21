package httpx

import (
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/gin"
)

// CurrentUser is the signed in user the authentication middleware put into
// the request. The middleware stores either the value or a pointer to it;
// ok is false when there is no user, and then the user is the zero value.
func CurrentUser(c *gin.Context) (authentication.User, bool) {
	value, _ := c.Get(authentication.UserKey)
	switch u := value.(type) {
	case authentication.User:
		return u, true
	case *authentication.User:
		if u != nil {
			return *u, true
		}
	}
	return authentication.User{}, false
}
