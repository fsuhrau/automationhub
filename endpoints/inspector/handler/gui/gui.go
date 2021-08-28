package gui

import (
	"fmt"
	"github.com/fsuhrau/automationhub/endpoints/inspector/handler/visitor"
	"github.com/fsuhrau/automationhub/hub/manager"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// List all articles
func Index(m manager.Devices) func(*gin.Context) {
	return func(c *gin.Context) {
		deviceList := visitor.DeviceList(m)

		files, _ := ioutil.ReadDir("./upload")
		var pathes []string
		pathes = append(pathes, "UnityEditor")
		pathes = append(pathes, "/Users/fabian.suhrau/projects/game_foe_mobile1/proj.ios/build/Debug/foe_mobile_develop.app")
		pathes = append(pathes, "/Users/fabian.suhrau/projects/game_mobilecity1/frontend/BuildPipeline/builds/android/Client.apk")

		for _, f := range files {
			pathes = append(pathes, filepath.Join("upload", f.Name()))
		}
		c.HTML(http.StatusOK, "gui/index", gin.H{
			"apps":    pathes,
			"devices": deviceList,
		})
	}
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
