package gui

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

// List all articles
func Index(c *gin.Context) {

	c.HTML(http.StatusOK, "gui/index", gin.H{
	})
}

func UploadFile(c *gin.Context) {

	file, err := c.FormFile("test_target")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	filename := filepath.Base(file.Filename)
	os.MkdirAll("upload", os.ModePerm)

	if err := c.SaveUploadedFile(file, filepath.Join("upload", filename)); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully.", file.Filename))
}

