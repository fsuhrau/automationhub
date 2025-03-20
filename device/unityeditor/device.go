package unityeditor

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/app"

	"github.com/fsuhrau/automationhub/device"
)

const ConnectionTimeout = 60 * time.Minute

type Device struct {
	generic.Device
	deviceOSName     string
	deviceOSVersion  string
	deviceOSInfos    string
	unityVersion     string
	deviceName       string
	deviceID         string
	deviceIP         net.IP
	deviceState      device.State
	recordingSession *exec.Cmd
	lastUpdateAt     time.Time
	startedAt        time.Time
	updated          bool

	// Create client
	client            *http.Client
	managerConnection *websocket.Conn
	sendChannel       chan []byte

	ctx             context.Context
	cancel          context.CancelFunc
	process         *exec.Cmd
	instanceLogFile string
}

func (d *Device) DeviceModel() string {
	return ""
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
}

func (d *Device) DeviceOSVersion() string {
	return d.deviceOSVersion
}

func (d *Device) TargetVersion() string {
	return d.unityVersion
}

func (d *Device) DeviceName() string {
	return d.deviceName
}

func (d *Device) DeviceID() string {
	return d.deviceID
}

func (d *Device) DeviceIP() net.IP {
	return d.deviceIP
}

func (d *Device) DeviceState() device.State {
	return d.deviceState
}

func (d *Device) SetDeviceState(state string) {
	switch state {
	case "StateBooted":
		d.deviceState = device.StateBooted
	case "StateShutdown":
		d.deviceState = device.StateShutdown
	case "StateRemoteDisconnected":
		d.deviceState = device.StateRemoteDisconnected
	default:
		d.deviceState = device.StateShutdown
	}
}

func (d *Device) UpdateDeviceInfos() error {
	return nil
}

func (d *Device) IsAppInstalled(params *app.Parameter) (bool, error) {
	return true, nil
}

func (d *Device) InstallApp(params *app.Parameter) error {
	return nil
}

func (d *Device) UninstallApp(params *app.Parameter) error {
	return nil
}

func (d *Device) getHTTPClient() *http.Client {
	if d.client == nil {
		d.client = &http.Client{}
	}
	return d.client
}

func (d *Device) StartApp(params *app.Parameter, sessionId string, nodeUrl string) error {

	type request struct {
		Action    string
		HostUrl   string
		SessionID string
	}
	req := request{
		Action:    "start",
		SessionID: sessionId,
		HostUrl:   nodeUrl,
	}
	buffer, _ := json.Marshal(req)
	d.sendChannel <- buffer
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	type request struct {
		Action string
	}
	req := request{
		Action: "stop",
	}
	buffer, _ := json.Marshal(req)
	d.sendChannel <- buffer
	return nil
}

func (d *Device) IsAppConnected() bool {
	return d.Connection() != nil
}

func (d *Device) StartRecording(path string) error {
	return nil
}

func (d *Device) StopRecording() error {
	var err error
	return err
}

func (d *Device) GetScreenshot() ([]byte, int, int, error) {
	return nil, 0, 0, nil
}

func (d *Device) HasFeature(string) bool {
	return false
}

func (d *Device) Execute(string) {

}

func (d *Device) ConnectionTimeout() time.Duration {
	return ConnectionTimeout
}

func (d *Device) RunNativeScript(script []byte) {

}

func (d *Device) HandleManagerConnection() {
	d.deviceState = device.StateBooted
	d.updated = true
	d.sendChannel = make(chan []byte, 10)

	go func(d *Device) {
		for {
			select {
			case message := <-d.sendChannel:
				err := d.managerConnection.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					return
				}
			}
		}
	}(d)

	for {
		_, msg, err := d.managerConnection.ReadMessage()
		if err != nil {
			break
		}
		d.lastUpdateAt = time.Now().UTC()
		_ = msg
	}

	d.deviceState = device.StateShutdown
	d.updated = true
}

func (d *Device) UnityLogStartListening() {
	d.ctx, d.cancel = context.WithCancel(context.Background())
	go d.processUnityLog()
}

func (d *Device) processUnityLog() {
	time.Sleep(1 * time.Second)

	file, err := os.Open(d.instanceLogFile)
	if err != nil {
		return
	}
	defer func() {
		file.Close()
	}()

	reader := bufio.NewReader(file)
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return
			}
			if len(line) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
	}
}

func (d *Device) UnityLogStopListening() {
	if d.cancel != nil {
		d.cancel()
	}
}
