package api

import (
	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Message string
}

func (s *Service) error(c *gin.Context, status int, err error) {
	c.JSON(status, &ErrorResponse{
		Message:   err.Error(),
	})
}