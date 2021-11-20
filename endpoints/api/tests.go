package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester/unity"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *ApiService) getTests(session *Session, c *gin.Context) {
	var tests []models.Test
	if err := s.db.Preload("TestRuns").Preload("TestConfig").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").Preload("TestConfig.Devices").Find(&tests).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, tests)
}

func (s *ApiService) newTest(session *Session, c *gin.Context) {
	type TestFunc struct {
		Assembly string
		Class    string
		Method   string
	}
	type Request struct {
		Name               string
		TestType           models.TestType
		ExecutionType      models.ExecutionType
		UnityAllTests      bool
		UnitySelectedTests []TestFunc
		AllDevices         bool
		SelectedDevices    []uint
	}

	var request Request
	if err := c.Bind(&request); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	name := strings.TrimSpace(request.Name)
	if len(name) == 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("missing name"))
		return
	}

	switch request.TestType {
	case models.TestTypeUnity:
		if request.UnityAllTests == false && len(request.UnitySelectedTests) == 0 {
			s.error(c, http.StatusBadRequest, fmt.Errorf("missing tests"))
			return
		}
	default:
		s.error(c, http.StatusBadRequest, fmt.Errorf("unsupported Test Type"))
		return
	}

	if request.AllDevices == false && len(request.SelectedDevices) == 0 {
		s.error(c, http.StatusBadRequest, fmt.Errorf("unsupported Test Type"))
		return
	}

	tx := s.db.Begin()
	defer func() {
		tx.Rollback()
	}()
	test := models.Test{
		CompanyID: 1,
		Name:      request.Name,
	}
	if err := tx.Create(&test).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	config := models.TestConfig{
		TestID:        test.ID,
		Type:          request.TestType,
		AllDevices:    request.AllDevices,
		ExecutionType: request.ExecutionType,
	}
	if err := tx.Create(&config).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	switch request.TestType {
	case models.TestTypeUnity:
		unityConfig := models.TestConfigUnity{
			TestConfigID: config.ID,
			RunAllTests:  request.UnityAllTests,
		}
		if err := tx.Create(&unityConfig).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		for _, r := range request.UnitySelectedTests {

			function := models.UnityTestFunction{
				TestConfigUnityID: unityConfig.ID,
				Assembly:          r.Assembly,
				Class:             r.Class,
				Method:            r.Method,
			}
			if err := tx.Create(&function).Error; err != nil {
				s.error(c, http.StatusInternalServerError, err)
				return
			}
		}
	}

	for _, d := range request.SelectedDevices {
		dev := models.TestConfigDevice{
			TestConfigID: config.ID,
			DeviceID:     d,
		}
		if err := tx.Create(&dev).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
	}
	tx.Commit()

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

func extractParams(param string) map[string]string {
	var env map[string]string
	env = make(map[string]string)

	params := strings.Split(param, "\n")
	for _, p := range params {
		kv := strings.Split(p, "=")
		if len(kv) > 1 {
			env[kv[0]] = kv[1]
		}
	}
	return env
}

type RunTestRequest struct {
	AppID  uint
	Params string
}

func (s *ApiService) runTest(c *gin.Context) {
	testId := c.Param("test_id")
	var req RunTestRequest
	if err := c.Bind(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	environmentParams := extractParams(req.Params)

	var app models.App
	if err := s.db.First(&app, req.AppID).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	var devices []models.Device
	if test.TestConfig.AllDevices {
		if err := s.db.Find(&devices).Error; err != nil {
			s.error(c, http.StatusNotFound, err)
			return
		}
	} else {
		if err := s.db.Find(&devices, test.TestConfig.GetDeviceIds()).Error; err != nil {
			s.error(c, http.StatusNotFound, err)
			return
		}
	}

	for i := range devices {
		devices[i].Dev = s.devicesManager.GetDevice(devices[i].DeviceIdentifier)
	}

	var run *models.TestRun
	if test.TestConfig.Type == models.TestTypeUnity {
		tr := unity.New(s.db, s.hostIP, s.devicesManager, s)
		if err := tr.Initialize(test, environmentParams); err != nil {
			s.error(c, http.StatusInternalServerError, err) // Todo status code
			return
		}
		var err error
		run, err = tr.Run(devices, app)
		if err != nil {
			s.error(c, http.StatusInternalServerError, err) // Todo status code
			return
		}
	}

	c.JSON(http.StatusOK, run)
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
	if err := s.db.Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("App").Preload("Test").Preload("Protocols.Performance").Where("test_id = ?", testId).Order("id desc").First(&run).Error; err != nil {
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
	if err := s.db.Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("App").Preload("Test").Preload("Protocols.Performance").First(&run, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, run)
}

func (s *ApiService) getTestRunProtocol(session *Session, c *gin.Context) {
	runId := c.Param("run_id")
	protocolId := c.Param("protocol_id")

	var run models.TestRun
	if err := s.db.Preload("Protocols", "ID = ?", protocolId).Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Protocols.Performance").Preload("Log").Preload("App").Preload("Test").First(&run, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}



	c.JSON(http.StatusOK, run)
}

func (s *ApiService) getData(session *Session, c *gin.Context) {
	name := c.Param("name")
	c.File(fmt.Sprintf("test/data/"+ name))
}
