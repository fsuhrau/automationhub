// Package runs holds the HTTP handlers of the test runs of an app: paging
// through them, the run page, starting and cancelling a run. They parse the
// request, call the run service and write the answer.
package runs

import (
	"context"
	"net/http"
	"strconv"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/endpoints/httpx"
	"github.com/fsuhrau/automationhub/service/run"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/gin-gonic/gin"
)

// Service is what the handlers need from the run service.
type Service interface {
	Start(ctx context.Context, req run.StartRequest) (run.RunRef, error)
	Cancel(ctx context.Context, projectID, runID uint) error
	Get(ctx context.Context, key run.Key) (*run.Page, error)
	Last(ctx context.Context, ref testsvc.TestRef) (*run.Page, error)
	List(ctx context.Context, q run.ListQuery) (*run.RunsPage, error)
}

// Handler serves the run routes of an app.
type Handler struct {
	svc Service
}

// New builds the handler on the service.
func New(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Register adds the routes to the group of an app, which must run the
// ResolveProject and ResolveApp middlewares.
func (h *Handler) Register(appAPI *gin.RouterGroup) {
	appAPI.GET("/test/:test_id/runs", h.list)
	appAPI.GET("/test/:test_id/runs/last", h.last)
	appAPI.GET("/test/:test_id/run/:run_id", h.get)
	appAPI.POST("/test/:test_id/run", h.start)
	appAPI.POST("/test/:test_id/run/:run_id/cancel", h.cancel)
}

// testRef is the test of the path inside the project and app of the
// request.
func testRef(c *gin.Context) (testsvc.TestRef, error) {
	id, err := httpx.ParseID(c, "test_id")
	if err != nil {
		return testsvc.TestRef{}, err
	}
	scope := testsvc.Scope{ProjectID: httpx.Project(c).ID, AppID: httpx.App(c).ID}
	return testsvc.TestRef{Scope: scope, TestID: id}, nil
}

// runKey is the run of the path inside the test of the path.
func runKey(c *gin.Context) (run.Key, error) {
	ref, err := testRef(c)
	if err != nil {
		return run.Key{}, err
	}
	id, err := httpx.ParseID(c, "run_id")
	return run.Key{TestRef: ref, RunID: id}, err
}

// queryID reads an optional id from the query; anything but a positive
// number is 0.
func queryID(c *gin.Context, name string) uint {
	v, err := strconv.ParseUint(c.Query(name), 10, 0)
	if err != nil {
		return 0
	}
	return uint(v)
}

func (h *Handler) list(c *gin.Context) {
	ref, err := testRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	query := run.ListQuery{TestRef: ref, Before: queryID(c, "before"), Limit: httpx.Limit(c, run.PageDefault, run.PageMax)}
	page, err := h.svc.List(c.Request.Context(), query)
	respond(c, page, err)
}

func (h *Handler) last(c *gin.Context) {
	ref, err := testRef(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	page, err := h.svc.Last(c.Request.Context(), ref)
	respond(c, page, err)
}

func (h *Handler) get(c *gin.Context) {
	key, err := runKey(c)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	page, err := h.svc.Get(c.Request.Context(), key)
	respond(c, page, err)
}

// StartRequest is the request body of a run: the bundle, or the start URL
// for web and editor apps, and the run parameters.
type StartRequest struct {
	AppBinaryID int    `json:"appBinaryId"`
	StartURL    string `json:"startUrl"`
	Params      string `json:"params"`
}

func (h *Handler) start(c *gin.Context) {
	testID, err := httpx.ParseID(c, "test_id")
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid test id"))
		return
	}
	var body StartRequest
	if err := httpx.BindJSON(c, &body); err != nil {
		httpx.Fail(c, err)
		return
	}
	ref, err := h.svc.Start(c.Request.Context(), run.StartRequest{
		Project: httpx.Project(c), TestID: testID, BundleID: uint(body.AppBinaryID),
		StartURL: body.StartURL, Params: body.Params,
	})
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, ref.Run)
}

func (h *Handler) cancel(c *gin.Context) {
	runID, err := httpx.ParseID(c, "run_id")
	if err != nil {
		httpx.Fail(c, apperr.Invalid("invalid run id"))
		return
	}
	if err := h.svc.Cancel(c.Request.Context(), httpx.Project(c).ID, runID); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// respond writes the page, or the error that kept it from loading.
func respond(c *gin.Context, page any, err error) {
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}
