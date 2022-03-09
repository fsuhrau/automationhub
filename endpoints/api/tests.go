package api

import (
	"fmt"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tester"
	"github.com/fsuhrau/automationhub/tester/scenario"
	"github.com/fsuhrau/automationhub/tester/unity"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func (s *Service) getTests(c *gin.Context) {
	var tests []models.Test
	if err := s.db.Preload("TestRuns").Preload("TestConfig").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").Preload("TestConfig.Devices").Find(&tests).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, tests)
}

func (s *Service) newTest(c *gin.Context) {
	type TestFunc struct {
		Assembly string
		Class    string
		Method   string
	}
	type Request struct {
		Name               string
		TestType           models.TestType
		ExecutionType      models.ExecutionType
		PlatformType       models.PlatformType
		UnityAllTests      bool
		UnitySelectedTests []TestFunc
		AllDevices         bool
		SelectedDevices    []uint
		Categories         []string
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
		Platform:      request.PlatformType,
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
		if len(request.Categories) > 0 {
			unityConfig.Categories = strings.Join(request.Categories, ",")
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

func (s *Service) getTest(c *gin.Context) {
	testId := c.Param("test_id")

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Devices.Device").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, test)
}

func (s *Service) updateTest(c *gin.Context) {
	testId := c.Param("test_id")
	var newTestData models.Test
	c.Bind(&newTestData)

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Devices.Device").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	test.Name = newTestData.Name
	test.TestConfig.ExecutionType = newTestData.TestConfig.ExecutionType
	test.TestConfig.AllDevices = newTestData.TestConfig.AllDevices
	for i := range test.TestConfig.Devices {
		stillNeeded := false
		for d := range newTestData.TestConfig.Devices {
			if test.TestConfig.Devices[i].DeviceID == newTestData.TestConfig.Devices[d].DeviceID {
				stillNeeded = true
				break
			}
		}
		if !stillNeeded {
			s.db.Delete(&test.TestConfig.Devices)
		}
	}
	for i := range newTestData.TestConfig.Devices {
		needCreation := true
		for d := range test.TestConfig.Devices {
			if test.TestConfig.Devices[d].DeviceID == newTestData.TestConfig.Devices[i].DeviceID {
				needCreation = false
			}
		}

		if needCreation {
			newTestData.TestConfig.Devices[i].TestConfigID = test.TestConfig.ID
			s.db.Create(&newTestData.TestConfig.Devices[i])
		}
	}
	test.TestConfig.Devices = newTestData.TestConfig.Devices

	if err := s.db.Save(&test).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
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

func (s *Service) runTest(c *gin.Context) {
	testId := c.Param("test_id")
	var req RunTestRequest
	if err := c.Bind(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	environmentParams := extractParams(req.Params)

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	var app *models.App
	if test.TestConfig.Platform != models.PlatformTypeEditor {
		app = &models.App{}
		if err := s.db.First(app, req.AppID).Error; err != nil {
			s.error(c, http.StatusNotFound, err)
			return
		}
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
	var testRunner tester.Interface

	switch test.TestConfig.Type {
	case models.TestTypeUnity:
		testRunner = unity.New(s.db, s.hostIP, s.devicesManager, s)
		if err := s.db.Preload("UnityTestFunctions").Where("test_config_id = ?", test.TestConfig.ID).First(&test.TestConfig.Unity).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err) // Todo status code
			return
		}
	case models.TestTypeScenario:
		testRunner = scenario.New(s.db, s.hostIP, s.devicesManager, s)
	}

	if err := testRunner.Initialize(test, environmentParams); err != nil {
		s.error(c, http.StatusInternalServerError, err) // Todo status code
		return
	}
	var err error
	run, err = testRunner.Run(devices, app)
	if err != nil {
		s.error(c, http.StatusInternalServerError, err) // Todo status code
		return
	}

	c.JSON(http.StatusOK, run)
}

func (s *Service) getTestRuns(c *gin.Context) {
	testId := c.Param("test_id")

	var testRuns []models.TestRun
	if err := s.db.Preload("Protocols").Find(&testRuns, "test_id = ?", testId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, testRuns)
}

func (s *Service) getLastTestRun(c *gin.Context) {
	type Response struct {
		TestRun   models.TestRun
		PrevRunId uint
		NextRunId uint
	}

	testId := c.Param("test_id")

	var resp Response
	if err := s.db.Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("App").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").Where("test_id = ?", testId).Order("id desc").First(&resp.TestRun).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	// get prev
	s.db.Table("test_runs").Where("test_id = ? and id < ?", testId, resp.TestRun.ID).Order("created_at desc").Limit(1).Select("id").Scan(&resp.PrevRunId)

	// get next
	s.db.Table("test_runs").Where("test_id = ? and id > ?", testId, resp.TestRun.ID).Order("created_at asc").Limit(1).Select("id").Scan(&resp.NextRunId)

	c.JSON(http.StatusOK, resp)
}

func (s *Service) getTestRun(c *gin.Context) {
	type Response struct {
		TestRun   models.TestRun
		PrevRunId uint
		NextRunId uint
	}

	testId := c.Param("test_id")
	runId := c.Param("run_id")

	var resp Response
	if err := s.db.Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("App").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").First(&resp.TestRun, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	// get prev
	s.db.Table("test_runs").Where("test_id = ? and id < ?", testId, runId).Order("created_at desc").Limit(1).Select("id").Scan(&resp.PrevRunId)

	// get next
	s.db.Table("test_runs").Where("test_id = ? and id > ?", testId, runId).Order("created_at asc").Limit(1).Select("id").Scan(&resp.NextRunId)

	c.JSON(http.StatusOK, resp)
}

func (s *Service) getTestRunProtocol(c *gin.Context) {
	runId := c.Param("run_id")
	protocolId := c.Param("protocol_id")

	var run models.TestRun
	if err := s.db.Preload("Protocols", "ID = ?", protocolId).Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Protocols.Performance").Preload("Log").Preload("App").Preload("Test").First(&run, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, run)
}

func (s *Service) getData(c *gin.Context) {
	name := c.Param("name")
	c.File(fmt.Sprintf("test/data/" + name))
}
