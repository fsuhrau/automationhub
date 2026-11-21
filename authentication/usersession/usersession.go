// Package usersession keeps the signed-in user in a cookie session. All sign
// in methods share it; they differ only in the method name and login URL.
package usersession

import (
	"encoding/gob"
	"net/http"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Options describe how a sign in method presents itself to the front end.
type Options struct {
	// Method is reported by the session endpoint ("github", "oauth2", "password").
	Method string
	// LoginURL is where the front end sends users that are not signed in.
	LoginURL string
	// StatusIncludesURL adds LoginURL to the "not signed in" session answer.
	StatusIncludesURL bool
}

// Guard reads and writes the user in the session of a request.
type Guard struct {
	store sessions.CookieStore
	opts  Options
}

// New returns a guard whose cookies are signed with secret.
func New(secret []byte, opts Options) *Guard {
	gob.Register(authentication.User{})
	return &Guard{store: sessions.NewCookieStore(secret), opts: opts}
}

// Middleware loads the cookie session called name into every request.
func (g *Guard) Middleware(name string) gin.HandlerFunc {
	return sessions.Sessions(name, g.store)
}

// User returns the signed-in user of the request. A session without a hub
// user id (a cookie from before sessions carried it) names nobody, so its
// holder is signed out and has to sign in again.
func (g *Guard) User(ctx *gin.Context) (authentication.User, bool) {
	user, ok := sessions.Default(ctx).Get(authentication.SessionKey).(authentication.User)
	return user, ok && user.ID != 0
}

// SignIn makes user the signed-in user of the request and of the session.
func (g *Guard) SignIn(ctx *gin.Context, user authentication.User) {
	ctx.Set(authentication.UserKey, user)
	session := sessions.Default(ctx)
	session.Set(authentication.SessionKey, user)
	if err := session.Save(); err != nil {
		logrus.Errorf("[%s] Failed to save session: %v", g.opts.Method, err)
	}
}

// RequireUser lets signed-in users pass and answers everybody else with 401.
func (g *Guard) RequireUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if user, ok := g.User(ctx); ok {
			ctx.Set(authentication.UserKey, user)
			ctx.Next()
			return
		}
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "url": g.opts.LoginURL})
	}
}

// Status answers with the signed-in user, or 401 and the sign in method.
func (g *Guard) Status(ctx *gin.Context) {
	if user, ok := g.User(ctx); ok {
		ctx.Set(authentication.UserKey, user)
		ctx.JSON(http.StatusOK, gin.H{"user": user})
		return
	}
	body := gin.H{"error": "not signed in", "method": g.opts.Method}
	if g.opts.StatusIncludesURL {
		body["url"] = g.opts.LoginURL
	}
	ctx.JSON(http.StatusUnauthorized, body)
}

// Logout removes the user from the session.
func (g *Guard) Logout(ctx *gin.Context) {
	if _, ok := g.User(ctx); ok {
		session := sessions.Default(ctx)
		session.Delete(authentication.SessionKey)
		if err := session.Save(); err != nil {
			logrus.Errorf("[%s] Failed to save session: %v", g.opts.Method, err)
		}
	}
	ctx.JSON(http.StatusOK, nil)
}

// Register adds the session and logout routes every method has.
func (g *Guard) Register(r gin.IRoutes) {
	r.GET("/auth/session", g.Status)
	r.POST("/auth/logout", g.Logout)
}
