package sessions

import (
	"github.com/fsuhrau/automationhub/hub/manager"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Index(sessionManager manager.Sessions) func(*gin.Context) {
	return func(c *gin.Context) {
		sessionManager.GetSessions()
		sessions := sessionManager.GetSessions()
		c.HTML(http.StatusOK, "sessions/index", gin.H{
			"sessions": sessions,
		})
	}
}

func Show(sessionManager manager.Sessions) func(*gin.Context) {
	return func(c *gin.Context) {
		sessionID := c.Param("sessionId")
		c.SetCookie("session_id", sessionID, 356, "/", "", true, false)
		session := sessionManager.GetSessionDetails(sessionID)
		c.HTML(http.StatusOK, "sessions/show", gin.H{
			"session": session,
		})
	}
}
