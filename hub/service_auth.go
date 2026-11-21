package hub

import "github.com/gin-gonic/gin"

// Auth is the authentication of the hub, built by the composition root from
// the configured provider.
type Auth struct {
	// Session is the middleware that loads the session of every request.
	Session gin.HandlerFunc
	// Guard protects the authenticated routes.
	Guard gin.HandlerFunc
	// Routes adds the login routes to the router.
	Routes func(*gin.Engine)
}

// mount adds the session middleware and the login routes to the router and
// returns the group of the authenticated routes. Without a session
// middleware the hub is open: the session check always answers ok.
func (a Auth) mount(router *gin.Engine) *gin.RouterGroup {
	if a.Session == nil {
		router.GET("/auth/session", openSession)
	} else {
		router.Use(a.Session)
	}
	if a.Routes != nil {
		a.Routes(router)
	}
	group := router.Group("/")
	if a.Guard != nil {
		group.Use(a.Guard)
	}
	return group
}
