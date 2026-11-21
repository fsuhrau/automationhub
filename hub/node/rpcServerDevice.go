package node

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/videostream"
	"github.com/sirupsen/logrus"
)

// Control applies a native control call to a local device.
func (s *RPCNode) Control(req *device.ControlRequest, resp *device.ControlResponse) error {
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		result, err := device.ApplyControl(dev, *req)
		*resp = result
		return err
	})
}

// StartVideoStream starts the video of a local device and pumps it into a
// websocket to the master's /node/video/:stream_id endpoint.
func (s *RPCNode) StartVideoStream(req *device.VideoStreamRequest, resp *Void) error {
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		streamer, ok := dev.(device.VideoStreamer)
		if !ok {
			return device.ErrUnsupported
		}
		stream, err := streamer.StartVideo(context.Background(), req.Options)
		if err != nil {
			return err
		}
		master := videostream.MasterTarget{Host: s.config.MasterURL, AuthToken: s.authToken()}
		return videostream.ServeToMaster(master, req.StreamID, stream)
	})
}

func (s *RPCNode) IsConnected(req *DeviceRequest, resp *BoolResponse) error {
	logrus.Info("RPC: IsConnected")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		resp.Value = dev.IsAppConnected()
		return nil
	})
}

func (s *RPCNode) TakeScreenshot(req *DeviceRequest, resp *ScreenShotResponse) error {
	logrus.Info("RPC: TakeScreenshot")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		b, w, h, err := dev.GetScreenshot()
		if err != nil {
			return err
		}
		resp.Hash = req.DeviceID
		resp.Size = int32(len(b))
		resp.Width = int32(w)
		resp.Height = int32(h)

		s.screenshotMutex.Lock()
		s.screenshots[req.DeviceID] = b
		s.screenshotMutex.Unlock()
		return nil
	})
}

// GetScreenshotData returns a slice of the last screenshot taken for the
// device. The buffer is released once its last chunk was requested.
func (s *RPCNode) GetScreenshotData(req *ScreenShotDataRequest, resp *ScreenShotDataResponse) error {
	s.screenshotMutex.Lock()
	defer s.screenshotMutex.Unlock()
	data, ok := s.screenshots[req.Hash]
	if !ok {
		return fmt.Errorf("no screenshot data for %s", req.Hash)
	}
	start, end := int(req.Start), int(req.End)
	if start < 0 || end < start || start > len(data) {
		return fmt.Errorf("invalid screenshot range %d-%d (size %d)", start, end, len(data))
	}
	end = min(end, len(data))
	resp.Data = data[start:end]
	if end == len(data) {
		delete(s.screenshots, req.Hash)
	}
	return nil
}

func (s *RPCNode) HasFeature(req *FeatureRequest, resp *BoolResponse) error {
	logrus.Info("RPC: HasFeature")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		resp.Value = dev.HasFeature(req.Feature)
		return nil
	})
}

func (s *RPCNode) Execute(req *ExecuteRequest, resp *Void) error {
	logrus.Info("RPC: Execute")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		dev.Execute(req.Data)
		return nil
	})
}

func (s *RPCNode) ConnectionTimeout(req *DeviceRequest, resp *TimeoutResponse) error {
	logrus.Info("RPC: ConnectionTimeout")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		resp.Timeout = int64(dev.ConnectionTimeout())
		return nil
	})
}

func (s *RPCNode) RunNativeScript(req *ExecuteRequest, resp *Void) error {
	logrus.Info("RPC: RunNativeScript")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		dev.RunNativeScript([]byte(req.Data))
		return nil
	})
}

func (s *RPCNode) SendAction(req *ExecuteRequest, resp *Void) error {
	logrus.Info("RPC: SendAction")
	return s.withDevice(req.DeviceID, func(dev device.Device) error {
		return dev.Send([]byte(req.Data))
	})
}
