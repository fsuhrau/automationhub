package oauth2

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	ProviderName = "oauth2"
	CodeKey      = "code"
)

type Credentials struct {
	ClientID     string `json:"clientid"`
	ClientSecret string `json:"secret"`
}

var (
	conf           *oauth2.Config
	state          string
	store          sessions.CookieStore
	userRequestURL string
)

func randToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		logrus.Fatalf("[OAuth2] Failed to read rand: %v\n", err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func Setup(redirectURL, authURL, tokenURL, userURL, credFile string, scopes []string, secret []byte) {
	userRequestURL = userURL
	store = sessions.NewCookieStore(secret)
	var c Credentials
	file, err := ioutil.ReadFile(credFile)
	if err != nil {
		logrus.Fatalf("[OAuth2] File error: %v\n", err)
	}
	err = json.Unmarshal(file, &c)
	if err != nil {
		logrus.Fatalf("[OAuth2] Failed to unmarshal client credentials: %v\n", err)
	}
	conf = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
}

func Session(name string) gin.HandlerFunc {
	return sessions.Sessions(name, store)
}

func loginRedirect(ctx *gin.Context) {
	state = randToken()
	session := sessions.Default(ctx)
	session.Set(authentication.StateKey, state)
	session.Save()
	ctx.Redirect(http.StatusTemporaryRedirect, GetLoginURL(state))
}

func GetLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

type userResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Login string `json:"userid"`
}

func init() {
	gob.Register(authentication.User{})
}

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			ok       bool
			authUser authentication.User
			user     userResponse
		)

		// Handle the exchange code to initiate a transport.
		session := sessions.Default(ctx)
		mysession := session.Get(authentication.SessionKey)
		if authUser, ok = mysession.(authentication.User); ok {
			ctx.Set(authentication.UserKey, authUser)
			ctx.Next()
			return
		}

		retrievedState := session.Get(authentication.StateKey)
		if retrievedState != ctx.Query(authentication.StateKey) {
			loginRedirect(ctx)
			ctx.Abort()
			return
		}

		// TODO: oauth2.NoContext -> context.Context from stdlib
		tok, err := conf.Exchange(oauth2.NoContext, ctx.Query(CodeKey))
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("Failed to do exchange: %v", err))
			return
		}

		client := conf.Client(oauth2.NoContext, tok)

		req, err := http.NewRequest("GET", userRequestURL, nil)
		// Headers
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tok.AccessToken))
		req.Header.Add("Content-Type", "application/json; charset=utf-8")

		// Fetch Request
		resp, err := client.Do(req)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("Failed to get user: %v", err))
			return
		}

		respBody, _ := io.ReadAll(resp.Body)
		json.Unmarshal(respBody, &user)

		// save userinfo, which could be used in Handlers
		authUser = authentication.User{
			Login: user.Login,
			Name:  user.Name,
		}
		ctx.Set(authentication.UserKey, authUser)

		// populate cookie
		session.Set(authentication.SessionKey, authUser)
		if err := session.Save(); err != nil {
			logrus.Errorf("Failed to save session: %v", err)
		}
	}
}
