package hub

import (
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsuhrau/automationhub/devices"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"howett.net/plist"
)

var recordSession = false

type BinaryParameter struct {
	Identifier string
	Version    string
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func extractAppRequirements(applicationPath string, properties *DeviceProperties) (BinaryParameter, error) {
	var params BinaryParameter
	extension := filepath.Ext(applicationPath)
	if extension == ".app" {
		// apple app
		iosAppInfoPlist := filepath.Join(applicationPath, "Info.plist")

		if fileExists(iosAppInfoPlist) {
			plistData, err := ioutil.ReadFile(iosAppInfoPlist)
			if err != nil {
				return params, err
			}

			plistContent := map[string]interface{}{}
			_, err = plist.Unmarshal(plistData, &plistContent)
			if err != nil {
				return params, err
			}

			if val, ok := plistContent["DTPlatformName"]; ok {
				properties.OS = val.(string)
			}
			if val, ok := plistContent["CFBundleVersion"]; ok {
				params.Version = val.(string)
			}
			if val, ok := plistContent["CFBundleIdentifier"]; ok {
				params.Identifier = val.(string)
			}
		}
		macAppInfoPlist := filepath.Join(applicationPath, "Contents/Info.plist")
		if fileExists(macAppInfoPlist) {
			plistData, err := ioutil.ReadFile(macAppInfoPlist)
			if err != nil {
				return params, err
			}

			plistContent := map[string]interface{}{}
			_, err = plist.Unmarshal(plistData, &plistContent)
			if err != nil {
				return params, err
			}

			if val, ok := plistContent["DTSDKName"]; ok {
				properties.OS = val.(string)
			}
			if val, ok := plistContent["CFBundleVersion"]; ok {
				params.Version = val.(string)
			}
			if val, ok := plistContent["CFBundleIdentifier"]; ok {
				params.Identifier = val.(string)
			}
		}

	} else if extension == ".apk" {
		// android app
		properties.OS = "android"
		// cmd.Execute("aapt", "dump", "badging", applicationPath)

	}
	return params, nil
}

func mapCapabilities(req map[string]interface{}, properties *DeviceProperties) {
	if v, ok := req["app"]; ok {
		properties.App = v.(string)
	}
	if v, ok := req["app_id"]; ok {
		properties.AppId = v.(string)
	}
	if v, ok := req["device_name"]; ok {
		properties.Name = v.(string)
	}
	if v, ok := req["device_os"]; ok {
		properties.OS = v.(string)
	}
}

func (s *Server) getSessions(c *gin.Context) {
	c.JSON(200, s.sessions)
}

func (s *Server) getDevices(c *gin.Context) {
	type device struct {
		Name string
		OS string
		ID string
	}
	var deviceList []device
	devices, _ := s.deviceManager.Devices()
	for _, d := range devices {
		deviceList = append(deviceList, device{
			ID: d.DeviceID(),
			Name: d.DeviceName(),
			OS: d.DeviceOSName(),
		})
	}
	c.JSON(200, deviceList)
}

func (s *Server) InitNewTestSession(c *gin.Context) {
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

	properties := &DeviceProperties{}

	mapCapabilities(req.DesiredCapabilities, properties)

	deviceType := 3
	//var properties *DeviceProperties
	switch deviceType {
	case 1:
		properties = &DeviceProperties{
			Name:  "MacOS",
			Type: "macos_app",
			App:   "/Users/fabian.suhrau/projects/game_foe_mobile/proj.ios/build/Debug/foe_mac_develop.app",
			AppId: "com.innogames.enterprise.iforge",
		}
		break
	case 2:
		properties = &DeviceProperties{
			Name: "iPad (5th generation)",
			Type: "ios_simulator",
			OS:   "iOS-13-5",
			App:   "/Users/fabian.suhrau/projects/game_foe_mobile1/proj.ios/build/Debug/foe_mobile_develop.app",
			AppId: "com.innogames.enterprise.iforge",
		}
		break
	case 3:
		//properties = &DeviceProperties{
		//	Name:  "iPad von Fabian Suhrau",
		//	Type: "ios_device",
		//	App:   "/Users/fabian.suhrau/projects/game_foe_mobile/proj.ios/build/Debug/foe_mobile_develop.app",
		//	AppId: "com.innogames.enterprise.iforge",
		//}
		properties = &DeviceProperties{
			Name:  "[FoE]06237",
			Type: "ios_device",
			OS: "iphoneos",
			App:   "/Users/fabian.suhrau/projects/game_foe_mobile1/proj.ios/build/Debug/foe_mobile_develop.app",
			AppId: "com.innogames.enterprise.iforge",
		}
		break
	case 4:
		properties = &DeviceProperties{
			// Name:  "shieldtablet",
			Name:  "Tinker_Board_S",
			Type: "android_device",
			App:   "/Users/fabian.suhrau/projects/game_foe_mobile/proj.android/app/build/outputs/apk/googleV7aUnity/debug/app-google-v7a-unity-debug.apk",
			AppId: "com.innogames.foeandroid",
		}
		break
	}

	params, err := extractAppRequirements(properties.App, properties)
	if err != nil {
		s.renderError(c, errors.WithMessage(err, "unable to identify app requirements"))
	}

	_ = params

	session := createNewSession(s.logger, properties)

	err = s.deviceManager.LockDevice(session, properties)

	if err != nil {
		s.renderError(c, errors.WithMessage(err, "no available device found"))
		return
	}

	if session.Lock.Device.DeviceState() != devices.Booted {
		if err := s.deviceManager.Start(session.Lock.Device); err != nil {
			logrus.Errorf("DeviceState: %v", err)
			s.deviceManager.UnlockDevice(session)
			s.renderError(c, errors.WithMessage(err, "could not start device"))
			return
		}
	}

	if !session.Lock.Device.IsAppInstalled(properties.AppId) {
		logrus.Infof("InstallApp on device: %s", session.Lock.Device.DeviceName())
		if err := session.Lock.Device.InstallApp(properties.App); err != nil {
			logrus.Errorf("InstallApp: %v", err)
			s.deviceManager.UnlockDevice(session)
			s.renderError(c, errors.WithMessage(err, "could not install app on device"))
			return
		}
	} else {
		logrus.Infof("InstallApp is already installed on device: %s", session.Lock.Device.DeviceName())
	}

	if err := session.Lock.Device.ExtractAppParameters(properties.App); err != nil {
		s.deviceManager.UnlockDevice(session)
		s.renderError(c, errors.WithMessage(err, "could not identify app parameter"))
		return
	}

	logrus.Infof("Stop App on Device: %s", session.Lock.Device.DeviceName())
	if !viper.GetBool("debugger") {
		session.Lock.Device.StopApp(properties.App, properties.AppId)
	}

	if recordSession {
		session.Recorder = &Recorder{
			Storage: session.Storage,
			Device: session.Lock.Device,
		}

		logrus.Infof("Start Recording of session for device: %s", session.Lock.Device.DeviceName())
		if err := session.Recorder.Start(); err != nil {
			logrus.Errorf("start recording failed: %v", err)
		}
	}

	logrus.Infof("Start App on device: %s", session.Lock.Device.DeviceName())
	if !viper.GetBool("debugger") {
		if err := session.Lock.Device.StartApp(properties.App, properties.AppId, session.SessionID, s.hostIP); err != nil {
			logrus.Errorf("StartApp: %v", err)
			s.deviceManager.UnlockDevice(session)
			s.renderError(c, errors.WithMessage(err, "could not start app on device"))
			return
		}
	}

	logrus.Infof("Wait for app to be started and connected for device: %s", session.Lock.Device.DeviceName())
	if err := session.WaitForConnection(); err != nil {
		logrus.Errorf("StartApp: %v", err)
		s.deviceManager.UnlockDevice(session)
		s.renderError(c, errors.WithMessage(err, "app not connected"))
		return
	}

	// s.deviceManager.Send(session, fmt.Sprintf("event:SetSession///%s", session.SessionID))

	s.sessions[session.SessionID] = session

	c.JSON(http.StatusOK, &ServerResponse{
		SessionID: session.SessionID,
		State:     "Success",
		HCode:     time.Now().UTC().Unix(),
		Value: Value{
			DesiredCapabilities: Capabilities{
				DeviceID: session.Lock.Device.DeviceID(),
			},
			SessionID: session.SessionID,
			Platform:  "MAC",
		},
	})
}

func (s *Server) StopTestingSession(session *Session, c *gin.Context) {
	type SessionResponse struct {
		SessionID string `json:"sessionId"`
		State     string `json:"state"`
		HCode     int64  `json:"hcode"`
		Status    int64  `json:"status"`
	}

	if session.Recorder != nil {
		if err := session.Recorder.Stop(); err != nil {
			logrus.Errorf("stop recording session failed: %v", err)
		}
	}

	s.deviceManager.UnlockDevice(session)
	session.Storage.Close()
	delete(s.sessions, session.SessionID)
	c.String(http.StatusOK, "")
}
