package authentication

import "github.com/gin-gonic/gin"

// Provider is one way to sign in: it owns the session store and brings the
// session middleware, the guard for the authenticated routes and its routes.
type Provider interface {
	// Session returns the middleware that loads the cookie session of the given name.
	Session(name string) gin.HandlerFunc
	// SessionHandler returns the guard that lets signed-in users pass.
	SessionHandler() gin.HandlerFunc
	// Routes registers the sign in, sign out and session routes.
	Routes(r *gin.Engine)
}
