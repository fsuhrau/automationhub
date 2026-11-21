package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
)

func TestRunTestRefusesAMalformedBody(t *testing.T) {
	env := apitest.New(t)
	project := testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "proj" })
	app := testkit.App(env.DB, func(a *models.App) { a.ProjectID = project.ID })

	rec := env.Do("POST", fmt.Sprintf("/api/proj/app/%d/test/1/run", app.ID), map[string]any{"appBinaryId": "x"})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400: %s", rec.Code, rec.Body)
	}
}
