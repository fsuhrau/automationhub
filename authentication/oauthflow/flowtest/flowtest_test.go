package flowtest_test

import (
	"net/http"
	"testing"

	"github.com/fsuhrau/automationhub/authentication/oauth2"
	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/fsuhrau/automationhub/authentication/oauthflow/flowtest"
	"gorm.io/gorm"
)

func TestGenericOAuth2Flow(t *testing.T) {
	flowtest.Run(t, flowtest.Case{
		Name:     oauth2.ProviderName,
		UserJSON: `{"userid":"octo","name":"Octo Cat"}`,
		NoEmailUsersJSON: [2]string{
			`{"userid":"first","name":"First"}`,
			`{"userid":"second","name":"Second"}`,
		},
		CallbackStatus: http.StatusOK,
		Build: func(db *gorm.DB, idp string, s oauthflow.Settings) (*oauthflow.Flow, error) {
			return oauth2.New(db, oauth2.Endpoints{AuthURL: idp + "/auth", TokenURL: idp + "/token", UserURL: idp + "/user"}, s)
		},
	})
}
