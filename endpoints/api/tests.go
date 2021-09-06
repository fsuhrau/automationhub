package api

import (
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/unity"
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
	if err := s.db.Find(&tests).Preload("TestConfig").Error; err != nil {
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
		s.error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, test)
}

func (s *ApiService) runTest(c *gin.Context) {
	type Request struct {
		AppID   uint
		Devices []uint
	}

	testId := c.Param("test_id")
	var req Request
	if err := c.Bind(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	var app models.App
	if err := s.db.First(&app, req.AppID).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	var devices []models.Device
	if err := s.db.Find(&devices, req.Devices).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	for i := range devices {
		devices[i].Dev = s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
	}

	if test.TestConfig.Type == models.TestTypeUnity {
		tr := unity.New(s.db, s.hostIP, s.devicesManager)
		if err := tr.Initialize(test); err != nil {
			s.error(c, http.StatusInternalServerError, err) // Todo status code
			return
		}
		if err := tr.Run(devices, app); err != nil {
			s.error(c, http.StatusInternalServerError, err) // Todo status code
			return
		}
	}

	c.JSON(http.StatusOK, test)
}

func (s *ApiService) getTestRuns(session *Session, c *gin.Context) {
	testId := c.Param("test_id")

	var testRuns []models.TestRun
	if err := s.db.Preload("Protocols").Find(&testRuns, "test_id = ?", testId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, testRuns)
}

func (s *ApiService) getLastTestRun(session *Session, c *gin.Context) {
	testId := c.Param("test_id")
	_ = testId

	var run models.TestRun
	if err := s.db.Preload("Protocols").Preload("Protocols.Device").Preload("Log").Where("test_id = ?", testId).Order("id desc").First(&run).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, run)
}


func (s *ApiService) getTestRun(session *Session, c *gin.Context) {
	testId := c.Param("test_id")
	_ = testId
	runId := c.Param("run_id")

	var run models.TestRun
	if err := s.db.Preload("Protocols").Preload("Log").First(&run, runId).Error; err != nil {
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
