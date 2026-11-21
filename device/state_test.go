package device

import "testing"

func TestParseState(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want State
	}{
		{"booted", "StateBooted", StateBooted},
		{"shutdown", "StateShutdown", StateShutdown},
		{"remote disconnected", "StateRemoteDisconnected", StateRemoteDisconnected},
		{"unauthorized", "StateUnauthorized", StateUnauthorized},
		{"node disconnected", "StateNodeDisconnected", StateNodeDisconnected},
		{"unknown", "StateUnknown", StateUnknown},
		// androiddevice, iosdevice, iossim, macos, web and node map an unknown
		// name to StateUnknown. unityeditor maps it to StateShutdown today;
		// ParseState gives StateUnknown, G27 switches unityeditor to it.
		{"unknown name", "StateSomethingNew", StateUnknown},
		{"empty", "", StateUnknown},
		{"locked has no wire name", "StateLocked", StateUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseState(tt.in); got != tt.want {
				t.Errorf("ParseState(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
