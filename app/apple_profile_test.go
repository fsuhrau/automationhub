package app

import (
	"testing"
	"time"
)

// signedProfile wraps the plist like a mobileprovision: junk around the xml
func signedProfile(t *testing.T, p provisioningPlist) []byte {
	t.Helper()
	return append(append([]byte("\x30\x80junk"), mustPlist(t, p)...), []byte("signature")...)
}

func TestReadProvisioning(t *testing.T) {
	cert := testCertificate(t, "Dev")
	p := provisioningPlist{
		Name: "Game", UUID: "u-1", AppIDName: "App", TeamName: "Team", TeamIdentifier: []string{"TEAM1", "TEAM2"},
		CreationDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), ExpirationDate: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
		ProvisionedDevices:    []string{"d1"},
		Entitlements:          map[string]interface{}{"get-task-allow": true},
		DeveloperCertificates: [][]byte{cert.Raw, []byte("not a certificate")},
	}
	profile, certs, err := readProvisioning(signedProfile(t, p))
	if err != nil {
		t.Fatal(err)
	}
	if profile.Name != "Game" || profile.UUID != "u-1" || profile.AppIDName != "App" || profile.TeamName != "Team" || profile.TeamID != "TEAM1" {
		t.Errorf("profile = %+v", profile)
	}
	if !profile.CreatedAt.Equal(p.CreationDate) || !profile.ExpiresAt.Equal(p.ExpirationDate) || len(profile.Devices) != 1 || profile.AllDevices {
		t.Errorf("profile = %+v", profile)
	}
	if profile.Entitlements["get-task-allow"] != true {
		t.Errorf("entitlements = %v", profile.Entitlements)
	}
	if len(certs) != 1 || certs[0].SHA256 != certificateOf(cert).SHA256 {
		t.Errorf("certs = %+v, the unreadable certificate is skipped", certs)
	}
}

func TestProfileKind(t *testing.T) {
	tests := []struct {
		name string
		p    provisioningPlist
		want string
	}{
		{"all devices", provisioningPlist{ProvisionsAllDevices: true, ProvisionedDevices: []string{"d"}}, "enterprise"},
		{"development", provisioningPlist{ProvisionedDevices: []string{"d"}, Entitlements: map[string]interface{}{"get-task-allow": true}}, "development"},
		{"ad hoc", provisioningPlist{ProvisionedDevices: []string{"d"}, Entitlements: map[string]interface{}{"get-task-allow": false}}, "ad-hoc"},
		{"app store", provisioningPlist{}, "app-store"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := profileKind(&tt.p); got != tt.want {
				t.Errorf("kind = %s, want %s", got, tt.want)
			}
			if got := profileOf(&tt.p); got.Kind != tt.want || got.TeamID != "" {
				t.Errorf("profile = %+v", got)
			}
		})
	}
}

func TestReadProvisioningErrors(t *testing.T) {
	for name, data := range map[string][]byte{
		"no plist":       []byte("binary only"),
		"end before xml": []byte("</plist> <?xml"),
		"broken plist":   []byte("<?xml version=\"1.0\"?><plist><dict><key>x</key></plist>"),
	} {
		t.Run(name, func(t *testing.T) {
			if profile, certs, err := readProvisioning(data); err == nil || profile != nil || certs != nil {
				t.Errorf("profile = %v, certs = %v, err = %v", profile, certs, err)
			}
		})
	}
}
