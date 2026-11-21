package oauth2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/fsuhrau/automationhub/authentication/oauthflow/flowtest"
	"gorm.io/gorm"
)

func TestProfileErrors(t *testing.T) {
	closed := httptest.NewServer(http.NotFoundHandler())
	closed.Close()
	for name, url := range map[string]string{
		"bad url":     "http://%zz",
		"unreachable": closed.URL,
	} {
		t.Run(name, func(t *testing.T) {
			s := strategy{endpoints: Endpoints{UserURL: url}}
			if _, err := s.Profile(context.Background(), http.DefaultClient); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestStrategyConstants(t *testing.T) {
	s := strategy{endpoints: Endpoints{AuthURL: "a", TokenURL: "t"}}
	if s.Name() != ProviderName || s.SuccessRedirect() != "" || s.Endpoint().AuthURL != "a" || s.Endpoint().TokenURL != "t" {
		t.Errorf("%+v", s)
	}
}

func TestGenericOAuth2Flow(t *testing.T) {
	flowtest.Run(t, flowtest.Case{
		Name:           ProviderName,
		UserJSON:       `{"userid":"octo","name":"Octo Cat"}`,
		CallbackStatus: http.StatusOK,
		Build: func(db *gorm.DB, idp string, s oauthflow.Settings) (*oauthflow.Flow, error) {
			return New(db, Endpoints{AuthURL: idp + "/auth", TokenURL: idp + "/token", UserURL: idp + "/user"}, s)
		},
	})
}
