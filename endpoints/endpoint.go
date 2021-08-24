package endpoints

import (
	"github.com/gin-gonic/gin"
)

type ServiceEndpoint interface {
	RegisterRoutes(*gin.Engine) error
}