package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"
)

const (
	ProviderName = "github"
	CodeKey      = "code"
)

type Credentials struct {
	ClientID     string `json:"clientid"`
	ClientSecret string `json:"secret"`
}

var (
	conf  *oauth2.Config
	state string
	store sessions.CookieStore
	db    *gorm.DB
)

func randToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		logrus.Fatalf("[Github-OAuth2] Failed to read rand: %v\n", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func Setup(gormDb *gorm.DB, redirectURL, credFile string, scopes []string, secret []byte) {
	store = sessions.NewCookieStore(secret)
	db = gormDb
	var c Credentials
	file, err := ioutil.ReadFile(credFile)
	if err != nil {
		logrus.Fatalf("[Github-OAuth2] File error: %v\n", err)
	}
	err = json.Unmarshal(file, &c)
	if err != nil {
		logrus.Fatalf("[Github-OAuth2] Failed to unmarshal client credentials: %v\n", err)
	}
	conf = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     oauth2gh.Endpoint,
	}
}

func Session(name string) gin.HandlerFunc {
	return sessions.Sessions(name, store)
}

func loginRedirect(ctx *gin.Context) {
	state = randToken()
	session := sessions.Default(ctx)
	session.Set(authentication.StateKey, state)
	if err := session.Save(); err != nil {
		fmt.Println(err.Error())
	}
	ctx.Redirect(http.StatusTemporaryRedirect, GetLoginURL(state))
}

func GetLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

func init() {
	gob.Register(authentication.User{})
}

func Routes(r *gin.Engine) {
	r.GET("/auth/session", validateSessionRoute)
	r.POST("/auth/logout", logoutRoute)
	r.Any("/auth/oauth2", oauthRoute)
}

func SessionHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var (
			ok       bool
			authUser authentication.User
		)

		// Handle the exchange code to initiate a transport.
		session := sessions.Default(ctx)
		mysession := session.Get(authentication.SessionKey)
		if authUser, ok = mysession.(authentication.User); ok {
			ctx.Set(authentication.UserKey, authUser)
			ctx.Next()
			return
		}

		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "url": "/auth/oauth2"})
	}
}

func validateSessionRoute(ctx *gin.Context) {
	session := sessions.Default(ctx)

	mysession := session.Get(authentication.SessionKey)
	if authUser, ok := mysession.(authentication.User); ok {
		ctx.Set(authentication.UserKey, authUser)
		ctx.JSON(http.StatusOK, gin.H{"user": authUser})
		return
	}

	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
}

func logoutRoute(ctx *gin.Context) {
	session := sessions.Default(ctx)

	mysession := session.Get(authentication.SessionKey)
	if _, ok := mysession.(authentication.User); ok {
		session.Delete(authentication.SessionKey)
		if err := session.Save(); err != nil {
			logrus.Errorf("Failed to save session: %v", err)
		}
	}

	ctx.JSON(http.StatusOK, nil)
	return
}

func oauthRoute(ctx *gin.Context) {
	var (
		ok       bool
		authUser authentication.User
		user     *github.User
		context  = context.Background()
	)

	// Handle the exchange code to initiate a transport.
	session := sessions.Default(ctx)
	mysession := session.Get(authentication.SessionKey)
	if authUser, ok = mysession.(authentication.User); ok {
		ctx.Set(authentication.UserKey, authUser)
		ctx.Next()
		return
	}

	sessionState := session.Get(authentication.StateKey)
	retrievedState := ctx.Query(authentication.StateKey)
	if retrievedState != sessionState {
		loginRedirect(ctx)
		return
	}

	tok, err := conf.Exchange(context, ctx.Query(CodeKey))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("Failed to do exchange: %v", err))
		return
	}
	client := github.NewClient(conf.Client(context, tok))
	user, _, err = client.Users.Get(context, "")
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("Failed to get user: %v", err))
		return
	}

	auth := models.UserAuth{
		ProviderUserID: fmt.Sprintf("%d", user.GetID()),
		Email:          user.GetEmail(),
		Provider:       ProviderName,
	}

	if err := db.Debug().First(&auth, "provider = ? and provider_user_id = ?", ProviderName, auth.ProviderUserID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		u := models.User{
			Name: user.GetName(),
			Role: "",
			Auth: []models.UserAuth{auth},
		}
		if err := db.Create(&u).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "detail": err.Error()})
			return
		}
	}

	// save userinfo, which could be used in Handlers
	authUser = authentication.User{
		Login: user.GetLogin(),
		Name:  user.GetName(),
		URL:   user.GetURL(),
	}
	ctx.Set(authentication.UserKey, authUser)

	// populate cookie
	session.Set(authentication.SessionKey, authUser)
	if err := session.Save(); err != nil {
		logrus.Errorf("[Github-OAuth2]Failed to save session: %v", err)
	}

	ctx.Redirect(http.StatusTemporaryRedirect, "/")
}
