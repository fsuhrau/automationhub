package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/fsuhrau/automationhub/authentication/oauthflow/flowtest"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestNewUsesThePublicGitHub(t *testing.T) {
	cred := filepath.Join(t.TempDir(), "cred.json")
	if err := os.WriteFile(cred, []byte(`{"clientid":"c","secret":"s"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	flow, err := New(nil, oauthflow.Settings{CredFile: cred, Secret: []byte("test-secret-test-secret-test-sec")})
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(flow.Session("hub"))
	flow.Routes(r)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/auth/oauth2", nil))
	if loc := w.Header().Get("Location"); w.Code != http.StatusTemporaryRedirect || !strings.HasPrefix(loc, "https://github.com/login/oauth/authorize?") {
		t.Errorf("%d %s", w.Code, loc)
	}
}

func TestProfileRejectsABrokenAPIURL(t *testing.T) {
	s := strategy{endpoints: Endpoints{APIURL: "http://%zz"}}
	if _, err := s.Profile(context.Background(), http.DefaultClient); err == nil {
		t.Fatal("expected an error")
	}
}

func TestStrategyConstants(t *testing.T) {
	s := strategy{}
	if s.Name() != ProviderName || s.SuccessRedirect() != "/" {
		t.Errorf("%s %s", s.Name(), s.SuccessRedirect())
	}
}

func TestGitHubFlow(t *testing.T) {
	flowtest.Run(t, flowtest.Case{
		Name:              ProviderName,
		UserJSON:          `{"id":42,"login":"octo","name":"Octo Cat","email":"octo@example.com","url":"https://api.example/users/octo"}`,
		CallbackStatus:    http.StatusTemporaryRedirect,
		FailsOnBrokenUser: true,
		Build: func(db *gorm.DB, idp string, s oauthflow.Settings) (*oauthflow.Flow, error) {
			return NewWithEndpoints(db, Endpoints{AuthURL: idp + "/auth", TokenURL: idp + "/token", APIURL: idp + "/"}, s)
		},
	})
}
