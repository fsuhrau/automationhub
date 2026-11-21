package oauthflow_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/fsuhrau/automationhub/authentication/oauthflow/flowtest"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// idpStrategy reads the user from the fake provider like a minimal OAuth2 provider.
type idpStrategy struct{ idp string }

func (idpStrategy) Name() string            { return "fake" }
func (idpStrategy) SuccessRedirect() string { return "/" }

func (s idpStrategy) Endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{AuthURL: s.idp + "/auth", TokenURL: s.idp + "/token"}
}

func (s idpStrategy) Profile(ctx context.Context, c *http.Client) (oauthflow.Profile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.idp+"/user", nil)
	if err != nil {
		return oauthflow.Profile{}, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return oauthflow.Profile{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return oauthflow.Profile{}, errors.New("user endpoint: " + resp.Status)
	}
	var u struct{ Login, Name string }
	_ = json.NewDecoder(resp.Body).Decode(&u)
	return oauthflow.Profile{ProviderUserID: u.Login, Login: u.Login, Name: u.Name}, nil
}

func TestFlowWithAStrategy(t *testing.T) {
	flowtest.Run(t, flowtest.Case{
		Name:              "fake",
		UserJSON:          `{"login":"octo","name":"Octo Cat"}`,
		CallbackStatus:    http.StatusTemporaryRedirect,
		FailsOnBrokenUser: true,
		Build: func(db *gorm.DB, idp string, s oauthflow.Settings) (*oauthflow.Flow, error) {
			return oauthflow.New(db, idpStrategy{idp: idp}, s)
		},
	})
}
