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

const (
	AVG_HISTORY_LIMIT = 20
)

func (s *Service) getTests(c *gin.Context, project *models.Project, application *models.App) {
	var tests []models.Test
	if err := s.db.Where("app_id = ?", application.ID).Preload("TestRuns").Preload("TestConfig").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").Preload("TestConfig.Devices").Find(&tests).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, tests)
}

func (s *Service) newTest(c *gin.Context, project *models.Project, application *models.App) {

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
		AppID: application.ID,
		Name:  request.Name,
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

func (s *Service) getTest(c *gin.Context, project *models.Project, application *models.App) {
	testId := c.Param("test_id")

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Devices.Device").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, test)
}

func (s *Service) updateTest(c *gin.Context, project *models.Project, application *models.App) {
	testId := c.Param("test_id")

	type request struct {
		Name          string
		ExecutionType models.ExecutionType
		AllDevices    bool
		Devices       []uint
		RunAllTests   bool
		Categories    string
		TestFunctions []models.UnityTestFunction
	}

	var req request
	c.Bind(&req)

	var test models.Test
	if err := s.db.Preload("TestConfig").Preload("TestConfig.Devices").Preload("TestConfig.Devices.Device").Preload("TestConfig.Unity").Preload("TestConfig.Unity.UnityTestFunctions").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	test.Name = req.Name
	test.TestConfig.ExecutionType = req.ExecutionType
	test.TestConfig.AllDevices = req.AllDevices

	if req.AllDevices {
		// since all devices are selected we don't need to specify them
		s.db.Where("test_config_id = ?", test.TestConfig.ID).Delete(&models.TestConfigDevice{})
		test.TestConfig.Devices = []models.TestConfigDevice{}
	} else {
		// remove devices that are not needed anymore
		for i := range test.TestConfig.Devices {

			stillNeeded := false
			for _, d := range req.Devices {
				if test.TestConfig.Devices[i].DeviceID == d {
					stillNeeded = true
					break
				}
			}

			if !stillNeeded {
				if err := s.db.Delete(&test.TestConfig.Devices[i]).Error; err != nil {
					s.error(c, http.StatusBadRequest, err)
					return
				}
			}
		}

		for _, i := range req.Devices {
			needCreation := true
			for d := range test.TestConfig.Devices {
				if test.TestConfig.Devices[d].DeviceID == i {
					needCreation = false
				}
			}

			if needCreation {
				newDevice := models.TestConfigDevice{
					TestConfigID: test.TestConfig.ID,
					DeviceID:     i,
				}
				if err := s.db.Create(&newDevice).Error; err != nil {
					s.error(c, http.StatusBadRequest, err)
					return
				}
				test.TestConfig.Devices = append(test.TestConfig.Devices, newDevice)
			}
		}
	}

	if test.TestConfig.Type == models.TestTypeUnity {
		test.TestConfig.Unity.RunAllTests = req.RunAllTests
		test.TestConfig.Unity.Categories = req.Categories
		if err := s.db.Save(&test.TestConfig.Unity).Error; err != nil {
			s.error(c, http.StatusBadRequest, err)
			return
		}
	}

	if err := s.db.Save(&test.TestConfig).Error; err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

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
	AppBinaryID int
	Params      string
}

func (s *Service) runTest(c *gin.Context, project *models.Project, application *models.App) {

	testId := c.Param("test_id")
	var req RunTestRequest
	if err := c.Bind(&req); err != nil {
		s.error(c, http.StatusBadRequest, err)
		return
	}

	environmentParams := extractParams(req.Params)

	var test models.Test
	if err := s.db.Preload("App").Preload("TestConfig").Preload("TestConfig.Devices").First(&test, testId).Error; err != nil {
		s.error(c, http.StatusNotFound, err)
		return
	}

	var binary *models.AppBinary
	if test.App.Platform != models.PlatformTypeEditor {
		binary = &models.AppBinary{}
		if err := s.db.Preload("App").First(binary, req.AppBinaryID).Error; err != nil {
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
	run, err = testRunner.Run(devices, binary)
	if err != nil {
		s.error(c, http.StatusInternalServerError, err) // Todo status code
		return
	}

	c.JSON(http.StatusOK, run)
}

func (s *Service) getTestRuns(c *gin.Context, project *models.Project, application *models.App) {
	testId := c.Param("test_id")

	var testRuns []models.TestRun
	if err := s.db.Preload("Protocols").Find(&testRuns, "test_id = ?", testId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, testRuns)
}

func (s *Service) getLastTestRun(c *gin.Context, project *models.Project, application *models.App) {
	type Response struct {
		TestRun   models.TestRun
		PrevRunId uint
		NextRunId uint
	}

	testId := c.Param("test_id")

	var resp Response
	if err := s.db.Preload("DeviceStatus").Preload("DeviceStatus.Device").Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("AppBinary").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").Where("test_id = ?", testId).Order("id desc").First(&resp.TestRun).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	// get prev
	s.db.Table("test_runs").Where("test_id = ? and id < ?", testId, resp.TestRun.ID).Order("created_at desc").Limit(1).Select("id").Scan(&resp.PrevRunId)

	// get next
	s.db.Table("test_runs").Where("test_id = ? and id > ?", testId, resp.TestRun.ID).Order("created_at asc").Limit(1).Select("id").Scan(&resp.NextRunId)

	c.JSON(http.StatusOK, resp)
}

func (s *Service) getTestRun(c *gin.Context, project *models.Project, application *models.App) {
	type Response struct {
		TestRun   models.TestRun
		PrevRunId uint
		NextRunId uint
	}

	testId := c.Param("test_id")
	runId := c.Param("run_id")

	var resp Response
	if err := s.db.Preload("DeviceStatus").Preload("DeviceStatus.Device").Preload("Protocols").Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Log").Preload("AppBinary").Preload("Test").Preload("Test.TestConfig").Preload("Protocols.Performance").First(&resp.TestRun, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	for i := range resp.TestRun.DeviceStatus {
		var histStatus []models.TestRunDeviceStatus
		if err := s.db.Where("device_id = ? and test_run_id in (select tr.id from test_runs tr where tr.id < ? order by tr.id desc limit ?)", resp.TestRun.DeviceStatus[i].DeviceID, resp.TestRun.DeviceStatus[i].TestRunID, AVG_HISTORY_LIMIT).Find(&histStatus).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		for _, t := range histStatus {
			resp.TestRun.DeviceStatus[i].HistAvgStartupTime += t.StartupTime
		}
		resp.TestRun.DeviceStatus[i].HistAvgStartupTime = resp.TestRun.DeviceStatus[i].HistAvgStartupTime / uint(len(histStatus))
	}

	// get prev
	s.db.Table("test_runs").Where("test_id = ? and id < ?", testId, runId).Order("created_at desc").Limit(1).Select("id").Scan(&resp.PrevRunId)

	// get next
	s.db.Table("test_runs").Where("test_id = ? and id > ?", testId, runId).Order("created_at asc").Limit(1).Select("id").Scan(&resp.NextRunId)

	c.JSON(http.StatusOK, resp)
}

func (s *Service) getTestRunProtocol(c *gin.Context, project *models.Project, application *models.App) {
	runId := c.Param("run_id")
	protocolId := c.Param("protocol_id")

	var run models.TestRun
	if err := s.db.Preload("Protocols", "ID = ?", protocolId).Preload("Protocols.Device").Preload("Protocols.Entries").Preload("Protocols.Performance").Preload("Log").Preload("AppBinary").Preload("Test").First(&run, runId).Error; err != nil {
		s.error(c, http.StatusInternalServerError, err)
		return
	}

	if len(run.Protocols) > 0 {
		var histProtocols []models.TestProtocol
		if err := s.db.Where("device_id = ? and test_name = ?", run.Protocols[0].DeviceID, run.Protocols[0].TestName).Order("id desc").Limit(AVG_HISTORY_LIMIT).Preload("Performance").Find(&histProtocols).Error; err != nil {
			s.error(c, http.StatusInternalServerError, err)
			return
		}
		for _, p := range histProtocols {
			run.Protocols[0].HistAvgFPS += p.AvgFPS
			run.Protocols[0].HistAvgMEM += p.AvgMEM
			run.Protocols[0].HistAvgCPU += p.AvgCPU
		}
		run.Protocols[0].HistAvgFPS = run.Protocols[0].HistAvgFPS / float32(len(histProtocols))
		run.Protocols[0].HistAvgMEM = run.Protocols[0].HistAvgMEM / float32(len(histProtocols))
		run.Protocols[0].HistAvgCPU = run.Protocols[0].HistAvgCPU / float32(len(histProtocols))
		run.Protocols[0].TestProtocolHistory = histProtocols
	}

	c.JSON(http.StatusOK, run)
}

func (s *Service) getData(c *gin.Context) {
	name := c.Param("name")
	c.File(fmt.Sprintf("test/data/" + name))
}
