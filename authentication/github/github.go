package github

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	oauth2gh "golang.org/x/oauth2/github"
)

type Credentials struct {
	ClientID     string `json:"clientid"`
	ClientSecret string `json:"secret"`
}

var (
	conf  *oauth2.Config
	state string
	store sessions.CookieStore
)

func randToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		logrus.Fatalf("[Github-OAuth2] Failed to read rand: %v\n", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func Setup(redirectURL, credFile string, scopes []string, secret []byte) {
	store = sessions.NewCookieStore(secret)
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
	session.Set("state", state)
	session.Save()
	ctx.Redirect(http.StatusTemporaryRedirect, GetLoginURL(state))
}

func GetLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

type AuthUser struct {
	Login   string `json:"login"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	URL     string `json:"url"`
}

func init() {
	gob.Register(AuthUser{})
}

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			ok       bool
			authUser AuthUser
			user     *github.User
		)

		// Handle the exchange code to initiate a transport.
		session := sessions.Default(ctx)
		mysession := session.Get("ginoauthgh")
		if authUser, ok = mysession.(AuthUser); ok {
			ctx.Set("user", authUser)
			ctx.Next()
			return
		}

		retrievedState := session.Get("state")
		if retrievedState != ctx.Query("state") {
			loginRedirect(ctx)
			ctx.Abort()
			return
		}

		// TODO: oauth2.NoContext -> context.Context from stdlib
		tok, err := conf.Exchange(oauth2.NoContext, ctx.Query("code"))
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("Failed to do exchange: %v", err))
			return
		}
		client := github.NewClient(conf.Client(oauth2.NoContext, tok))
		user, _, err = client.Users.Get(oauth2.NoContext, "")
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("Failed to get user: %v", err))
			return
		}

		// save userinfo, which could be used in Handlers
		authUser = AuthUser{
			Login: *user.Login,
			Name:  *user.Name,
			URL:   *user.URL,
		}
		ctx.Set("user", authUser)

		// populate cookie
		session.Set("ginoauthgh", authUser)
		if err := session.Save(); err != nil {
			logrus.Errorf("[Github-OAuth2]Failed to save session: %v", err)
		}
	}
}
