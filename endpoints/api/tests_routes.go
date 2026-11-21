package api

import (
	testsapi "github.com/fsuhrau/automationhub/endpoints/api/tests"
	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/teststore"
	"github.com/gin-gonic/gin"
)

// testService is the test definition service on the database of this api,
// with the run state of this master.
func (s *Service) testService() *testsvc.Service {
	return testsvc.New(testsvc.Deps{Repo: teststore.New(s.db), RunState: s.runState})
}

func (s *Service) registerTestRoutes(appApi *gin.RouterGroup) {
	testsapi.New(s.testService()).Register(appApi)
}
