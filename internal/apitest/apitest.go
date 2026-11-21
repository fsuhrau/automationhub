// Package apitest builds the real api router for tests: the services the
// composition root (internal/wiring) builds and the same RegisterRoutes the
// hub calls, backed by an in-memory database and the testkit fakes, so tests
// go through HTTP and not through the internals of the api.
//
// The device control reaper runs until the test ends.
package apitest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/internal/wiring"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Env is a running api router and the fakes behind it.
type Env struct {
	DB      *gorm.DB
	Devices *testkit.FakeDeviceManager
	Nodes   *testkit.FakeNodes
	Router  *gin.Engine
	// Paths are the folders of the environment, temporary ones unless the
	// config of WithConfig names them.
	Paths config.Paths

	t   testing.TB
	cfg *settings
}

type settings struct {
	cfg        config.Service
	user       *authentication.User
	viaToken   bool
	tokenOwner uint
}

// Option changes how New builds the environment.
type Option func(*settings)

// WithConfig sets the service configuration.
func WithConfig(cfg config.Service) Option { return func(s *settings) { s.cfg = cfg } }

// WithNodeURL sets the node url the api hands out.
func WithNodeURL(url string) Option { return func(s *settings) { s.cfg.NodeUrl = url } }

// WithUser signs the requests in as a browser session of the user.
func WithUser(user authentication.User) Option { return func(s *settings) { s.user = &user } }

// WithAccessToken makes the requests look like they came in with the
// access token with the id, not with a browser session.
func WithAccessToken(id uint) Option {
	return func(s *settings) { s.viaToken, s.tokenOwner = true, id }
}

// New registers the api routes on a fresh router with an empty database.
func New(t testing.TB, opts ...Option) *Env {
	t.Helper()
	var cfg settings
	for _, opt := range opts {
		opt(&cfg)
	}
	cfg.cfg.Paths = withTempFolders(t, cfg.cfg.Paths)
	env := &Env{DB: testkit.DB(t), Devices: &testkit.FakeDeviceManager{}, Nodes: &testkit.FakeNodes{}, Paths: cfg.cfg.Paths, t: t, cfg: &cfg}
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	built := wiring.BuildAPI(wiring.Infra{DB: env.DB, Devices: env.Devices, Nodes: env.Nodes, Config: cfg.cfg, Logger: logger})
	built.Start(t.Context())

	gin.SetMode(gin.TestMode)
	env.Router = gin.New()
	failOn(t, built.Endpoint.RegisterRoutes(env.Router, env.Router.Group("", cfg.authenticate)))
	return env
}

// withTempFolders gives every folder the config leaves out a temporary one.
func withTempFolders(t testing.TB, paths config.Paths) config.Paths {
	t.Helper()
	return paths.WithDefaults(config.Paths{AppStore: t.TempDir(), Bundles: t.TempDir(), TestData: t.TempDir()})
}

// failOn stops the test when setting up the environment failed.
func failOn(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("apitest: register routes: %v", err)
	}
}

// authenticate puts the configured identity into the request context, the
// way the authentication middleware of the hub does.
func (s *settings) authenticate(c *gin.Context) {
	if s.user != nil {
		c.Set(authentication.UserKey, *s.user)
	}
	if s.viaToken {
		c.Set(token.AuthTokenKey, s.tokenOwner)
	}
}

// SignInAs makes the following requests come from the browser session of
// the user, in place of the one the environment started with.
func (e *Env) SignInAs(user authentication.User) { e.cfg.user = &user }

// Do sends a request to the router; a non-nil body is sent as JSON.
func (e *Env) Do(method, path string, body any) *httptest.ResponseRecorder {
	e.t.Helper()
	return testkit.DoJSON(e.t, e.Router, method, path, body)
}

// Request sends a prepared request, for tests that need headers or
// non-JSON bodies.
func (e *Env) Request(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.Router.ServeHTTP(rec, req)
	return rec
}
