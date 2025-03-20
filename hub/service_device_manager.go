package hub

import (
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/iosdevice"
	"github.com/fsuhrau/automationhub/device/iossim"
	"github.com/fsuhrau/automationhub/device/macos"
	"github.com/fsuhrau/automationhub/device/unityeditor"
	"github.com/fsuhrau/automationhub/device/web"
)

func (s *Service) RegisterDeviceHandler(masterUrl string) {
	// start device observer thread
	if d, ok := s.cfg.DeviceManager[androiddevice.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager android_device")
		s.deviceManager.AddHandler(androiddevice.NewHandler(s.sd))
	}
	if d, ok := s.cfg.DeviceManager[iossim.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager ios_sim")
		s.deviceManager.AddHandler(iossim.NewHandler(s.sd))
	}
	if d, ok := s.cfg.DeviceManager[iosdevice.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager ios_device")
		s.deviceManager.AddHandler(iosdevice.NewHandler(s.sd))
	}
	if d, ok := s.cfg.DeviceManager[macos.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager macos")
		s.deviceManager.AddHandler(macos.NewHandler(s.sd))
	}
	if d, ok := s.cfg.DeviceManager[unityeditor.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager unity_editor")
		s.deviceManager.AddHandler(unityeditor.NewHandler(d, s.sd))
	}
	if d, ok := s.cfg.DeviceManager[web.Manager]; ok && d.Enabled {
		s.logger.Info("adding manager web")
		s.deviceManager.AddHandler(web.NewHandler(d, s.sd))
	}
}
