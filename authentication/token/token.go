package token

import (
	"errors"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func SessionHandler(db *gorm.DB, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var (
			token   models.AccessToken
			project models.Project
		)

		xauth := ctx.GetHeader("X-Auth-Token")
		if xauth != "" {

			// fail closed: only a token that was found signs the request in
			if err := lookup(db, xauth, &token); errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			} else if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "token lookup failed"})
				return
			}

			if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized token expired"})
				return
			}

			// a personal token acts as its user, in every project
			if token.UserID != nil {
				if !authenticatePersonal(db, ctx, &token) {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
				ctx.Next()
				return
			}

			if token.ProjectID != nil {
				projectParam := ctx.Param("project_id")
				if projectParam == "" {
					// tokens work only with params
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "X-Auth-Token is project dependent"})
					return
				}

				if err := db.First(&project, "id = ?", token.ProjectID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}

				if projectParam != project.Identifier {
					ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "X-Auth-Token is project dependent"})
					return
				}
			}
			if token.NodeID != nil {
				var node models.Node

				if err := db.First(&node, "id = ?", token.NodeID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
					ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			}

			ctx.Set(authentication.UserKey, token.Name)
			ctx.Set(AuthTokenKey, token.ID)
			ctx.Next()
			return
		}

		handler(ctx)
	}
}
