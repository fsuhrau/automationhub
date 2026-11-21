package api

import "github.com/gin-gonic/gin"

// registerSettingsRoutes adds the project settings. They hand out and revoke
// tokens, so only a browser session may use them: a project token must not
// mint a node token, or tokens for other projects.
func (s *Service) registerSettingsRoutes(api *gin.RouterGroup) {
	settings := api.Group("/:project_id/settings", s.requireSession)
	settings.GET("/users", s.getUsers)

	settings.POST("/access_token", s.createAccessToken)
	settings.GET("/access_tokens", s.getAccessTokens)
	settings.DELETE("/access_token/:token_id", s.deleteAccessToken)

	settings.POST("/node", s.createNode)
	settings.GET("/nodes", s.getNodes)
	settings.DELETE("/nodes/:node_id", s.deleteNode)
	settings.GET("/nodes/:node_id", s.getNodeStatus)
	settings.POST("/nodes/:node_id/token", s.issueNodeTokenHandler)
}
