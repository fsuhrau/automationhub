package bundles

import (
	"net/http"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/gin-gonic/gin"
)

func (h *Handler) upload(c *gin.Context) {
	src, header, err := c.Request.FormFile(uploadField)
	if err != nil {
		httpx.Fail(c, apperr.Invalid("get form err: %s", err))
		return
	}
	defer func() { _ = src.Close() }()
	bundle, err := h.svc.Upload(c.Request.Context(), scope(c), appsvc.UploadFile{Name: header.Filename, Content: src})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, bundle)
}
