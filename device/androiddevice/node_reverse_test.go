package androiddevice

import "testing"

func TestLoopbackPort(t *testing.T) {
	cases := []struct {
		url  string
		port string
		ok   bool
	}{
		{"ws://localhost:8004/device/connect", "8004", true},
		{"ws://127.0.0.1:8004/device/connect", "8004", true},
		{"ws://[::1]:9000/device/connect", "9000", true},
		{"ws://localhost/device/connect", "80", true},
		{"wss://localhost/device/connect", "443", true},
		{"ws://192.168.0.10:8004/device/connect", "", false},
		{"ws://hub.example.com:8004/device/connect", "", false},
		{"ftp://localhost/x", "", false},
		{"", "", false},
		{"::not a url", "", false},
	}
	for _, c := range cases {
		port, ok := loopbackPort(c.url)
		if port != c.port || ok != c.ok {
			t.Errorf("loopbackPort(%q) = %q, %v; want %q, %v", c.url, port, ok, c.port, c.ok)
		}
	}
}
