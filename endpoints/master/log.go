package master

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *nodeMaster) Log(c *gin.Context) {
	type request struct {
		Node        string  `json:"node"`
		DeviceID    string  `json:"device_id"`
		Source      string  `json:"source"`
		Type        int     `json:"type"`
		Message     string  `json:"message"`
		Checkpoint  string  `json:"checkpoint"`
		Cpu         float64 `json:"cpu"`
		Fps         float64 `json:"fps"`
		Mem         float64 `json:"mem"`
		VertexCount float64 `json:"vertex_count"`
		Triangles   float64 `json:"triangles"`
		Other       string  `json:"other"`
	}

	var req request

	if err := c.BindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	device, _ := s.dm.GetDevice(req.DeviceID)
	if device != nil {
		switch req.Type {
		case 0:
			device.LogPerformance(req.Checkpoint, req.Cpu, req.Fps, req.Mem, req.VertexCount, req.Triangles, req.Other)
		case 1:
			device.Data(req.Source, req.Message)
		case 2:
			device.Log(req.Source, req.Message)
		case 3:
			device.Error(req.Source, req.Message)
		default:
			fmt.Println("unhandled: " + req.Source + " " + req.Message)
		}
	}
}
