package wiring

import (
	"github.com/fsuhrau/automationhub/config"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/device/androiddevice"
	"github.com/fsuhrau/automationhub/device/iosdevice"
	"github.com/fsuhrau/automationhub/device/iossim"
	"github.com/fsuhrau/automationhub/device/macos"
	"github.com/fsuhrau/automationhub/device/unityeditor"
	"github.com/fsuhrau/automationhub/device/web"
	"github.com/fsuhrau/automationhub/storage"
	"github.com/sirupsen/logrus"
)

// handlerFactory builds the handler of one device manager.
type handlerFactory struct {
	name  string
	build func(cfg config.Manager, store storage.Device) device.Handler
}

// handlerFactories lists the device managers a node can run.
func handlerFactories() []handlerFactory {
	return []handlerFactory{
		{androiddevice.Manager, func(_ config.Manager, ds storage.Device) device.Handler { return androiddevice.NewHandler(ds) }},
		{iossim.Manager, func(_ config.Manager, ds storage.Device) device.Handler { return iossim.NewHandler(ds) }},
		{iosdevice.Manager, newIOSHandler},
		{macos.Manager, func(_ config.Manager, ds storage.Device) device.Handler { return macos.NewHandler(ds) }},
		{unityeditor.Manager, func(cfg config.Manager, ds storage.Device) device.Handler { return unityeditor.NewHandler(cfg, ds) }},
		{web.Manager, func(cfg config.Manager, ds storage.Device) device.Handler { return web.NewHandler(cfg, ds) }},
	}
}

// newIOSHandler builds the iOS handler and adds the hook that reads the
// WebDriverAgent URL from the go-ios log output to the standard logger.
func newIOSHandler(cfg config.Manager, ds storage.Device) device.Handler {
	h := iosdevice.NewHandlerWithDeps(cfg, ds, iosdevice.Deps{})
	logrus.AddHook(h.Hook())
	return h
}

// nodeHandlers builds the handlers of the enabled device managers of the
// config, in the order of handlerFactories.
func nodeHandlers(cfg config.Service, store storage.Device, log logrus.FieldLogger) []device.Handler {
	var handlers []device.Handler
	for _, f := range handlerFactories() {
		managerCfg, ok := cfg.DeviceManager[f.name]
		if !ok || !managerCfg.Enabled {
			continue
		}
		log.Infof("adding manager %s", f.name)
		handlers = append(handlers, f.build(managerCfg, store))
	}
	return handlers
}
