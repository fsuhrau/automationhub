// Package mcpproxy is a host-side proxy for the master's MCP endpoint. It
// keeps the access token outside of dev containers: the container talks to
// the proxy without a hub token, and the proxy adds the token, which it reads
// on the host (macOS Keychain, a file or the environment).
//
// Only the MCP endpoint and its build uploads are forwarded, so whoever
// reaches the proxy can't use the token for the rest of the API.
//
// Uploads: the begin_upload tool returns a URL next to the MCP endpoint. The
// proxy tells the master the address the client used (X-Forwarded-Host,
// -Proto, -Prefix), so that URL points back at the proxy, and forwards
// "PUT <mcp path>/uploads/<id>". That path needs no proxy key: the upload id
// is a single-use secret only the MCP session that created it knows.
package mcpproxy

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

// uploadIDPattern matches the ids of begin_upload.
var uploadIDPattern = regexp.MustCompile(`^[a-f0-9]{32,128}$`)

// projectPattern matches project identifiers in proxy paths. They must start
// with a letter or digit, so "." and ".." (also percent-encoded) are refused.
var projectPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// Config of the proxy.
type Config struct {
	// Upstream is the master, e.g. http://localhost:8002.
	Upstream *url.URL
	// Project is used for requests to /mcp. Requests to /<project>/mcp
	// name the project themselves.
	Project string
	// Token supplies the hub access token (an ahp_ personal token).
	Token TokenSource
	// Key, when set, must be sent by clients as "Authorization: Bearer <key>".
	// It is a local key of this proxy, not a hub token: it limits who on the
	// docker network can use the proxy and is worthless anywhere else.
	Key string
	// Logf logs requests and errors; nil uses log.Printf.
	Logf func(format string, args ...any)
}

// Handler returns the proxy handler.
func Handler(cfg Config) http.Handler {
	logf := cfg.Logf
	if logf == nil {
		logf = log.Printf
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			rt := r.In.Context().Value(targetKey{}).(route)
			r.SetURL(cfg.Upstream)
			r.Out.URL.Path = strings.TrimSuffix(cfg.Upstream.Path, "/") + rt.target
			r.Out.URL.RawPath = ""
			r.Out.Host = cfg.Upstream.Host
			// the address the client used, for URLs the master hands out
			r.SetXForwarded()
			r.Out.Header.Set("X-Forwarded-Prefix", rt.mcpPath)
		},
		// MCP answers with server-sent events; pass them on at once
		FlushInterval: -1,
		ModifyResponse: func(resp *http.Response) error {
			// the master's session cookie belongs to the host
			resp.Header.Del("Set-Cookie")
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logf("mcp-proxy: upstream %s: %v", cfg.Upstream, err)
			http.Error(w, "automationhub master not reachable: "+err.Error(), http.StatusBadGateway)
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rt, ok := resolve(r.URL.Path, cfg.Project)
		if !ok {
			http.Error(w, "only the MCP endpoint is forwarded: /mcp or /<project>/mcp (and their uploads/<id>)", http.StatusNotFound)
			return
		}
		allowed := r.Method == http.MethodGet || r.Method == http.MethodPost || r.Method == http.MethodDelete
		if rt.upload {
			allowed = r.Method == http.MethodPut
		}
		if !allowed {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Key != "" && !rt.upload && !validKey(r.Header.Get("Authorization"), cfg.Key) {
			http.Error(w, "missing or wrong proxy key", http.StatusUnauthorized)
			return
		}
		token, err := cfg.Token.Token()
		if err != nil {
			logf("mcp-proxy: %v", err)
			http.Error(w, "no AutomationHub token on the host: "+err.Error(), http.StatusBadGateway)
			return
		}

		// nothing of the client's credentials goes upstream
		for _, h := range []string{"Authorization", "Cookie", "X-Auth-Token", "Proxy-Authorization"} {
			r.Header.Del(h)
		}
		r.Header.Set("X-Auth-Token", token)
		proxy.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), targetKey{}, rt)))
	})
}

// route is a request path the proxy forwards.
type route struct {
	// target is the path on the master
	target string
	// mcpPath is the client's path of the MCP endpoint, e.g. /game/mcp
	mcpPath string
	upload  bool
}

// resolve maps /mcp and /<project>/mcp to the master's /api/<project>/mcp,
// and <those>/uploads/<id> to /api/<project>/mcp/uploads/<id>.
func resolve(path, defaultProject string) (route, bool) {
	mcpPath, uploadID, isUpload := strings.Cut(path, "/uploads/")
	if isUpload && !uploadIDPattern.MatchString(uploadID) {
		return route{}, false
	}
	project := defaultProject
	switch {
	case mcpPath == "/mcp":
	case strings.HasSuffix(mcpPath, "/mcp") && strings.Count(mcpPath, "/") == 2:
		project = strings.TrimSuffix(strings.TrimPrefix(mcpPath, "/"), "/mcp")
	default:
		return route{}, false
	}
	if project == "" || !projectPattern.MatchString(project) {
		return route{}, false
	}
	rt := route{target: fmt.Sprintf("/api/%s/mcp", project), mcpPath: mcpPath}
	if isUpload {
		rt.target += "/uploads/" + uploadID
		rt.upload = true
	}
	return rt, true
}

func validKey(header, key string) bool {
	given, ok := strings.CutPrefix(header, "Bearer ")
	return ok && subtle.ConstantTimeCompare([]byte(given), []byte(key)) == 1
}

type targetKey struct{}
