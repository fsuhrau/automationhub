package flowtest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/accessstore"
	"github.com/fsuhrau/automationhub/storage/models"
)

// sessionUser reads the signed-in user the session endpoint reports.
func sessionUser(t *testing.T, h *hub, c *http.Client) authentication.User {
	t.Helper()
	_, body := get(t, c, h.URL+"/auth/session")
	var got struct{ User authentication.User }
	check(t, json.Unmarshal([]byte(body), &got) == nil, "session json: %s", body)
	return got.User
}

// signIn runs a whole login in a new browser.
func signIn(t *testing.T, h *hub) *http.Client {
	t.Helper()
	c := browser(t)
	statusOf(t, c, callback(h, login(t, h, c), goodCode))
	return c
}

// A user who renames themselves at the provider to the name of another hub
// user stays themselves: the session and the store identify by hub user id.
func displayNameDoesNotSelectTheAccount(t *testing.T, p Case) {
	h := newHub(t, p)
	own := sessionUser(t, h, signIn(t, h))
	check(t, own.ID != 0, "the session carries no hub user id: %+v", own)
	victim := testkit.User(h.db, testkit.WithName[models.User]("admin"))

	h.idp.answerUsersWith(http.StatusOK, strings.Replace(p.UserJSON, "Octo Cat", victim.Name, 1))
	renamed := sessionUser(t, h, signIn(t, h))
	check(t, renamed.ID == own.ID && renamed.ID != victim.ID, "signed in as user %d, want %d (victim %d)", renamed.ID, own.ID, victim.ID)
	check(t, renamed.Name == "Octo Cat", "display name of the session: %q", renamed.Name)

	found, err := accessstore.New(h.db).FindUser(context.Background(), renamed)
	check(t, err == nil && found.ID == own.ID, "stored user of the session: %+v, %v", found, err)
}

// Providers that give no email (the generic one always, GitHub for private
// emails) must not make accounts collide: every provider user gets their own.
func usersWithoutEmailGetOwnAccounts(t *testing.T, p Case) {
	h := newHub(t, p)
	var ids [2]uint
	for i, answer := range p.NoEmailUsersJSON {
		h.idp.answerUsersWith(http.StatusOK, answer)
		c := signIn(t, h)
		check(t, statusOf(t, c, h.URL+"/private") == http.StatusOK, "user %d is not signed in", i)
		ids[i] = sessionUser(t, h, c).ID
	}
	check(t, ids[0] != 0 && ids[0] != ids[1], "accounts of the two users: %v", ids)
	check(t, countUsers(h) == 2, "users: %d, want 2", countUsers(h))
}
