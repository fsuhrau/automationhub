package api

import (
	"github.com/gin-gonic/gin"
)

func (s *ApiService) error(c *gin.Context, status int, err error) {
	type Response struct {
		Message string
	}

	c.JSON(status, &Response{
		Message:   err.Error(),
	})
}