package device

import "github.com/fsuhrau/automationhub/storage/models"

// DisplayName is how runs, logs and tools name a device to a person: its
// alias, else its name, else its identifier. A nil device has no name.
func DisplayName(d *models.Device) string {
	if d == nil {
		return ""
	}
	for _, name := range []string{d.Alias, d.Name} {
		if name != "" {
			return name
		}
	}
	return d.DeviceIdentifier
}
