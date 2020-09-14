package hub

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fsuhrau/automationhub/app"

	"github.com/fsuhrau/automationhub/device"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"howett.net/plist"
)

var (
	AndroidAPKInfosRegex = regexp.MustCompile(`package: name='(.*)' versionCode='(.*)' versionName='(.*)' compileSdkVersion='(.*)' compileSdkVersionCodename='(.*)'`)
	LaunchActivityRegex  = regexp.MustCompile(`launchable-activity:\s+name='([a-zA-Z0-9.]+)'\s+label='(.*)'\sicon='.*'`)
	recordSession        = false
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func extractAppRequirements(applicationPath string, properties *device.Properties) (*app.Parameter, error) {
	params := &app.Parameter{}
	params.AppPath = applicationPath
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
		cmd := device.NewCommand("aapt", "dump", "badging", applicationPath)
		output, err := cmd.Output()
		if err != nil {
			return params, err
		}
		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			line := scanner.Text()
			if matches := AndroidAPKInfosRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
				params.Identifier = matches[0][1]
				params.Version = matches[0][3]
				params.Additional = fmt.Sprintf("versionCode: %s compileSdkVersion %s", matches[0][2], matches[0][4])
				continue
			}
			if matches := LaunchActivityRegex.FindAllStringSubmatch(line, -1); len(matches) > 0 {
				params.LaunchActivity = matches[0][1]
				params.Name = matches[0][2]
				continue
			}
		}
	}
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

func (s *Server) getDevices(c *gin.Context) {
	type device struct {
		Name string
		OS   string
		ID   string
	}
	var deviceList []device
	devices, _ := s.deviceManager.Devices()
	for _, d := range devices {
		deviceList = append(deviceList, device{
			ID:   d.DeviceID(),
			Name: d.DeviceName(),
			OS:   d.DeviceOSName(),
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

	properties := &device.Properties{}

	mapCapabilities(req.DesiredCapabilities, properties)
	mapCapabilities(req.RequiredCapabilities, properties)

	if !properties.AreValid() {
		s.renderError(c, fmt.Errorf("session can't be created because of missing parameters"))
		return
	}

	appParameter, err := extractAppRequirements(properties.App, properties)
	if err != nil {
		s.renderError(c, errors.WithMessage(err, "unable to identify app requirements"))
	}

	session := s.sessionManager.CreateNewSession(s.logger, properties, appParameter)

	retryTimer := 500 * time.Millisecond

	for counter := 0; counter < 5; counter++ {
		if err = s.deviceManager.LockDevice(session, properties); err != nil {
			// wit to unlock
			time.Sleep(retryTimer)
			err = errors.WithMessage(err, "no available device found")
			continue
		}

		if session.Lock.Device.DeviceState() != device.Booted {
			if err := s.deviceManager.Start(session.Lock.Device); err != nil {
				logrus.Errorf("DeviceState: %v", err)
				s.deviceManager.UnlockDevice(session)
				time.Sleep(retryTimer)
				continue
			}
		}

		isInstalled, err := session.Lock.Device.IsAppInstalled(appParameter)
		if err != nil {
			s.deviceManager.UnlockDevice(session)
			err = errors.WithMessage(err, "could not check if app is installed on device")
			time.Sleep(retryTimer)
			continue
		}
		if !isInstalled {
			logrus.Infof("InstallApp on device: %s", session.Lock.Device.DeviceName())
			if err := session.Lock.Device.InstallApp(appParameter); err != nil {
				logrus.Errorf("InstallApp: %v", err)
				s.deviceManager.UnlockDevice(session)
				err = errors.WithMessage(err, "could not install app on device")
				time.Sleep(retryTimer)
				continue
			}
		} else {
			logrus.Infof("InstallApp is already installed on device: %s", session.Lock.Device.DeviceName())
		}

		logrus.Infof("Stop App on Device: %s", session.Lock.Device.DeviceName())
		if true {
			session.Lock.Device.StopApp(appParameter)
		}

		if recordSession {
			session.Recorder = &Recorder{
				Storage: session.Storage,
				Device:  session.Lock.Device,
			}

			logrus.Infof("Start Recording of session for device: %s", session.Lock.Device.DeviceName())
			if err := session.Recorder.Start(); err != nil {
				logrus.Errorf("start recording failed: %v", err)
			}
		}
		break
	}

	if err != nil {
		s.deviceManager.UnlockDevice(session)
		s.renderError(c, err)
		return
	}

	logrus.Infof("Start App on device: %s", session.Lock.Device.DeviceName())
	if true {
		if err := session.Lock.Device.StartApp(appParameter, session.SessionID, s.hostIP); err != nil {
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

	s.sessionManager.AddSession(session)

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
	s.sessionManager.StopSession(session)
	c.String(http.StatusOK, "")
}
