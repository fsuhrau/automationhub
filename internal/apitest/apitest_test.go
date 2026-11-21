package apitest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestDoServesTheRealRoutes(t *testing.T) {
	env := apitest.New(t)
	testkit.Project(env.DB, testkit.WithName[models.Project]("alpha"))

	rec := env.Do(http.MethodGet, "/api/projects", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
	if got := testkit.Decode[[]models.Project](t, rec); len(got) != 1 || got[0].Name != "alpha" {
		t.Fatalf("projects = %+v", got)
	}
}

func TestRequestSendsThePreparedRequest(t *testing.T) {
	env := apitest.New(t, apitest.WithConfig(config.Service{}), apitest.WithNodeURL("http://node"))
	req := httptest.NewRequest(http.MethodGet, "/api/unknown/overview", nil)

	if rec := env.Request(req); rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestWithUserSignsTheRequestsIn(t *testing.T) {
	env := apitest.New(t, apitest.WithUser(authentication.User{Name: "Alice"}))
	testkit.User(env.DB, testkit.WithName[models.User]("Alice"))

	if rec := env.Do(http.MethodGet, "/api/me", nil); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
}

func TestWithAccessTokenMarksTokenRequests(t *testing.T) {
	env := apitest.New(t, apitest.WithAccessToken(7))

	if rec := env.Do(http.MethodPost, "/api/me/tokens", map[string]string{"name": "x"}); rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body %s", rec.Code, rec.Body)
	}
}

func TestNewExposesTheFakesAndRoutes(t *testing.T) {
	env := apitest.New(t)

	if env.Devices == nil || env.Nodes == nil || len(env.Router.Routes()) == 0 {
		t.Fatal("env is not wired")
	}
}
