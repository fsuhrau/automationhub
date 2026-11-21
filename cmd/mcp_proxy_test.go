package cmd

import (
	"bytes"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/tools/mcpproxy"
)

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

func TestMCPProxyFlagDefaults(t *testing.T) {
	cmd := newMCPProxyCmd()
	for name, want := range map[string]string{"listen": "127.0.0.1:18092", "upstream": "http://localhost:8002", "keychain-service": "automationhub", "no-key": "false", "project": ""} {
		if got := cmd.Flags().Lookup(name).DefValue; got != want {
			t.Errorf("--%s default = %q, want %q", name, got, want)
		}
	}
}

func TestProxyPaths(t *testing.T) {
	if got := proxyPaths(""); got != "/<project>/mcp" {
		t.Errorf("got %q", got)
	}
	if got := proxyPaths("demo"); got != "/mcp (demo) and /<project>/mcp" {
		t.Errorf("got %q", got)
	}
}

func TestProxyKey(t *testing.T) {
	keyFile := filepath.Join(t.TempDir(), "state", "key")

	if key, err := proxyKey(mcpProxyFlags{noKey: true}); key != "" || err != nil {
		t.Errorf("no-key: %q, %v", key, err)
	}
	key, err := proxyKey(mcpProxyFlags{keyFile: keyFile})
	if err != nil || key == "" {
		t.Fatalf("created: %q, %v", key, err)
	}
	if again, _ := proxyKey(mcpProxyFlags{keyFile: keyFile}); again != key {
		t.Error("key changed")
	}
}

func TestWarnIfNotLoopback(t *testing.T) {
	tests := map[string]bool{
		"127.0.0.1:1":   false,
		"localhost:1":   false,
		"[::1]:1":       false,
		"0.0.0.0:1":     true,
		"example.org:1": true,
		"no-port":       false,
	}
	for listen, warns := range tests {
		t.Run(listen, func(t *testing.T) {
			buf := captureLog(t)
			warnIfNotLoopback(listen)
			if got := strings.Contains(buf.String(), "not only on loopback"); got != warns {
				t.Errorf("warned = %v, want %v", got, warns)
			}
		})
	}
}

func TestWarnAboutToken(t *testing.T) {
	tests := map[string]struct {
		source mcpproxy.TokenFunc
		want   string
	}{
		"missing":      {func() (string, error) { return "", errors.New("no token") }, "no token"},
		"not personal": {func() (string, error) { return "plain", nil }, "not a personal token"},
		"personal":     {func() (string, error) { return "ahp_abc", nil }, ""},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			buf := captureLog(t)
			warnAboutToken(tc.source)
			if tc.want == "" && buf.Len() != 0 || !strings.Contains(buf.String(), tc.want) {
				t.Errorf("log = %q, want %q", buf, tc.want)
			}
		})
	}
}

func TestRunMCPProxyRejectsAnInvalidUpstream(t *testing.T) {
	for _, upstream := range []string{"", "://bad", "no-host"} {
		err := runMCPProxy(mcpProxyFlags{upstream: upstream})
		if err == nil || !strings.Contains(err.Error(), "invalid --upstream") {
			t.Errorf("%q: err = %v", upstream, err)
		}
	}
}

func TestRunMCPProxyFailsOnAKeyFileItCannotCreate(t *testing.T) {
	captureLog(t)
	dir := t.TempDir()
	flags := mcpProxyFlags{upstream: "http://localhost:1", tokenFile: filepath.Join(dir, "none"), keyFile: dir}

	if err := runMCPProxy(flags); err == nil {
		t.Error("no error")
	}
}

func TestRunMCPProxyReturnsTheListenError(t *testing.T) {
	buf := captureLog(t)
	t.Setenv("AUTOMATIONHUB_TOKEN", "ahp_token")
	flags := mcpProxyFlags{
		upstream: "http://localhost:1/", project: "demo", noKey: true,
		listen: "not-an-address",
	}

	err := runMCPProxy(flags)

	if err == nil {
		t.Error("no error")
	}
	if !strings.Contains(buf.String(), "/mcp (demo) and /<project>/mcp") {
		t.Errorf("log = %s", buf)
	}
}

func TestMCPProxyCommandRuns(t *testing.T) {
	captureLog(t)
	resetViper(t)
	t.Setenv("AUTOMATIONHUB_TOKEN", "ahp_token")

	_, err := execute(t, "--config", writeConfigFile(t, sqliteConfig(t)), "mcp-proxy", "--no-key", "--listen", "bad")

	if err == nil {
		t.Error("no error")
	}
}
