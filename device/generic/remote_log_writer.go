package generic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/device"
	"net/http"
	"time"
)

type RemoteLogWriter struct {
	masterURL string
	node      string
	deviceId  string
}

func (w *RemoteLogWriter) Device() interface{} {
	return nil
}

func (w *RemoteLogWriter) Parent() device.LogWriter {
	return nil
}

type LogType int32

const (
	LogType_Performance LogType = 0
	LogType_Data        LogType = 1
	LogType_Log         LogType = 2
	LogType_Error       LogType = 3
)

type request struct {
	Node        string  `json:"node"`
	DeviceID    string  `json:"device_id"`
	Source      string  `json:"source"`
	Type        LogType `json:"type"`
	Message     string  `json:"message"`
	Checkpoint  string  `json:"checkpoint"`
	Cpu         float64 `json:"cpu"`
	Fps         float64 `json:"fps"`
	Mem         float64 `json:"mem"`
	VertexCount float64 `json:"vertex_count"`
	Triangles   float64 `json:"triangles"`
	Other       string  `json:"other"`
}

func NewRemoteLogWriter(masterURL, nodeIdentifier, deviceId string) *RemoteLogWriter {
	return &RemoteLogWriter{masterURL: masterURL, node: nodeIdentifier, deviceId: deviceId}
}

func (w *RemoteLogWriter) sendLog(log *request) error {
	log.DeviceID = w.deviceId
	log.Node = w.node

	logData, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log data: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/node/log", w.masterURL), bytes.NewBuffer(logData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send log data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response: %v", resp.Status)
	}

	return nil
}

func (w *RemoteLogWriter) Log(source, format string, params ...interface{}) {
	log := &request{
		Source:  source,
		Message: fmt.Sprintf(format, params...),
		Type:    LogType_Log,
	}
	if err := w.sendLog(log); err != nil {
		fmt.Printf("failed to send log: %v\n", err)
	}
}

func (w *RemoteLogWriter) Error(source, format string, params ...interface{}) {
	log := &request{
		Source:  source,
		Message: fmt.Sprintf(format, params...),
		Type:    LogType_Error,
	}
	if err := w.sendLog(log); err != nil {
		fmt.Printf("failed to send log: %v\n", err)
	}
}

func (w *RemoteLogWriter) LogPerformance(checkpoint string, cpu, fps, mem, vertexCount, triangles float64, other string) {
	log := &request{
		Source:      "performance",
		Checkpoint:  checkpoint,
		Cpu:         cpu,
		Fps:         fps,
		Mem:         mem,
		VertexCount: vertexCount,
		Triangles:   triangles,
		Other:       other,
		Type:        LogType_Performance,
	}
	if err := w.sendLog(log); err != nil {
		fmt.Printf("failed to send log: %v\n", err)
	}
}

func (w *RemoteLogWriter) Data(source, path string) {
	log := &request{
		Source:  source,
		Message: fmt.Sprintf("Data path: %s", path),
		Type:    LogType_Data,
	}
	if err := w.sendLog(log); err != nil {
		fmt.Printf("failed to send log: %v\n", err)
	}
}

func (w *RemoteLogWriter) TestProtocolId() *uint {
	return nil
}
