package flowtest

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/fsuhrau/automationhub/authentication/oauthflow"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type hub struct {
	*httptest.Server
	db  *gorm.DB
	idp *idp
}

func newHub(t *testing.T, p Case) *hub {
	t.Helper()
	gin.SetMode(gin.TestMode)
	idp := newIDP(t, p.UserJSON)
	cred := filepath.Join(t.TempDir(), "cred.json")
	check(t, os.WriteFile(cred, []byte(`{"clientid":"client-1","secret":"s3"}`), 0o600) == nil, "write credentials")
	db := testkit.DB(t)
	flow, err := p.Build(db, idp.URL, oauthflow.Settings{
		RedirectURL: redirectTo, CredFile: cred, Scopes: []string{"read", "user"}, Secret: []byte("test-secret-test-secret-test-sec"),
	})
	check(t, err == nil, "build provider: %v", err)
	r := gin.New()
	r.Use(flow.Session("hub"))
	flow.Routes(r)
	r.GET("/private", flow.SessionHandler(), func(c *gin.Context) { c.String(http.StatusOK, "in") })
	h := &hub{Server: httptest.NewServer(r), db: db, idp: idp}
	t.Cleanup(h.Close)
	return h
}

// browser is a client with its own cookies that doesn't follow redirects.
func browser(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	check(t, err == nil, "cookie jar: %v", err)
	return &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

func get(t *testing.T, c *http.Client, u string) (*http.Response, string) {
	t.Helper()
	resp, err := c.Get(u)
	check(t, err == nil, "GET %s: %v", u, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	check(t, err == nil, "read body: %v", err)
	return resp, string(body)
}

// startLogin follows the first redirect and returns the provider URL it points to.
func startLogin(t *testing.T, h *hub, c *http.Client) *url.URL {
	t.Helper()
	resp, _ := get(t, c, h.URL+"/auth/oauth2")
	check(t, resp.StatusCode == http.StatusTemporaryRedirect, "login start: %d", resp.StatusCode)
	loc, err := resp.Location()
	check(t, err == nil, "redirect location: %v", err)
	return loc
}

func callback(h *hub, state, code string) string {
	return h.URL + "/auth/oauth2?" + url.Values{"state": {state}, "code": {code}}.Encode()
}
