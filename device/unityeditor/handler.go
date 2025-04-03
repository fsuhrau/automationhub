package unityeditor

import (
	"context"
	"fmt"
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/fsuhrau/automationhub/hub/node"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/tools/exec"
	sync2 "github.com/fsuhrau/automationhub/utils/sync"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/fsuhrau/automationhub/device"
)

const (
	Manager        = "unity_editor"
	StartupTimeout = 2 * time.Minute
)

var (
	ErrorGetProjectVersion = fmt.Errorf("Could not determinate project project / unity version")
)

type Handler struct {
	devices       map[string]*Device
	deviceStorage storage.Device
	init          bool
	managerCfg    config.Manager
	mu            sync.Mutex
}

func NewHandler(cfg config.Manager, ds storage.Device) *Handler {
	return &Handler{devices: make(map[string]*Device), deviceStorage: ds, managerCfg: cfg}
}

func (m *Handler) Name() string {
	return Manager
}

func (m *Handler) Init(masterUrl, nodeIdentifier string, authToken *string) error {
	m.init = true
	defer func() {
		m.init = false
	}()
	devs, err := m.deviceStorage.GetDevices(Manager)
	if err != nil {
		return err
	}
	for i := range devs {
		deviceId := devs[i].DeviceIdentifier
		dev := &Device{
			deviceOSName:    devs[i].OS,
			deviceOSVersion: devs[i].OSVersion,
			deviceOSInfos:   devs[i].OSInfos,
			unityVersion:    devs[i].TargetVersion,
			deviceName:      devs[i].Name,
			deviceID:        devs[i].DeviceIdentifier,
		}
		dev.SetConfig(devs[i])
		dev.SetLogWriter(generic.NewRemoteLogWriter(masterUrl, nodeIdentifier, dev.deviceID, authToken))
		dev.AddActionHandler(node.NewRemoteActionHandler(masterUrl, nodeIdentifier, dev.deviceID, authToken))
		m.mu.Lock()
		m.devices[deviceId] = dev
		m.mu.Unlock()
	}

	if err := m.RefreshDevices(true); err != nil {
		return err
	}
	return nil
}

func (m *Handler) Start() error {
	return nil
}

func (m *Handler) Stop() error {
	return nil
}

func (m *Handler) StartDevice(deviceID string) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer func() {
		cancelFunc()
	}()

	wg := sync2.NewExtendedWaitGroup(ctx)
	m.mu.Lock()
	dev, ok := m.devices[deviceID]
	m.mu.Unlock()

	if ok {
		config := dev.GetConfig()
		var projectDir string
		if config != nil {
			projectDir = config.GetAttribute("projectDir")
		}
		instanceFile := filepath.Join(projectDir, ".hub")
		_ = os.Remove(instanceFile)

		projectVersion, err := GetUsedUnityVersion(projectDir)
		if err != nil {
			return err
		}

		unityEditorPath := GetUnityEditorPath(m.managerCfg.UnityPath, projectVersion)

		// Generate a random log file name
		logFileName := fmt.Sprintf("automation_hub_unity_%d.log", rand.Int())
		dev.instanceLogFile = filepath.Join(projectDir, logFileName)

		unityParams := []string{
			"-buildTarget",
			m.managerCfg.UnityBuildTarget,
			"-projectPath", projectDir, "-overrideProfile", "automation", "-executeMethod", "AutomationLoader.LoadSceneAndConnect", "-logFile", dev.instanceLogFile, "-debugCodeOptimization", "--ump-channel-service-on-startup",
		}
		dev.process = exec.NewCommand(unityEditorPath, unityParams...)
		dev.startedAt = time.Now()
		if err := dev.process.Start(); err != nil {
			return err
		}

		dev.UnityLogStartListening()

		m.mu.Lock()
		m.devices[deviceID] = dev
		m.mu.Unlock()
		waitUntil := dev.startedAt.Add(StartupTimeout)

		wg.Add(1)

		var functionError error

		go func() {
			triggerTime := 5 * time.Second

			defer wg.Done()
			timer := time.NewTimer(triggerTime)
			defer timer.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					if _, err := os.Stat(instanceFile); os.IsNotExist(err) {
						if time.Now().After(waitUntil) {
							functionError = fmt.Errorf("device didn't start in time")
							return
						}
						timer.Reset(triggerTime)
						continue
					}
				}
				break
			}
			timer.Reset(triggerTime)
			for {
				select {
				case <-ctx.Done():
					return
				case <-timer.C:
					m.mu.Lock()
					if m.devices[deviceID].DeviceState() == device.StateBooted {
						m.mu.Unlock()
						break
					}
					m.mu.Unlock()

					if time.Now().After(waitUntil) {
						functionError = fmt.Errorf("device didn't start in time")
						return
					}
					timer.Reset(triggerTime)
					continue
				}
				break
			}
		}()

		err = wg.WaitUntil(waitUntil)

		if functionError != nil {
			return functionError
		}

		return err
	}

	return device.DeviceNotFoundError
}

func GetUnityEditorPath(editorPath string, projectVersion string) string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(editorPath, projectVersion, "Unity.app/Contents/MacOS/Unity")
	}
	return ""
}

func GetUsedUnityVersion(projectDir string) (string, error) {
	projectVersionFile := filepath.Join(projectDir, "ProjectSettings", "ProjectVersion.txt")
	data, err := os.ReadFile(projectVersionFile)
	if err != nil {
		return "", ErrorGetProjectVersion
	}

	var projectVersion string
	re := regexp.MustCompile(`m_EditorVersion:\s*([0-9]+\.[0-9]+\.[0-9]+[a-z0-9]*)`)
	match := re.FindStringSubmatch(string(data))
	if len(match) > 1 {
		projectVersion = match[1]
	} else {
		return "", ErrorGetProjectVersion
	}
	return projectVersion, nil
}

func (m *Handler) StopDevice(deviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if dev, ok := m.devices[deviceID]; ok {
		if dev.process != nil && dev.process.Process != nil {
			if err := dev.process.Process.Kill(); err != nil {
				return err
			}
			dev.process = nil
		}
		dev.UnityLogStopListening()
	}
	return nil
}

func (m *Handler) GetDevices() ([]device.Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	devices := make([]device.Device, 0, len(m.devices))
	for _, d := range m.devices {
		devices = append(devices, d)
	}
	return devices, nil
}

func (m *Handler) RefreshDevices(force bool) error {
	now := time.Now().UTC()

	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.devices {
		if now.Sub(m.devices[i].lastUpdateAt) > 1*time.Minute {
			if m.devices[i].deviceState != device.StateShutdown {
				m.devices[i].deviceState = device.StateShutdown
				m.devices[i].updated = true
			}
		}
		if m.devices[i].updated {
			m.deviceStorage.Update(m.Name(), m.devices[i])
			m.devices[i].updated = false
		}
	}
	return nil
}

func (m *Handler) HasDevice(dev device.Device) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, device := range m.devices {
		if device == dev {
			return true
		}
	}
	return false
}

func (m *Handler) RegisterDevice(data device.RegisterData) (device.Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Printf("Register device: %v\n", data)

	lastUpdate := time.Now().UTC()

	if d, ok := m.devices[data.DeviceID]; ok {
		d.deviceOSName = data.DeviceOS
		d.deviceOSVersion = data.DeviceOSVersion
		d.unityVersion = data.TargetVersion
		d.deviceOSInfos = data.DeviceOSInfos
		d.deviceName = fmt.Sprintf("Unity on %s", data.Name)
		d.deviceIP = data.DeviceIP
		d.lastUpdateAt = lastUpdate
		d.managerConnection = data.Conn
		d.updated = true
		m.devices[data.DeviceID] = d
		m.deviceStorage.Update(m.Name(), d)

	} else {
		d = &Device{
			deviceName:        fmt.Sprintf("Unity on %s", data.Name),
			deviceID:          data.DeviceID,
			deviceOSName:      data.DeviceOS,
			deviceOSVersion:   data.DeviceOSVersion,
			unityVersion:      data.TargetVersion,
			deviceOSInfos:     data.DeviceOSInfos,
			deviceIP:          data.DeviceIP,
			lastUpdateAt:      lastUpdate,
			managerConnection: data.Conn,
			updated:           true,
		}
		dev := models.Device{
			DeviceIdentifier: data.DeviceID,
			DeviceType:       models.DeviceTypeUnityEditor,
			Name:             fmt.Sprintf("Unity on %s", data.Name),
			Manager:          Manager,
			OS:               data.DeviceOS,
			OSVersion:        data.DeviceOSVersion,
			OSInfos:          data.DeviceOSInfos,
			TargetVersion:    data.TargetVersion,
			ConnectionParameter: &models.ConnectionParameter{
				ConnectionType: models.ConnectionTypeRemote,
			},
			CustomParameter: []models.CustomParameter{
				{
					Key:   "projectDir",
					Value: data.ProjectDir,
				},
			},
		}
		d.SetConfig(&dev)
		m.deviceStorage.NewDevice(m.Name(), dev)
		m.devices[data.DeviceID] = d
	}
	go m.devices[data.DeviceID].HandleManagerConnection()
	return m.devices[data.DeviceID], nil
}
