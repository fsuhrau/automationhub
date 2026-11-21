package settings_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

// nodeBody is the node with its token as the settings send it.
type nodeBody struct {
	models.Node
	Token string `json:"token"`
}

// fixture is an api with the projects mine and other.
type fixture struct {
	*apitest.Env
	mine, other *models.Project
}

func newFixture(t *testing.T, opts ...apitest.Option) fixture {
	t.Helper()
	env := apitest.New(t, opts...)
	return fixture{
		Env:   env,
		mine:  testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "mine" }),
		other: testkit.Project(env.DB, func(p *models.Project) { p.Identifier = "other" }),
	}
}

func liveTokens(db *gorm.DB, where string, args ...any) int64 {
	var n int64
	db.Model(&models.AccessToken{}).Where(where, args...).Count(&n)
	return n
}

func raw(env *apitest.Env, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return env.Request(req)
}

func TestUsers(t *testing.T) {
	f := newFixture(t)
	testkit.User(f.DB, testkit.WithName[models.User]("Alice"))
	rec := f.Do("GET", "/api/mine/settings/users", nil)
	if users := testkit.Decode[[]models.User](t, rec); rec.Code != http.StatusOK || len(users) != 1 || users[0].Name != "Alice" {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}

func TestSettingsOfAnUnknownProjectAreNotFound(t *testing.T) {
	f := newFixture(t)
	for _, path := range []string{"/users", "/access_tokens", "/nodes"} {
		if rec := f.Do("GET", "/api/nope/settings"+path, nil); rec.Code != http.StatusNotFound {
			t.Errorf("%s: %d", path, rec.Code)
		}
	}
}

// A project's settings only revoke the project's own tokens; before, any
// token could be deleted by id.
func TestProjectTokens(t *testing.T) {
	f := newFixture(t)
	if rec := f.Do("POST", "/api/mine/settings/access_token", map[string]string{"name": "  "}); rec.Code != http.StatusBadRequest {
		t.Errorf("no name: %d", rec.Code)
	}
	if rec := raw(f.Env, "POST", "/api/mine/settings/access_token", `{`); rec.Code != http.StatusBadRequest {
		t.Errorf("malformed: %d", rec.Code)
	}
	rec := f.Do("POST", "/api/mine/settings/access_token", map[string]string{"name": " ci "})
	mine := testkit.Decode[models.AccessToken](t, rec)
	if rec.Code != http.StatusCreated || mine.Name != "ci" || mine.Token == "" || *mine.ProjectID != f.mine.ID {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	foreign := testkit.Token(f.DB, func(a *models.AccessToken) { a.ProjectID, a.Name, a.Token = &f.other.ID, "theirs", "t-other" })
	userID := uint(1)
	personal := testkit.Token(f.DB, func(a *models.AccessToken) { a.UserID, a.Name, a.Token = &userID, "personal", "sha256:x" })

	list := testkit.Decode[[]models.AccessToken](t, f.Do("GET", "/api/mine/settings/access_tokens", nil))
	if len(list) != 1 || list[0].ID != mine.ID {
		t.Fatalf("list %+v", list)
	}
	for _, id := range []uint{foreign.ID, personal.ID, 999} {
		if rec := f.Do("DELETE", fmt.Sprintf("/api/mine/settings/access_token/%d", id), nil); rec.Code != http.StatusNotFound {
			t.Errorf("deleting token %d of someone else: %d", id, rec.Code)
		}
	}
	if liveTokens(f.DB, "id IN ?", []uint{foreign.ID, personal.ID}) != 2 {
		t.Errorf("another project's or a personal token was deleted")
	}
	if rec := f.Do("DELETE", "/api/mine/settings/access_token/abc", nil); rec.Code != http.StatusBadRequest {
		t.Errorf("token id not a number: %d", rec.Code)
	}
	if rec := f.Do("DELETE", fmt.Sprintf("/api/mine/settings/access_token/%d", mine.ID), nil); rec.Code != http.StatusOK {
		t.Errorf("own token: %d", rec.Code)
	}
	if liveTokens(f.DB, "id = ?", mine.ID) != 0 {
		t.Errorf("own token still there")
	}
}

// A node is set up with its identifier and token: creating it returns the
// token once, a new token replaces the old one, deleting the node revokes it.
func TestNodes(t *testing.T) {
	f := newFixture(t)
	f.Nodes.Status = 3
	if rec := f.Do("POST", "/api/mine/settings/node", map[string]string{"name": ""}); rec.Code != http.StatusBadRequest {
		t.Errorf("no name: %d", rec.Code)
	}
	if rec := raw(f.Env, "POST", "/api/mine/settings/node", `{`); rec.Code != http.StatusBadRequest {
		t.Errorf("malformed: %d", rec.Code)
	}
	rec := f.Do("POST", "/api/mine/settings/node", map[string]string{"name": " Mac mini "})
	created := testkit.Decode[nodeBody](t, rec)
	if rec.Code != http.StatusCreated || created.Name != "Mac mini" || created.Identifier == "" || created.Token == "" {
		t.Fatalf("create: %d %s", rec.Code, rec.Body)
	}
	if liveTokens(f.DB, "node_id = ? AND token = ?", created.ID, created.Token) != 1 {
		t.Fatalf("the returned token isn't the node's")
	}

	nodes := testkit.Decode[[]models.Node](t, f.Do("GET", "/api/mine/settings/nodes", nil))
	if len(nodes) != 1 || nodes[0].Status != 3 {
		t.Fatalf("nodes %+v", nodes)
	}
	path := fmt.Sprintf("/api/mine/settings/nodes/%d", created.ID)
	if got := testkit.Decode[models.Node](t, f.Do("GET", path, nil)); got.Identifier != created.Identifier {
		t.Fatalf("node %+v", got)
	}

	rec = f.Do("POST", path+"/token", nil)
	issued := testkit.Decode[nodeBody](t, rec)
	if rec.Code != http.StatusOK || issued.Token == "" || issued.Token == created.Token {
		t.Fatalf("new token: %d %s", rec.Code, rec.Body)
	}
	if liveTokens(f.DB, "node_id = ?", created.ID) != 1 || liveTokens(f.DB, "token = ?", created.Token) != 0 {
		t.Errorf("the old token still works")
	}

	if rec := f.Do("DELETE", path, nil); rec.Code != http.StatusOK {
		t.Fatalf("delete: %d", rec.Code)
	}
	if liveTokens(f.DB, "node_id = ?", created.ID) != 0 {
		t.Errorf("a deleted node's token still works")
	}
}

// Regression: getNodeStatus passed the id to gorm as a raw string, so an
// unknown node was a 500; the other node routes answered a bad id with 404.
func TestNodeIDs(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct{ method, path string }{
		{"GET", "/api/mine/settings/nodes/%s"},
		{"DELETE", "/api/mine/settings/nodes/%s"},
		{"POST", "/api/mine/settings/nodes/%s/token"},
	} {
		for id, want := range map[string]int{"99": http.StatusNotFound, "abc": http.StatusBadRequest} {
			if rec := f.Do(tc.method, fmt.Sprintf(tc.path, id), nil); rec.Code != want {
				t.Errorf("%s %s: %d %s, want %d", tc.method, id, rec.Code, rec.Body, want)
			}
		}
	}
}

// The settings hand out tokens, so an access token must not reach them:
// before, a project token could create a node and get a node token that
// works in every project.
func TestSettingsRefuseAccessTokens(t *testing.T) {
	f := newFixture(t, apitest.WithAccessToken(1))
	for _, c := range []struct{ method, path string }{
		{"GET", "/users"},
		{"POST", "/node"},
		{"GET", "/nodes"},
		{"GET", "/nodes/1"},
		{"POST", "/nodes/1/token"},
		{"DELETE", "/nodes/1"},
		{"POST", "/access_token"},
		{"GET", "/access_tokens"},
		{"DELETE", "/access_token/1"},
	} {
		if rec := f.Do(c.method, "/api/mine/settings"+c.path, map[string]string{"name": "x"}); rec.Code != http.StatusForbidden {
			t.Errorf("%s %s with an access token: %d %s", c.method, c.path, rec.Code, rec.Body)
		}
	}
	if n := liveTokens(f.DB, "node_id IS NOT NULL"); n != 0 {
		t.Fatalf("an access token minted %d node tokens", n)
	}
}
