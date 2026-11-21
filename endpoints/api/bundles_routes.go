package api

import "github.com/gin-gonic/gin"

func (s *Service) registerBundleRoutes(appApi *gin.RouterGroup) {
	appApi.GET("/bundles", s.WithApp(s.getBinaries))
	appApi.GET("/bundle/:binary_id", s.WithApp(s.getBinary))
	appApi.GET("/bundle/:binary_id/icon", s.WithApp(s.getBinaryIcon))
	appApi.POST("/bundle/:binary_id/read", s.WithApp(s.readBinaryAgain))
	appApi.POST("/bundle/:binary_id/archive", s.WithApp(s.archiveBinary))
	appApi.PUT("/bundle/:binary_id", s.WithApp(s.updateBinary))
	appApi.DELETE("/bundle/:binary_id", s.WithApp(s.deleteBinary))
	appApi.POST("/upload", s.WithApp(s.uploadBinary))
}
