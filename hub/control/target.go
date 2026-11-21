package control

import "github.com/fsuhrau/automationhub/device"

// Target is the device a control call acts on and the token of the lease that
// allows the call.
type Target struct {
	Device device.Device
	Token  string
}

// authorize checks the token against the lease of the device and renews it.
func (s *Service) authorize(t Target) error {
	return s.Authorize(t.Device.DeviceID(), t.Token)
}

// input authorizes the call and returns the input role of the device.
func (s *Service) input(t Target) (device.InputCapable, error) {
	if err := s.authorize(t); err != nil {
		return nil, err
	}
	input, ok := t.Device.(device.InputCapable)
	if !ok {
		return nil, device.ErrUnsupported
	}
	return input, nil
}
