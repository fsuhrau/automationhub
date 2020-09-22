package sessions

import (
	"net/http"

	"github.com/fsuhrau/automationhub/inspector/handler/manager"
	"github.com/gin-gonic/gin"
)

func Index(sessionManager manager.SessionManager) func(*gin.Context) {
	return func(c *gin.Context) {
		sessionManager.GetSessions()
		sessions := sessionManager.GetSessions()
		c.HTML(http.StatusOK, "sessions/index", gin.H{
			"sessions": sessions,
		})
	}
}

func Show(sessionManager manager.SessionManager) func(*gin.Context) {
	return func(c *gin.Context) {
		sessionID := c.Param("sessionID")
		c.SetCookie("session_id", sessionID, 356, "/", "", true, false)
		session := sessionManager.GetSessionDetails(sessionID)
		c.HTML(http.StatusOK, "sessions/show", gin.H{
			"session": session,
		})
	}
}
