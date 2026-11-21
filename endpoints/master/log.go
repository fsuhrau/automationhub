package master

import (
	"fmt"
	"net/http"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gin-gonic/gin"
)

type logRequest struct {
	Node        string  `json:"node"`
	DeviceID    string  `json:"deviceId"`
	Source      string  `json:"source"`
	Type        int     `json:"type"`
	Message     string  `json:"message"`
	Checkpoint  string  `json:"checkpoint"`
	Cpu         float64 `json:"cpu"`
	Fps         float64 `json:"fps"`
	Mem         float64 `json:"mem"`
	VertexCount float64 `json:"vertexCount"`
	Triangles   float64 `json:"triangles"`
	Other       string  `json:"other"`
	Key         string  `json:"key"`
	Value       string  `json:"value"`
}

func (s *nodeMaster) Log(c *gin.Context) {
	var req logRequest

	if err := c.BindJSON(&req); err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if device, _ := s.dm.GetDevice(req.DeviceID); device != nil {
		req.applyTo(device)
	}
}

func (req logRequest) applyTo(device device.Device) {
	switch req.Type {
	case 0:
		device.LogPerformance(req.Checkpoint, req.Cpu, req.Fps, req.Mem, req.VertexCount, req.Triangles, req.Other)
	case 1:
		device.Data(req.Source, req.Message)
	case 2:
		device.Log(req.Source, req.Message)
	case 3:
		device.Error(req.Source, req.Message)
	case 4:
		device.Passed(req.Other == "passed")
	case 5:
		device.SetKey(req.Key, req.Value)
	default:
		fmt.Println("unhandled: " + req.Source + " " + req.Message)
	}
}
