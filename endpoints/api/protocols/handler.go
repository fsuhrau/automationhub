// Package protocols holds the HTTP handlers of the protocols of a test run:
// the protocol page and the baselines its performance is compared with.
// They parse the request, call the protocol service and write the answer.
package protocols

import (
	"context"
	"net/http"
	"strconv"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/protocol"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers need from the protocol service.
type Service interface {
	Page(ctx context.Context, key protocol.Key) (*models.TestRun, error)
	Baselines(ctx context.Context, key protocol.Key, f protocol.Filter) (*protocol.Baselines, error)
	Baseline(ctx context.Context, key protocol.Key, choice protocol.Choice) (*protocol.Baseline, error)
}

// Handler serves the protocol routes of an app.
type Handler struct {
	svc Service
}

// New builds the handler on the service.
func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterProtocols adds the protocol page to the group of an app, which
// must run the ResolveProject and ResolveApp middlewares.
func (h *Handler) RegisterProtocols(appAPI *gin.RouterGroup) {
	appAPI.GET("/test/:test_id/run/:run_id/:protocol_id", h.page)
}

// RegisterBaselines adds the baseline routes to the group of an app.
func (h *Handler) RegisterBaselines(appAPI *gin.RouterGroup) {
	appAPI.GET("/test/:test_id/run/:run_id/:protocol_id/baselines", h.baselines)
	appAPI.GET("/test/:test_id/run/:run_id/:protocol_id/baseline/:baseline_run_id", h.baseline)
}

// protocolKey is the protocol of the path inside the project and app of the
// request.
func protocolKey(c *gin.Context) (protocol.Key, error) {
	var ids [3]uint
	for i, name := range []string{"test_id", "run_id", "protocol_id"} {
		id, err := httpx.ParseID(c, name)
		if err != nil {
			return protocol.Key{}, err
		}
		ids[i] = id
	}
	scope := testsvc.Scope{ProjectID: httpx.Project(c).ID, AppID: httpx.App(c).ID}
	return protocol.Key{TestRef: testsvc.TestRef{Scope: scope, TestID: ids[0]}, RunID: ids[1], ProtocolID: ids[2]}, nil
}

func (h *Handler) page(c *gin.Context) {
	key, err := protocolKey(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	run, err := h.svc.Page(c.Request.Context(), key)
	respond(c, run, err)
}

func (h *Handler) baselines(c *gin.Context) {
	key, err := protocolKey(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	filter := protocol.Filter{
		Tag: c.Query("tag"), Query: c.Query("q"), AnyDevice: c.Query("anyDevice") == "true",
		Limit: httpx.Limit(c, protocol.DefaultCandidates, protocol.MaxCandidates),
	}
	found, err := h.svc.Baselines(c.Request.Context(), key, filter)
	respond(c, found, err)
}

func (h *Handler) baseline(c *gin.Context) {
	key, err := protocolKey(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	runID, err := httpx.ParseID(c, "baseline_run_id")
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	found, err := h.svc.Baseline(c.Request.Context(), key, protocol.Choice{RunID: runID, DeviceID: deviceID(c)})
	respond(c, found, err)
}

// deviceID reads the optional deviceId query: nil unless it is a number.
func deviceID(c *gin.Context) *uint {
	v, err := strconv.ParseUint(c.Query("deviceId"), 10, 0)
	if err != nil {
		return nil
	}
	id := uint(v)
	return &id
}

// respond writes the answer, or the error that kept it from loading.
func respond(c *gin.Context, answer any, err error) {
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, answer)
}
