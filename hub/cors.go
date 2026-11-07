package hub

import (
	"net"
	"net/url"
	"time"

	"github.com/gin-contrib/cors"
)

// corsConfig allows the origins listed under `cors` in the config. Without
// that list it allows only pages served from this machine (localhost on any
// port, such as the web UI dev server); the embedded web UI is same-origin
// and needs no entry. cors.New refuses an empty origin list, so the master
// would not start without this default.
func corsConfig(origins []string) cors.Config {
	cfg := cors.Config{
		AllowCredentials: true,
		AllowMethods:     []string{"POST, OPTIONS, GET, PUT", "DELETE"},
		AllowHeaders:     []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", "X-Auth-Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Access-Control-Allow-Origin"},
		MaxAge:           12 * time.Hour,
	}
	if len(origins) > 0 {
		cfg.AllowOrigins = origins
	} else {
		cfg.AllowOriginFunc = isLoopbackOrigin
	}
	return cfg
}

// isLoopbackOrigin reports whether origin is an http(s) page on this machine.
func isLoopbackOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}
	host := u.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
