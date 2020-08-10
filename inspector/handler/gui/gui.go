package gui

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

// List all articles
func Index(c *gin.Context) {
	files, _ := ioutil.ReadDir("./upload")
	var pathes []string
	pathes = append(pathes, "/Users/fabian.suhrau/projects/game_foe_mobile1/proj.ios/build/Debug/foe_mobile_develop.app")
	for _, f := range files {
		pathes = append(pathes, filepath.Join("upload", f.Name()))
	}
	c.HTML(http.StatusOK, "gui/index", gin.H{
		"apps": pathes,
	})
}

func UploadFile(c *gin.Context) {

	file, err := c.FormFile("test_target")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	filename := filepath.Base(file.Filename)
	filePath := filepath.Join("upload", filename)
	os.MkdirAll("upload", os.ModePerm)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"filename": filename,
		"app_path": filePath,
	})
}

