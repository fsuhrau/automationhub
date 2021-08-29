package api

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
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
*/

func (s *ApiService) getTests(session *Session, c *gin.Context) {
	var tests []models.Test
	if err := s.db.Find(&tests).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, tests)
}

func (s *ApiService) newTest(session *Session, c *gin.Context) {
	var test models.Test
	if err := c.Bind(&test); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	if err := s.db.Create(&test).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, test)
}

func (s *ApiService) getTest(session *Session, c *gin.Context) {
	testId := c.Param("test_id")

	var test models.Test
	if err := s.db.First(&test, testId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, test)
}

func (s *ApiService) getTestRuns(session *Session, c *gin.Context) {
	testId := c.Param("test_id")

	var testRuns []models.TestRun
	if err := s.db.Find(&testRuns, "testId = ?", testId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, testRuns)
}

func (s *ApiService) getTestRun(session *Session, c *gin.Context) {
	testId := c.Param("test_id")
	_ = testId
	runId := c.Param("run_id")

	var run models.TestRun
	if err := s.db.Find(&run, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, run)
}

func (s *ApiService) startTest(session *Session, c *gin.Context) {
	type Result struct {
	}
	c.JSON(http.StatusOK, &Result{})
}

func (s *ApiService) getStatus(session *Session, c *gin.Context) {
	type Result struct {
	}
	c.JSON(http.StatusOK, &Result{})
}