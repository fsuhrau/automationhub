package node

import (
	"time"

	"github.com/fsuhrau/automationhub/device"
)

// ScreenshotChunkSize is the number of bytes requested per GetScreenshotData call.
const ScreenshotChunkSize = 512 * 1024

// defaultConnectionTimeout is used when the node can't be asked.
const defaultConnectionTimeout = 5 * time.Minute

// Control forwards a native control call (tap, swipe, text, key, logs).
func (c *RPCClient) Control(req device.ControlRequest) (device.ControlResponse, error) {
	var resp device.ControlResponse
	if err := c.safeCall(rpcService+"Control", &req, &resp); err != nil {
		return resp, controlError(err)
	}
	return resp, nil
}

// StartVideoStream asks the node to stream a device's video to the master.
func (c *RPCClient) StartVideoStream(req device.VideoStreamRequest) error {
	if err := c.safeCall(rpcService+"StartVideoStream", &req, &Void{}); err != nil {
		return controlError(err)
	}
	return nil
}

func (c *RPCClient) IsConnected(deviceId string) bool {
	resp, err := call[BoolResponse](c, "IsConnected", &DeviceRequest{DeviceID: deviceId})
	return err == nil && resp.Value
}

func (c *RPCClient) HasFeature(deviceId string, feature string) bool {
	ok, err := c.forwardBool("HasFeature", &FeatureRequest{DeviceID: deviceId, Feature: feature})
	return err == nil && ok
}

func (c *RPCClient) Execute(deviceId string, data string) {
	c.notify("Execute", &ExecuteRequest{DeviceID: deviceId, Data: data})
}

func (c *RPCClient) ConnectionTimeout(deviceId string) time.Duration {
	resp, err := call[TimeoutResponse](c, "ConnectionTimeout", &DeviceRequest{DeviceID: deviceId})
	if err != nil {
		return defaultConnectionTimeout
	}
	return time.Duration(resp.Timeout)
}

func (c *RPCClient) RunNativeScript(deviceId string, script []byte) {
	c.notify("RunNativeScript", &ExecuteRequest{DeviceID: deviceId, Data: string(script)})
}

func (c *RPCClient) SendAction(deviceId string, script []byte) {
	c.notify("SendAction", &ExecuteRequest{DeviceID: deviceId, Data: string(script)})
}

// GetScreenshot has the node take a screenshot and fetches it in chunks.
func (c *RPCClient) GetScreenshot(deviceId string) ([]byte, int, int, error) {
	meta, err := call[ScreenShotResponse](c, "TakeScreenshot", &DeviceRequest{DeviceID: deviceId})
	if err != nil {
		return nil, 0, 0, err
	}
	data, err := c.screenshotData(&meta)
	if err != nil {
		return nil, 0, 0, err
	}
	return data, int(meta.Width), int(meta.Height), nil
}

func (c *RPCClient) screenshotData(meta *ScreenShotResponse) ([]byte, error) {
	size := int(meta.Size)
	data := make([]byte, 0, size)
	for start := 0; start < size; start += ScreenshotChunkSize {
		chunk, err := call[ScreenShotDataResponse](c, "GetScreenshotData", &ScreenShotDataRequest{
			Hash:  meta.Hash,
			Start: int32(start),
			End:   int32(min(start+ScreenshotChunkSize, size)),
		})
		if err != nil {
			return nil, err
		}
		data = append(data, chunk.Data...)
	}
	return data, nil
}
