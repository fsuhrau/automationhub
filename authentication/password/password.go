package password

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	ProviderName = "password"
)

var (
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

func Setup(gormDb *gorm.DB, secret []byte) {
	store = sessions.NewCookieStore(secret)
	db = gormDb
}

func Session(name string) gin.HandlerFunc {
	return sessions.Sessions(name, store)
}

func loginRedirect(ctx *gin.Context) {
	state = randToken()
	session := sessions.Default(ctx)
	session.Set(authentication.StateKey, state)
	session.Save()
	ctx.Redirect(http.StatusTemporaryRedirect, GetLoginURL())
}

func GetLoginURL() string {
	return "/login"
}

func init() {
	gob.Register(authentication.User{})
}

func Routes(r *gin.Engine) {
	r.POST("/auth/login", LoginRoute)
	r.POST("/auth/register", RegisterRoute)
	r.GET("/auth/session", SessionRoute)
	r.POST("/auth/logout", LogoutRoute)
}

func Auth() gin.HandlerFunc {
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

		loginRedirect(ctx)
		ctx.Abort()
	}
}

func SessionRoute(ctx *gin.Context) {
	session := sessions.Default(ctx)

	mysession := session.Get(authentication.SessionKey)
	if authUser, ok := mysession.(authentication.User); ok {
		ctx.Set(authentication.UserKey, authUser)
		ctx.JSON(http.StatusOK, gin.H{"user": authUser})
		return
	}

	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
	return
}

func LogoutRoute(ctx *gin.Context) {
	session := sessions.Default(ctx)

	mysession := session.Get(authentication.SessionKey)
	if _, ok := mysession.(authentication.User); ok {
		session.Delete(authentication.SessionKey)
		if err := session.Save(); err != nil {
			logrus.Errorf("[Github-OAuth2]Failed to save session: %v", err)
		}
	}

	ctx.JSON(http.StatusOK, nil)
	return
}

func LoginRoute(ctx *gin.Context) {
	session := sessions.Default(ctx)

	var (
		provider *models.UserAuth
		authUser authentication.User
	)

	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if err := db.Where("email = ? and provider = ?", req.Email, ProviderName).Preload("User").First(&provider).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(provider.PasswordHash), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// save userinfo, which could be used in Handlers
	authUser = authentication.User{
		Login: provider.User.Name,
		Role:  provider.User.Name,
		Name:  provider.User.Name,
		Email: provider.Email,
	}
	ctx.Set(authentication.UserKey, authUser)

	// populate cookie
	session.Set(authentication.SessionKey, authUser)
	if err := session.Save(); err != nil {
		logrus.Errorf("[Github-OAuth2]Failed to save session: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"user": authUser})
}

func RegisterRoute(ctx *gin.Context) {

	var (
		authUser authentication.User
	)

	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.Name = strings.TrimSpace(req.Name)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Name: req.Name,
		Role: "",
		Auth: []models.UserAuth{
			{
				Email:        req.Email,
				Provider:     ProviderName,
				PasswordHash: string(hashedPassword),
			},
		},
	}

	if err := db.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// save userinfo, which could be used in Handlers
	authUser = authentication.User{
		Login: user.Name,
		Name:  user.Name,
		Email: req.Email,
	}
	ctx.Set(authentication.UserKey, authUser)

	// populate cookie
	session := sessions.Default(ctx)
	session.Set(authentication.SessionKey, authUser)
	if err := session.Save(); err != nil {
		logrus.Errorf("[Github-OAuth2]Failed to save session: %v", err)
	}

	ctx.JSON(http.StatusOK, gin.H{"user": authUser})
}
