package control

import (
	"context"
	"fmt"

	"github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/device"
	"github.com/fsuhrau/automationhub/hub/manager"
)

// Launch says how a started app connects back to the hub: it connects to
// NodeURL with SessionID.
type Launch struct {
	SessionID string
	NodeURL   string
}

type nodeDevice interface {
	GetNodeID() manager.NodeIdentifier
	NodeManager() manager.Nodes
}

// InstallApp installs the app; for node devices the bundle is uploaded to the
// node first.
func (s *Service) InstallApp(ctx context.Context, t Target, params *app.Parameter) error {
	if err := s.authorize(t); err != nil {
		return err
	}
	if nd, ok := t.Device.(nodeDevice); ok {
		if err := nd.NodeManager().UploadApp(ctx, nd.GetNodeID(), params); err != nil {
			return fmt.Errorf("upload app to node: %w", err)
		}
	}
	return t.Device.InstallApp(params)
}

// StartApp starts the app. The app connects back as the launch says.
func (s *Service) StartApp(t Target, params *app.Parameter, launch Launch) error {
	if err := s.authorize(t); err != nil {
		return err
	}
	return t.Device.StartApp(nil, params, launch.SessionID, launch.NodeURL)
}

// StopApp stops the app and, when clearData is set and supported, wipes its data.
func (s *Service) StopApp(t Target, params *app.Parameter, clearData bool) error {
	if err := s.authorize(t); err != nil {
		return err
	}
	if err := t.Device.StopApp(params); err != nil {
		return err
	}
	if !clearData {
		return nil
	}
	clearer, ok := t.Device.(device.AppDataClearer)
	if !ok {
		return fmt.Errorf("clear app data: %w", device.ErrUnsupported)
	}
	return clearer.ClearAppData(params)
}
