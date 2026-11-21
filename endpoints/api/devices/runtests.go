package devices

import (
	"net/http"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/gin-gonic/gin"
)

type runTestsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h *Handler) runTests(c *gin.Context) {
	var req svc.TestRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		httpx.Fail(c, err)
		return
	}
	req.DeviceID = c.Param("device_id")
	if err := h.svc.RunTests(c.Request.Context(), httpx.Project(c), req); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, runTestsResponse{Success: true, Message: "Execution Finished"})
}
