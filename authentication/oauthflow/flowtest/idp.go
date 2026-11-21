package flowtest

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

const (
	goodCode   = "good-code"
	accessTok  = "access-1"
	redirectTo = "http://hub.test/auth/oauth2"
)

// idp is a fake identity provider: authorization is never called (the tests
// only read the redirect), token and user endpoints are real.
type idp struct {
	*httptest.Server
	mu       sync.Mutex
	userJSON string
	status   int
	auth     string // Authorization header of the last user request
}

func newIDP(t *testing.T, userJSON string) *idp {
	t.Helper()
	p := &idp{userJSON: userJSON, status: http.StatusOK}
	mux := http.NewServeMux()
	mux.HandleFunc("/token", p.token)
	mux.HandleFunc("/user", p.user)
	p.Server = httptest.NewServer(mux)
	t.Cleanup(p.Close)
	return p
}

// token hands out the access token for the good code only.
func (p *idp) token(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if r.Form.Get("code") != goodCode {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"access_token":"` + accessTok + `","token_type":"bearer"}`))
}

// user answers with the configured status and JSON and remembers who asked.
func (p *idp) user(w http.ResponseWriter, r *http.Request) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.auth = r.Header.Get("Authorization")
	w.WriteHeader(p.status)
	_, _ = w.Write([]byte(p.userJSON))
}

// answerUsersWith changes what the user endpoint answers.
func (p *idp) answerUsersWith(status int, body string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status, p.userJSON = status, body
}

// lastAuth returns the Authorization header of the last user request.
func (p *idp) lastAuth() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.auth
}
