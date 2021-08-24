package api

import (
	"github.com/fsuhrau/automationhub/hub/action"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *ApiService) getTests(session *Session, c *gin.Context) {
	type Response struct {
		Tests []*action.Test
	}
	log := session.GetLogger().WithField("prefix", "action")
	a := &action.TestsGet{}
	if err := s.devicesManager.SendAction(log, session, a); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, &Response{
		Tests: a.Tests,
	})
}

func (s *ApiService) startTest(session *Session, c *gin.Context) {
	type Response struct {
		Status string
	}
	// WIP
	log := session.GetLogger().WithField("prefix", "action")
	a := &action.TestStart{
		TestClass: "Innium.IntegrationTests.SmokeTests",
		// TestMethod: "System.Collections.IEnumerator ShortSanityTest()",
	}
	if err := s.devicesManager.SendAction(log, session, a); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, &Response{
		Status:     "success",
	})
}

func (s *ApiService)getStatus(session *Session, c *gin.Context) {
	type Response struct {
		Status string
	}

	c.JSON(http.StatusOK, &Response{
		Status:     "success",
	})
}