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

			if err := db.First(&token, "token = ?", xauth).Error; errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}

			if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized token expired"})
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
			ctx.Next()
			return
		}

		handler(ctx)
	}
}
