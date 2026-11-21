package oauthflow

import (
	"net/http"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// codeKey is the query parameter that carries the authorization code.
const codeKey = "code"

// login starts a sign in without a state and completes it with one.
func (f *Flow) login(ctx *gin.Context) {
	if _, ok := f.guard.User(ctx); ok {
		ctx.Next()
		return
	}
	state := ctx.Query(authentication.StateKey)
	if state == "" {
		f.redirectToProvider(ctx)
		return
	}
	if !consumeState(sessions.Default(ctx), state) {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}
	f.callback(ctx)
}

// redirectToProvider remembers a fresh state for this browser and sends it to the provider.
func (f *Flow) redirectToProvider(ctx *gin.Context) {
	state := newState()
	if err := storeState(sessions.Default(ctx), state); err != nil {
		logrus.Errorf("[%s] start login: %v", f.strategy.Name(), err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to start login"})
		return
	}
	ctx.Redirect(http.StatusTemporaryRedirect, f.conf.AuthCodeURL(state))
}

// callback trades the code for a token, loads the user and signs it in.
func (f *Flow) callback(ctx *gin.Context) {
	reqCtx := ctx.Request.Context()
	tok, err := f.conf.Exchange(reqCtx, ctx.Query(codeKey))
	if err != nil {
		logrus.Errorf("[%s] exchange: %v", f.strategy.Name(), err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to do exchange"})
		return
	}
	profile, err := f.strategy.Profile(reqCtx, f.conf.Client(reqCtx, tok))
	if err != nil {
		logrus.Errorf("[%s] get user: %v", f.strategy.Name(), err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to get user"})
		return
	}
	user, err := f.users.ensure(profile)
	if err != nil {
		logrus.Errorf("[%s] %v", f.strategy.Name(), err)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	f.guard.SignIn(ctx, authentication.User{ID: user.ID, Login: profile.Login, Name: user.Name, URL: profile.URL})
	if to := f.strategy.SuccessRedirect(); to != "" {
		ctx.Redirect(http.StatusTemporaryRedirect, to)
	}
}
