package token

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NodeRoute tells whether a node token may be used for the path: the node
// endpoints of the master (/node/...) and the bundles it installs (/upload/...).
func NodeRoute(path string) bool {
	return strings.HasPrefix(path, "/node/") || strings.HasPrefix(path, "/upload/")
}

// SessionHandler authenticates requests that carry an X-Auth-Token header
// and hands every other request to handler (the browser session guard).
func SessionHandler(db *gorm.DB, handler gin.HandlerFunc) gin.HandlerFunc {
	g := guard{db: db}
	return func(ctx *gin.Context) {
		value := ctx.GetHeader("X-Auth-Token")
		if value == "" {
			handler(ctx)
			return
		}
		if d := g.check(ctx, value); d != nil {
			ctx.AbortWithStatusJSON(d.status, gin.H{"error": d.message})
			return
		}
		ctx.Next()
	}
}
