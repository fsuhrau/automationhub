package api

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func (s *Service) error(c *gin.Context, status int, err error) {
	logrus.Error(err)
	c.AbortWithStatusJSON(status, &ErrorResponse{
		Message: err.Error(),
	})
}
