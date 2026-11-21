package devices

import (
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	svc "github.com/fsuhrau/automationhub/service/device"
	"github.com/gin-gonic/gin"
)

type runTestsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// invalid classifies a request body that could not be read.
func invalid(err error) error { return apperr.Invalid("%v", err) }

func (h *Handler) runTests(c *gin.Context) {
	var req svc.TestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, invalid(err))
		return
	}
	req.DeviceID = c.Param("device_id")
	if err := h.svc.RunTests(c.Request.Context(), httpx.Project(c), req); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, runTestsResponse{Success: true, Message: "Execution Finished"})
}
