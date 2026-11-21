package oauthflow

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type fakeStrategy struct{}

func (fakeStrategy) Name() string              { return "fake" }
func (fakeStrategy) Endpoint() oauth2.Endpoint { return oauth2.Endpoint{AuthURL: "http://idp/auth"} }
func (fakeStrategy) SuccessRedirect() string   { return "" }
func (fakeStrategy) Profile(context.Context, *http.Client) (Profile, error) {
	return Profile{}, nil
}

// failingSession remembers values and fails to save them.
type failingSession struct {
	sessions.Session
	values map[interface{}]interface{}
}

func (s failingSession) Get(k interface{}) interface{} { return s.values[k] }
func (s failingSession) Set(k, v interface{})          { s.values[k] = v }
func (s failingSession) Delete(k interface{})          { delete(s.values, k) }
func (s failingSession) Save() error                   { return errors.New("save failed") }

func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cred.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadCredentials(t *testing.T) {
	tests := map[string]struct {
		path    string
		want    Credentials
		wantErr bool
	}{
		"ok":      {writeFile(t, `{"clientid":"id","secret":"s"}`), Credentials{"id", "s"}, false},
		"missing": {filepath.Join(t.TempDir(), "none.json"), Credentials{}, true},
		"broken":  {writeFile(t, `{`), Credentials{}, true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := LoadCredentials(tc.path)
			if (err != nil) != tc.wantErr || got != tc.want {
				t.Errorf("got %+v, %v", got, err)
			}
		})
	}
}

func TestNewFailsWithoutCredentials(t *testing.T) {
	if _, err := New(nil, fakeStrategy{}, Settings{CredFile: filepath.Join(t.TempDir(), "none.json")}); err == nil {
		t.Fatal("expected an error")
	}
}

func TestStateIsRandomAndUsedOnce(t *testing.T) {
	first, second := newState(), newState()
	if first == second {
		t.Fatal("states repeat")
	}
	session := failingSession{values: map[interface{}]interface{}{}}
	_ = storeState(session, "abc")
	for name, tc := range map[string]struct {
		store, got string
		want       bool
	}{
		"match":    {"abc", "abc", true},
		"mismatch": {"abc", "abd", false},
		"none":     {"", "", false},
	} {
		s := failingSession{values: map[interface{}]interface{}{}}
		if tc.store != "" {
			_ = storeState(s, tc.store)
		}
		if got := consumeState(s, tc.got); got != tc.want {
			t.Errorf("%s: %v", name, got)
		}
		if consumeState(s, tc.got) {
			t.Errorf("%s: state worked twice", name)
		}
	}
}

func TestLoginStartFailsWhenTheStateCannotBeStored(t *testing.T) {
	f := &Flow{conf: &oauth2.Config{}, strategy: fakeStrategy{}}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/auth/oauth2", nil)
	ctx.Set(sessions.DefaultKey, failingSession{values: map[interface{}]interface{}{}})
	f.redirectToProvider(ctx)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status %d", w.Code)
	}
}

func TestEnsureUser(t *testing.T) {
	db := testkit.DB(t)
	u := users{db: db, provider: "fake"}
	p := Profile{ProviderUserID: "7", Name: "Ada", Email: "ada@example.com"}

	first, err := u.ensure(p)
	if err != nil || first.ID == 0 || first.Name != "Ada" {
		t.Fatalf("created %+v, %v", first, err)
	}
	again, err := u.ensure(p)
	if err != nil || again.ID != first.ID {
		t.Fatalf("second login got %+v, %v; want user %d", again, err, first.ID)
	}
	var n int64
	db.Model(&models.UserAuth{}).Where("provider_user_id = ? AND email = ?", "7", "ada@example.com").Count(&n)
	if n != 1 {
		t.Errorf("%d accounts, want 1", n)
	}
}

// A taken display name never refuses or takes over a login: the new account
// gets a free name of its own.
func TestEnsureUserWithATakenName(t *testing.T) {
	db := testkit.DB(t)
	admin := testkit.User(db, testkit.WithName[models.User]("admin"))
	u := users{db: db, provider: "fake"}

	got, err := u.ensure(Profile{ProviderUserID: "8", Login: "mallory", Name: "admin"})

	if err != nil || got.ID == admin.ID || got.Name != "admin (fake)" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestEnsureUserFailures(t *testing.T) {
	tests := map[string]struct {
		setup func(db *gorm.DB)
		p     Profile
	}{
		"auth already taken by an email": {func(db *gorm.DB) {
			db.Create(&models.User{Name: "Ada", Auth: []models.UserAuth{{Provider: "fake", ProviderUserID: "7", Email: "a@x"}}})
		}, Profile{ProviderUserID: "8", Name: "Bob", Email: "a@x"}},
		"no auth table": {func(db *gorm.DB) { _ = db.Migrator().DropTable(&models.UserAuth{}) }, Profile{ProviderUserID: "8"}},
		"no user table": {func(db *gorm.DB) { _ = db.Migrator().DropTable(&models.User{}) }, Profile{ProviderUserID: "8"}},
		"every name taken": {func(db *gorm.DB) {
			var all []models.User
			for _, name := range nameCandidates(Profile{Name: "Ada"}, "fake") {
				all = append(all, models.User{Name: name})
			}
			db.Create(&all)
		}, Profile{ProviderUserID: "8", Name: "Ada"}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			db := testkit.DB(t)
			tc.setup(db)
			if got, err := (users{db: db, provider: "fake"}).ensure(tc.p); err == nil {
				t.Fatalf("expected an error, got %+v", got)
			}
		})
	}
}

func TestNameCandidates(t *testing.T) {
	tests := []struct {
		name string
		p    Profile
		want []string
	}{
		{"display name", Profile{Name: "Ada", Login: "ada", ProviderUserID: "7"}, []string{"Ada", "Ada (github)", "Ada (github 2)", "Ada (github 3)"}},
		{"login", Profile{Login: "ada", ProviderUserID: "7"}, []string{"ada", "ada (github)", "ada (github 2)"}},
		{"provider user id", Profile{ProviderUserID: "7"}, []string{"7", "7 (github)", "7 (github 2)"}},
		{"nothing", Profile{}, []string{"github", "github (github)", "github (github 2)"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nameCandidates(tc.p, "github")
			if len(got) != maxNameCounter+1 || !slices.Equal(got[:len(tc.want)], tc.want) {
				t.Fatalf("got %d names, start %q", len(got), got[:len(tc.want)])
			}
		})
	}
}

// A parallel login that creates the account between lookup and create is not an error.
func TestEnsureUserSurvivesAParallelCreate(t *testing.T) {
	db := testkit.DB(t)
	p := Profile{ProviderUserID: "7", Name: "Ada"}
	armed := true
	err := db.Callback().Create().Before("gorm:create").Register("parallel", func(tx *gorm.DB) {
		if !armed {
			return
		}
		armed = false
		other := models.User{Name: "Ada", Auth: []models.UserAuth{{Provider: "fake", ProviderUserID: "7"}}}
		if err := db.Session(&gorm.Session{NewDB: true}).Create(&other).Error; err != nil {
			t.Error(err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := (users{db: db, provider: "fake"}).ensure(p)
	if err != nil || got.Name != "Ada" {
		t.Errorf("parallel create reported as error: %+v, %v", got, err)
	}
}
