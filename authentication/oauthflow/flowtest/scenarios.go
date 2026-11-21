package flowtest

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/fsuhrau/automationhub/storage/models"
)

func countUsers(h *hub) int64 {
	var n int64
	h.db.Model(&models.User{}).Count(&n)
	return n
}

// statusOf calls the url and returns the status code.
func statusOf(t *testing.T, c *http.Client, u string) int {
	t.Helper()
	resp, _ := get(t, c, u)
	return resp.StatusCode
}

// login starts a login in the browser and returns its state.
func login(t *testing.T, h *hub, c *http.Client) string {
	t.Helper()
	return startLogin(t, h, c).Query().Get("state")
}

func loginFlow(t *testing.T, p Case) {
	h := newHub(t, p)
	c := browser(t)

	loc := startLogin(t, h, c)
	q := loc.Query()
	check(t, strings.HasPrefix(loc.String(), h.idp.URL+"/auth"), "provider redirect: %s", loc)
	check(t, q.Get("client_id") == "client-1", "client id: %s", loc)
	check(t, q.Get("scope") == "read user", "scopes: %s", loc)
	check(t, q.Get("redirect_uri") == redirectTo, "redirect uri: %s", loc)
	check(t, q.Get("state") != "", "state: %s", loc)

	got := statusOf(t, c, callback(h, q.Get("state"), goodCode))
	check(t, got == p.CallbackStatus, "callback: %d", got)
	check(t, h.idp.lastAuth() == "Bearer "+accessTok, "user request auth: %q", h.idp.lastAuth())
	resp, body := get(t, c, h.URL+"/auth/session")
	check(t, resp.StatusCode == http.StatusOK && strings.Contains(body, `"login":"octo"`), "session after login: %d %s", resp.StatusCode, body)
	got = statusOf(t, c, h.URL+"/private")
	check(t, got == http.StatusOK, "guard after login: %d", got)
	var auth models.UserAuth
	err := h.db.Preload("User").First(&auth, "provider = ?", p.Name).Error
	check(t, err == nil && auth.User.Name == "Octo Cat", "user not stored: %v %+v", err, auth)

	// a second login finds the account instead of creating another
	c2 := browser(t)
	statusOf(t, c2, callback(h, login(t, h, c2), goodCode))
	check(t, countUsers(h) == 1, "users: %d, want 1", countUsers(h))
}

func refusesWrongOrReusedState(t *testing.T, p Case) {
	h := newHub(t, p)
	c := browser(t)
	state := login(t, h, c)

	got := statusOf(t, c, callback(h, "forged", goodCode))
	check(t, got == http.StatusBadRequest, "wrong state: %d", got)
	// a failed check burns the stored state: the real one no longer works
	got = statusOf(t, c, callback(h, state, goodCode))
	check(t, got == http.StatusBadRequest, "burnt state: %d", got)
	// a browser that never started a login has no state at all
	got = statusOf(t, browser(t), callback(h, state, goodCode))
	check(t, got == http.StatusBadRequest, "foreign state: %d", got)
	got = statusOf(t, c, h.URL+"/auth/session")
	check(t, got == http.StatusUnauthorized, "signed in despite refused state: %d", got)
	check(t, countUsers(h) == 0, "users created by refused callbacks: %d", countUsers(h))
}

func providerFailures(t *testing.T, p Case) {
	h := newHub(t, p)

	c := browser(t)
	got := statusOf(t, c, callback(h, login(t, h, c), "bad-code"))
	check(t, got == http.StatusBadRequest, "rejected code: %d", got)

	h.idp.answerUsersWith(http.StatusInternalServerError, "not json")
	c = browser(t)
	got = statusOf(t, c, callback(h, login(t, h, c), goodCode))
	// some providers refuse a broken user, the generic one signs in an empty user
	check(t, !p.FailsOnBrokenUser || got == http.StatusBadRequest, "broken user endpoint: %d", got)

	err := h.db.Migrator().DropTable(&models.UserAuth{}, &models.User{})
	check(t, err == nil, "drop tables: %v", err)
	h.idp.answerUsersWith(http.StatusOK, p.UserJSON)
	c = browser(t)
	got = statusOf(t, c, callback(h, login(t, h, c), goodCode))
	check(t, got == http.StatusInternalServerError, "store failure: %d", got)
}

func signedInUserPassesLoginRoute(t *testing.T, p Case) {
	h := newHub(t, p)
	c := browser(t)
	statusOf(t, c, callback(h, login(t, h, c), goodCode))
	got := statusOf(t, c, h.URL+"/auth/oauth2")
	check(t, got == http.StatusOK, "signed in user restarted the login: %d", got)
}

// Every browser has its own state, so parallel logins don't invalidate each other.
func concurrentLoginsDoNotInterfere(t *testing.T, p Case) {
	h := newHub(t, p)
	const n = 8
	states := make([]string, n)
	clients := make([]*http.Client, n)
	seen := map[string]bool{}
	for i := range clients {
		clients[i] = browser(t)
		states[i] = login(t, h, clients[i])
		check(t, !seen[states[i]], "two logins got the same state")
		seen[states[i]] = true
	}

	var wg sync.WaitGroup
	codes := make([]int, n)
	for i := range clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statusOf(t, clients[i], callback(h, states[i], goodCode))
			codes[i] = statusOf(t, clients[i], h.URL+"/auth/session")
		}()
	}
	wg.Wait()
	for i, code := range codes {
		check(t, code == http.StatusOK, "login %d: session status %d", i, code)
	}
}

func sessionAnswersNameTheMethod(t *testing.T, p Case) {
	h := newHub(t, p)
	resp, body := get(t, browser(t), h.URL+"/auth/session")
	var got map[string]string
	err := json.Unmarshal([]byte(body), &got)
	check(t, err == nil, "session json: %v", err)
	check(t, resp.StatusCode == http.StatusUnauthorized, "session status: %d", resp.StatusCode)
	check(t, got["method"] == p.Name && got["url"] == "/auth/oauth2", "session body: %v", got)
	resp, body = get(t, browser(t), h.URL+"/private")
	check(t, resp.StatusCode == http.StatusUnauthorized && strings.Contains(body, `"url":"/auth/oauth2"`), "guard: %d %s", resp.StatusCode, body)
}
