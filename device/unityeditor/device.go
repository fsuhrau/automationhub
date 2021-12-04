package unityeditor

import (
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/device/generic"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"os/exec"
	"time"

	"github.com/fsuhrau/automationhub/app"

	"github.com/fsuhrau/automationhub/device"
)

const CONNECTION_TIMEOUT = 60 * time.Minute

type Device struct {
	generic.Device
	deviceOSName     string
	deviceOSVersion  string
	deviceName       string
	deviceID         string
	deviceIP         net.IP
	deviceState      device.State
	recordingSession *exec.Cmd
	lastUpdateAt time.Time
	updated      bool

	// Create client
	client      *http.Client
	conn        *websocket.Conn
	sendChannel chan []byte
}

func (d *Device) DeviceOSName() string {
	return d.deviceOSName
}

func (d *Device) DeviceOSVersion() string {
	return d.deviceOSVersion
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
		d.deviceState = device.StateUnknown
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

func (d *Device) StartApp(params *app.Parameter, sessionId string, hostIP net.IP) error {

	type request struct {
		Action    string
		HostIP    string
		SessionID string
	}
	req := request{
		Action:    "start",
		SessionID: sessionId,
		HostIP:    hostIP.String(),
	}
	buffer, _ := json.Marshal(req)
	d.sendChannel <- buffer
	return nil
}

func (d *Device) StopApp(params *app.Parameter) error {
	type request struct {
		Action    string
	}
	req := request{
		Action:    "stop",
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
	return CONNECTION_TIMEOUT
}

func (d *Device) RunNativeScript(script []byte) {

}

func (d *Device) HandleSocketFunction() {
	d.deviceState = device.StateBooted
	d.updated = true
	d.sendChannel = make(chan []byte, 10)

	go func(d *Device) {
		for {
			select {
			case message := <-d.sendChannel:
				err := d.conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					return
				}
			}
		}
	}(d)

	for {
		_, msg, err := d.conn.ReadMessage()
		if err != nil {
			break
		}
		d.lastUpdateAt = time.Now().UTC()
		fmt.Println(string(msg))
	}

	d.deviceState = device.StateRemoteDisconnected
	d.updated = true
}
