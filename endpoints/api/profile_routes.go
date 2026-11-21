package api

import "github.com/gin-gonic/gin"

func (s *Service) registerProfileRoutes(api *gin.RouterGroup) {
	me := api.Group("/me")
	me.GET("", s.getProfile)
	me.GET("/tokens", s.getPersonalTokens)
	me.POST("/tokens", s.createPersonalToken)
	me.DELETE("/tokens/:token_id", s.deletePersonalToken)
}
