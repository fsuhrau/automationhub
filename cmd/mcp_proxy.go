package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/authentication/token"
	"github.com/fsuhrau/automationhub/tools/mcpproxy"
	"github.com/spf13/cobra"
)

type mcpProxyFlags struct {
	listen          string
	upstream        string
	project         string
	tokenFile       string
	keychainService string
	keyFile         string
	noKey           bool
}

func newMCPProxyCmd() *cobra.Command {
	var flags mcpProxyFlags
	cmd := &cobra.Command{
		Use:   "mcp-proxy",
		Short: "Forward the MCP endpoint and add the access token on this machine",
		Long: `Runs on the host and forwards /mcp (or /<project>/mcp) to the master's
/api/<project>/mcp, adding the X-Auth-Token header. Dev containers connect to
the proxy and never see the hub token. Only the MCP endpoint is forwarded.

The token (an ahp_ personal token from the profile page) is read, in order, from
  $AUTOMATIONHUB_TOKEN,
  the macOS Keychain (generic password, service --keychain-service, account =
  the master's host:port), e.g.
    security add-generic-password -U -s automationhub -a localhost:8002 -w
  the file --token-file (mode 600).
It is re-read every minute, so a rotated token needs no restart.

Clients must send "Authorization: Bearer <proxy key>". The key is generated
into --key-file; it only works against this proxy.`,
		RunE: func(*cobra.Command, []string) error { return runMCPProxy(flags) },
	}
	home, _ := os.UserHomeDir()
	f := cmd.Flags()
	f.StringVar(&flags.listen, "listen", "127.0.0.1:18092", "address to listen on")
	f.StringVar(&flags.upstream, "upstream", "http://localhost:8002", "URL of the master")
	f.StringVar(&flags.project, "project", "", "project identifier served at /mcp")
	f.StringVar(&flags.tokenFile, "token-file", filepath.Join(home, ".config/automationhub/token"), "file with the access token (mode 600)")
	f.StringVar(&flags.keychainService, "keychain-service", "automationhub", "macOS Keychain service of the access token")
	f.StringVar(&flags.keyFile, "key-file", filepath.Join(home, ".local/state/automationhub-mcp/proxy-key"), "file with the proxy key, created when missing")
	f.BoolVar(&flags.noKey, "no-key", false, "accept clients without the proxy key")
	return cmd
}

func runMCPProxy(flags mcpProxyFlags) error {
	upstream, err := url.Parse(flags.upstream)
	if err != nil || upstream.Host == "" {
		return fmt.Errorf("invalid --upstream %q", flags.upstream)
	}
	source := mcpproxy.CachedToken(mcpproxy.FirstToken(
		mcpproxy.EnvToken("AUTOMATIONHUB_TOKEN"),
		mcpproxy.KeychainToken(flags.keychainService, upstream.Host),
		mcpproxy.FileToken(flags.tokenFile),
	), time.Minute)
	warnAboutToken(source)

	key, err := proxyKey(flags)
	if err != nil {
		return err
	}
	warnIfNotLoopback(flags.listen)

	handler := mcpproxy.Handler(mcpproxy.Config{Upstream: upstream, Project: flags.project, Token: source, Key: key})
	log.Printf("mcp-proxy: %s → %s/api/<project>/mcp on %s", strings.TrimSpace(proxyPaths(flags.project)), strings.TrimSuffix(upstream.String(), "/"), flags.listen)
	server := &http.Server{Addr: flags.listen, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	return server.ListenAndServe()
}

// warnAboutToken checks the token at start, but keeps running: it can be added later.
func warnAboutToken(source mcpproxy.TokenSource) {
	t, err := source.Token()
	if err != nil {
		log.Printf("mcp-proxy: warning: %v", err)
	} else if !token.IsPersonal(t) {
		log.Printf("mcp-proxy: warning: the token is not a personal token (%s…); device locks will show the token name, not yours", token.PersonalPrefix)
	}
}

// proxyKey is the key clients must send, empty with --no-key.
func proxyKey(flags mcpProxyFlags) (string, error) {
	if flags.noKey {
		return "", nil
	}
	return mcpproxy.LoadOrCreateKey(flags.keyFile)
}

func warnIfNotLoopback(listen string) {
	host, _, err := net.SplitHostPort(listen)
	if err != nil {
		return
	}
	if ip := net.ParseIP(host); host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		log.Printf("mcp-proxy: warning: listening on %s, not only on loopback", listen)
	}
}

func proxyPaths(project string) string {
	paths := "/<project>/mcp"
	if project != "" {
		paths = "/mcp (" + project + ") and " + paths
	}
	return paths
}
