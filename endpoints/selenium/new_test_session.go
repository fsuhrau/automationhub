package selenium

import (
	"fmt"
	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func (s *SeleniumService) InitNewTestSession(c *gin.Context) {
	type Capabilities struct {
		DeviceID string `json:"device"`
	}
	type Value struct {
		DesiredCapabilities Capabilities
		SessionID           string `json:"webdriver.remote.sessionid"`
		BrowserName         string `json:"browserName"`
		JavascriptEnabled   bool   `json:"javascriptEnabled"`
		Version             string `json:"version"`
		Platform            string `json:"platform"`
		CSSSelectorsEnabled bool   `json:"cssSelectorsEnabled"`
	}

	type Request struct {
		DesiredCapabilities  map[string]interface{} `json:"desiredCapabilities"`
		RequiredCapabilities map[string]interface{} `json:"requiredCapabilities"`
	}

	var req Request
	c.Bind(&req)

	deviceProperties := &device.Properties{}

	mapCapabilities(req.DesiredCapabilities, deviceProperties)
	mapCapabilities(req.RequiredCapabilities, deviceProperties)

	if !deviceProperties.AreValid() {
		s.renderError(c, fmt.Errorf("session can't be created because of missing parameters"))
		return
	}

	appParameter, err := extractAppRequirements(deviceProperties.App, deviceProperties)
	if err != nil {
		s.renderError(c, errors.WithMessage(err, "unable to identify app requirements"))
		return
	}

	session := s.sessionsManager.CreateNewSession(s.logger, deviceProperties, appParameter)

	if err := s.ensureAppIsInstalled(session, deviceProperties, appParameter); err != nil {
		//s.devicesManager.UnlockDevice(session)
		s.renderError(c, err)
		return
	}

	logrus.Infof("Start App on device: %s", session.GetDevice().DeviceName())
	if err := session.GetDevice().StartApp(appParameter, session.GetSessionID(), s.hostIP); err != nil {
		logrus.Errorf("StartApp: %v", err)
		//s.devicesManager.UnlockDevice(session)
		s.renderError(c, errors.WithMessage(err, "could not start app on device"))
		return
	}

	logrus.Infof("Wait for app to be started and connected for device: %s", session.GetDevice().DeviceName())
	if err := session.WaitForConnection(); err != nil {
		logrus.Errorf("StartApp: %v", err)
		//s.devicesManager.UnlockDevice(session)
		s.renderError(c, errors.WithMessage(err, "app not connected"))
		return
	}

	s.sessionsManager.AddSession(session)

	c.JSON(http.StatusOK, &Response{
		SessionID: session.GetSessionID(),
		State:     "Success",
		HCode:     time.Now().UTC().Unix(),
		Value: Value{
			DesiredCapabilities: Capabilities{
				DeviceID: session.GetDevice().DeviceID(),
			},
			SessionID: session.GetSessionID(),
			Platform:  "MAC",
		},
	})
}

func extractAppRequirements(applicationPath string, properties *device.Properties) (*app.Parameter, error) {
	analyser := app.NewAnalyser(applicationPath)
	if err := analyser.AnalyseFile(); err != nil {
		return nil, err
	}
	params := analyser.GetParameter()
	properties.OS = params.Platform
	return params, nil
}

func mapCapabilities(req map[string]interface{}, properties *device.Properties) {
	if v, ok := req["app"]; ok {
		properties.App, ok = v.(string)
	}
	//if v, ok := req["app_id"]; ok {
	//	properties.AppId, ok = v.(string)
	//}
	if v, ok := req["device_name"]; ok {
		properties.Name, ok = v.(string)
	}
	if v, ok := req["device"]; ok {
		properties.DeviceID, ok = v.(string)
	}
	if v, ok := req["device_id"]; ok {
		properties.DeviceID, ok = v.(string)
	}
	if v, ok := req["device_os"]; ok {
		properties.OS, ok = v.(string)
	}
}

func (s *SeleniumService) getDevices(c *gin.Context) {
	type device struct {
		Name string
		OS   string
		ID   string
	}
	var deviceList []device
	devices, _ := s.devicesManager.Devices()
	for _, d := range devices {
		deviceList = append(deviceList, device{
			ID:   d.DeviceID(),
			Name: d.DeviceName(),
			OS:   d.DeviceOSName(),
		})
	}
	c.JSON(200, deviceList)
}

func (s *SeleniumService) ensureAppIsInstalled(session manager.Session, deviceProperties *device.Properties, appParameter *app.Parameter) error {
	retryTimer := 500 * time.Millisecond
	var err error
	for counter := 0; counter < 5; counter++ {
		var isInstalled bool
/*
		if _, err = s.devicesManager.LockDevice(session, deviceProperties); err != nil {
			// wait to unlock
			time.Sleep(retryTimer)
			err = errors.WithMessage(err, "no available device found")
			continue
		}
*/
		if session.GetDevice().DeviceState() != device.StateBooted {
			if err := s.devicesManager.Start(session.GetDevice()); err != nil {
				logrus.Errorf("DeviceState: %v", err)
				// s.devicesManager.UnlockDevice(session)
				time.Sleep(retryTimer)
				continue
			}
		}

		isInstalled, err = session.GetDevice().IsAppInstalled(appParameter)
		if err != nil {
			// s.devicesManager.UnlockDevice(session)
			err = errors.WithMessage(err, "could not check if app is installed on device")
			time.Sleep(retryTimer)
			continue
		}
		if !isInstalled {
			logrus.Infof("InstallApp on device: %s", session.GetDevice().DeviceName())
			if err := session.GetDevice().InstallApp(appParameter); err != nil {
				logrus.Errorf("InstallApp: %v", err)
				// s.devicesManager.UnlockDevice(session)
				err = errors.WithMessage(err, "could not install app on device")
				time.Sleep(retryTimer)
				continue
			}
		} else {
			logrus.Infof("InstallApp is already installed on device: %s", session.GetDevice().DeviceName())
		}

		if true {
			logrus.Infof("Stop App on Device: %s", session.GetDevice().DeviceName())
			if err := session.GetDevice().StopApp(appParameter); err != nil {
				logrus.Errorf("Stop App failed: %v", err)
			}
		}

		/*
		if viper.GetBool("screen_recording") {
			session.Recorder = &Recorder{
				Storage: session.Storage,
				Device:  session.Lock.Device,
			}

			logrus.Infof("Start Recording of session for device: %s", session.Lock.Device.DeviceName())
			if err := session.Recorder.Start(); err != nil {
				logrus.Errorf("start recording failed: %v", err)
			}
		}
		 */
		break
	}

	return err
}

func (s *SeleniumService) StopTestingSession(session *Session, c *gin.Context) {
	type SessionResponse struct {
		SessionID string `json:"sessionId"`
		State     string `json:"state"`
		HCode     int64  `json:"hcode"`
		Status    int64  `json:"status"`
	}
	s.sessionsManager.StopSession(session)
	c.String(http.StatusOK, "")
}
