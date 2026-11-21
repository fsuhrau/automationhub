package api

import "github.com/gin-gonic/gin"

func (s *Service) registerProjectRoutes(api *gin.RouterGroup) {
	api.GET("/projects", s.getProjects)
	api.POST("/project", s.createProject)
	api.DELETE("/project/:project_id", s.deleteProject)
	api.PUT("/project/:project_id", s.updateProject)
}
