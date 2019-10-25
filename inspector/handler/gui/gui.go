package gui

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// List all articles
func Index(c *gin.Context) {

	c.HTML(http.StatusOK, "gui/index", gin.H{
		// "markets": markets,
	})
}

func Start(c *gin.Context) {

	c.HTML(http.StatusOK, "gui/index", gin.H{
		// "markets": markets,
	})
}

