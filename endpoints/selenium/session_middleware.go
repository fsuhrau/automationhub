package selenium

import (
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
)

func SessionMiddleware(s manager.Sessions) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("sessionID")
		session, _ := s.GetSession(sessionID)
		if session != nil {
			c.Set("session", session)
			c.Next()
		}
	}
}

func HandleWithSession(f func(*Session, *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		s, _ := c.Get("session")
		if s == nil {
			return
		}
		session := s.(*Session)
		// session.LastAccess = time.Now()
		f(session, c)
	}
}
