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

var mcpProxyFlags struct {
	listen          string
	upstream        string
	project         string
	tokenFile       string
	keychainService string
	keyFile         string
	noKey           bool
}

var mcpProxyCmd = &cobra.Command{
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
	RunE: runMCPProxy,
}

func init() {
	home, _ := os.UserHomeDir()
	f := mcpProxyCmd.Flags()
	f.StringVar(&mcpProxyFlags.listen, "listen", "127.0.0.1:18092", "address to listen on")
	f.StringVar(&mcpProxyFlags.upstream, "upstream", "http://localhost:8002", "URL of the master")
	f.StringVar(&mcpProxyFlags.project, "project", "", "project identifier served at /mcp")
	f.StringVar(&mcpProxyFlags.tokenFile, "token-file", filepath.Join(home, ".config/automationhub/token"), "file with the access token (mode 600)")
	f.StringVar(&mcpProxyFlags.keychainService, "keychain-service", "automationhub", "macOS Keychain service of the access token")
	f.StringVar(&mcpProxyFlags.keyFile, "key-file", filepath.Join(home, ".local/state/automationhub-mcp/proxy-key"), "file with the proxy key, created when missing")
	f.BoolVar(&mcpProxyFlags.noKey, "no-key", false, "accept clients without the proxy key")
	rootCmd.AddCommand(mcpProxyCmd)
}

func runMCPProxy(cmd *cobra.Command, args []string) error {
	flags := mcpProxyFlags
	upstream, err := url.Parse(flags.upstream)
	if err != nil || upstream.Host == "" {
		return fmt.Errorf("invalid --upstream %q", flags.upstream)
	}

	source := mcpproxy.CachedToken(mcpproxy.FirstToken(
		mcpproxy.EnvToken("AUTOMATIONHUB_TOKEN"),
		mcpproxy.KeychainToken(flags.keychainService, upstream.Host),
		mcpproxy.FileToken(flags.tokenFile),
	), time.Minute)

	// check the token at start, but keep running: it can be added later
	if t, err := source.Token(); err != nil {
		log.Printf("mcp-proxy: warning: %v", err)
	} else if !token.IsPersonal(t) {
		log.Printf("mcp-proxy: warning: the token is not a personal token (%s…); device locks will show the token name, not yours", token.PersonalPrefix)
	}

	key := ""
	if !flags.noKey {
		if key, err = mcpproxy.LoadOrCreateKey(flags.keyFile); err != nil {
			return err
		}
	}

	if host, _, err := net.SplitHostPort(flags.listen); err == nil {
		if ip := net.ParseIP(host); host != "localhost" && (ip == nil || !ip.IsLoopback()) {
			log.Printf("mcp-proxy: warning: listening on %s, not only on loopback", flags.listen)
		}
	}

	handler := mcpproxy.Handler(mcpproxy.Config{Upstream: upstream, Project: flags.project, Token: source, Key: key})
	paths := "/<project>/mcp"
	if flags.project != "" {
		paths = "/mcp (" + flags.project + ") and " + paths
	}
	log.Printf("mcp-proxy: %s → %s/api/<project>/mcp on %s", strings.TrimSpace(paths), strings.TrimSuffix(upstream.String(), "/"), flags.listen)
	server := &http.Server{Addr: flags.listen, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	return server.ListenAndServe()
}
