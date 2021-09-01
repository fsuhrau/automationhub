package api

import (
	"github.com/fsuhrau/automationhub/device"
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

func (s *ApiService) runTest(session *Session, c *gin.Context) {
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

	var devs []device.Device
	for _, d := range devices {
		if dev := s.devicesManager.GetDevice(d.DeviceIdentifier); dev != nil {
			devs = append(devs, dev)
		}
	}

	if test.TestConfig.Type == models.TestTypeUnity {
		tr := unity.New(s.db)
		tr.Initialize(test)
		tr.Run(devs, app)
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
