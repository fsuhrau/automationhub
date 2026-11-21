package accessstore_test

import (
	"context"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/accessstore"
	"github.com/fsuhrau/automationhub/storage/models"
	"gorm.io/gorm"
)

func ctx() context.Context { return context.Background() }

func setup(t *testing.T) (*accessstore.Store, *gorm.DB) {
	t.Helper()
	db := testkit.DB(t)
	return accessstore.New(db), db
}

// closed is a store whose database is gone, to reach the error paths.
func closed(t *testing.T) *accessstore.Store {
	t.Helper()
	store, db := setup(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	_ = sqlDB.Close()
	return store
}

func ptr[T any](v T) *T { return &v }

func wantKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if apperr.KindOf(err) != kind {
		t.Fatalf("err %v has kind %v, want %v", err, apperr.KindOf(err), kind)
	}
}

func tokenCount(db *gorm.DB, where string, args ...any) int64 {
	var n int64
	db.Model(&models.AccessToken{}).Where(where, args...).Count(&n)
	return n
}

// The database is opened without gorm's error translation, so a duplicate
// is reported as the driver's error and not as a Conflict.
func TestCreateTokenFailsOnDuplicates(t *testing.T) {
	store, _ := setup(t)
	if err := store.CreateToken(ctx(), &models.AccessToken{Name: "a", Token: "same"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateToken(ctx(), &models.AccessToken{Name: "b", Token: "same"}); err == nil {
		t.Fatal("duplicate token stored")
	}
}

func TestProjectTokensAreScopedToTheProject(t *testing.T) {
	store, db := setup(t)
	mine, other := testkit.Project(db), testkit.Project(db)
	testkit.Token(db, func(a *models.AccessToken) { a.ProjectID, a.Name, a.Token = &mine.ID, "mine", "t1" })
	foreign := testkit.Token(db, func(a *models.AccessToken) { a.ProjectID, a.Name, a.Token = &other.ID, "theirs", "t2" })
	personal := testkit.Token(db, func(a *models.AccessToken) { a.UserID, a.Name, a.Token = ptr(uint(1)), "personal", "t3" })

	got, err := store.ProjectTokens(ctx(), mine.ID)
	if err != nil || len(got) != 1 || got[0].Name != "mine" {
		t.Fatalf("got %+v, %v", got, err)
	}
	for _, id := range []uint{foreign.ID, personal.ID, 999} {
		wantKind(t, store.DeleteProjectToken(ctx(), mine.ID, id), apperr.KindNotFound)
	}
	if tokenCount(db, "id IN ?", []uint{foreign.ID, personal.ID}) != 2 {
		t.Fatal("a token of someone else was deleted")
	}
	if err := store.DeleteProjectToken(ctx(), mine.ID, got[0].ID); err != nil || tokenCount(db, "id = ?", got[0].ID) != 0 {
		t.Fatalf("own token: %v", err)
	}
}

func TestDeleteNotFoundMessage(t *testing.T) {
	store, _ := setup(t)
	if err := store.DeleteProjectToken(ctx(), 1, 99); err == nil || err.Error() != "token 99 not found" {
		t.Fatalf("err %v", err)
	}
	if err := store.DeletePersonalToken(ctx(), 1, 99); err == nil || err.Error() != "token not found" {
		t.Fatalf("err %v", err)
	}
}

func TestPersonalTokensNewestFirstAndOnlyTheOwners(t *testing.T) {
	store, db := setup(t)
	alice, bob := testkit.User(db), testkit.User(db)
	first := testkit.Token(db, func(a *models.AccessToken) { a.UserID, a.Name, a.Token = &alice.ID, "first", "p1" })
	second := testkit.Token(db, func(a *models.AccessToken) { a.UserID, a.Name, a.Token = &alice.ID, "second", "p2" })
	bobs := testkit.Token(db, func(a *models.AccessToken) { a.UserID, a.Name, a.Token = &bob.ID, "bobs", "p3" })

	got, err := store.PersonalTokens(ctx(), alice.ID)
	if err != nil || len(got) != 2 || got[0].ID != second.ID || got[1].ID != first.ID {
		t.Fatalf("got %+v, %v", got, err)
	}
	wantKind(t, store.DeletePersonalToken(ctx(), alice.ID, bobs.ID), apperr.KindNotFound)
	if err := store.DeletePersonalToken(ctx(), alice.ID, first.ID); err != nil || tokenCount(db, "id = ?", first.ID) != 0 {
		t.Fatalf("own token: %v", err)
	}
}

func TestUsersAndFindUser(t *testing.T) {
	store, db := setup(t)
	db.Create(&models.User{Name: "Alice", Auth: []models.UserAuth{{Email: "alice@example.com", Provider: "password"}}})
	db.Create(&models.User{Name: "Bob", Auth: []models.UserAuth{{Email: "bob@example.com", Provider: "github", ProviderUserID: "bob-login"}}})

	if users, err := store.Users(ctx()); err != nil || len(users) != 2 {
		t.Fatalf("users %+v, %v", users, err)
	}
	tests := []struct {
		name string
		who  authentication.User
		want string
	}{
		{"first user", authentication.User{ID: 1}, "Alice"},
		{"second user", authentication.User{ID: 2}, "Bob"},
		// labels of the session never select the account
		{"name of another user", authentication.User{ID: 2, Name: "Alice", Email: "alice@example.com"}, "Bob"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := store.FindUser(ctx(), tc.who)
			if err != nil || got.Name != tc.want {
				t.Fatalf("got %+v, %v", got, err)
			}
		})
	}
	unknown := []authentication.User{
		{},
		{ID: 99},
		// no id: not found by name, login or email, however well they match
		{Name: "Alice", Login: "bob-login", Email: "alice@example.com"},
	}
	for _, who := range unknown {
		_, err := store.FindUser(ctx(), who)
		wantKind(t, err, apperr.KindNotFound)
	}
}

func TestFindUserReportsDatabaseFailures(t *testing.T) {
	store, db := setup(t)
	_ = db.Migrator().DropTable(&models.User{})

	_, err := store.FindUser(ctx(), authentication.User{ID: 1})

	wantKind(t, err, apperr.KindInternal)
}

func TestNodeLifecycle(t *testing.T) {
	store, db := setup(t)
	node := &models.Node{Identifier: "n1", Name: "Mac"}
	if err := store.CreateNode(ctx(), node, "first"); err != nil || node.ID == 0 {
		t.Fatalf("create: %+v, %v", node, err)
	}
	if tokenCount(db, "node_id = ? AND token = ?", node.ID, "first") != 1 {
		t.Fatal("first token missing")
	}
	if err := store.CreateNode(ctx(), &models.Node{Identifier: "n1", Name: "Dup"}, "x"); err == nil {
		t.Fatal("duplicate node stored")
	}

	if err := store.ReplaceNodeToken(ctx(), node, "second"); err != nil {
		t.Fatal(err)
	}
	if tokenCount(db, "node_id = ?", node.ID) != 1 || tokenCount(db, "token = ?", "first") != 0 {
		t.Fatal("the old token still works")
	}
	if nodes, err := store.Nodes(ctx()); err != nil || len(nodes) != 1 {
		t.Fatalf("nodes %+v, %v", nodes, err)
	}
	if got, err := store.Node(ctx(), node.ID); err != nil || got.Identifier != "n1" {
		t.Fatalf("node %+v, %v", got, err)
	}
	if err := store.DeleteNode(ctx(), node.ID); err != nil || tokenCount(db, "node_id = ?", node.ID) != 0 {
		t.Fatalf("delete: %v", err)
	}
	_, err := store.Node(ctx(), node.ID)
	wantKind(t, err, apperr.KindNotFound)
	wantKind(t, store.DeleteNode(ctx(), node.ID), apperr.KindNotFound)
}

func TestDatabaseErrorsAreInternal(t *testing.T) {
	store := closed(t)
	node := &models.Node{Model: models.Model{ID: 1}}
	calls := map[string]error{
		"create token":    store.CreateToken(ctx(), &models.AccessToken{}),
		"project tokens":  func() error { _, err := store.ProjectTokens(ctx(), 1); return err }(),
		"delete project":  store.DeleteProjectToken(ctx(), 1, 1),
		"personal tokens": func() error { _, err := store.PersonalTokens(ctx(), 1); return err }(),
		"delete personal": store.DeletePersonalToken(ctx(), 1, 1),
		"users":           func() error { _, err := store.Users(ctx()); return err }(),
		"nodes":           func() error { _, err := store.Nodes(ctx()); return err }(),
		"node":            func() error { _, err := store.Node(ctx(), 1); return err }(),
		"create node":     store.CreateNode(ctx(), node, "t"),
		"replace token":   store.ReplaceNodeToken(ctx(), node, "t"),
		"delete node":     store.DeleteNode(ctx(), 1),
	}
	for name, err := range calls {
		if apperr.KindOf(err) != apperr.KindInternal {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestDeleteNodeFailsWhenTheTokenDeleteFails(t *testing.T) {
	store, db := setup(t)
	node := &models.Node{Identifier: "n", Name: "n"}
	db.Create(node)
	if err := db.Migrator().DropTable(&models.AccessToken{}); err != nil {
		t.Fatal(err)
	}
	wantKind(t, store.DeleteNode(ctx(), node.ID), apperr.KindInternal)
}

func TestCreateNodeFailsWhenTheTokenTableIsGone(t *testing.T) {
	store, db := setup(t)
	if err := db.Migrator().DropTable(&models.AccessToken{}); err != nil {
		t.Fatal(err)
	}
	wantKind(t, store.CreateNode(ctx(), &models.Node{Identifier: "n", Name: "n"}, "t"), apperr.KindInternal)
	var nodes int64
	db.Model(&models.Node{}).Count(&nodes)
	if nodes != 0 {
		t.Fatal("the node stayed although its token failed")
	}
}

// A token or a node that already exists is a conflict (409), not an internal
// error (500): the database translates a duplicate key.
func TestADuplicateKeyIsAConflict(t *testing.T) {
	store, _ := setup(t)
	if err := store.CreateToken(ctx(), &models.AccessToken{Name: "ci", Token: "same"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateNode(ctx(), &models.Node{Identifier: "n1", Name: "one"}, "t1"); err != nil {
		t.Fatal(err)
	}

	wantKind(t, store.CreateToken(ctx(), &models.AccessToken{Name: "again", Token: "same"}), apperr.KindConflict)
	wantKind(t, store.CreateNode(ctx(), &models.Node{Identifier: "n1", Name: "two"}, "t2"), apperr.KindConflict)
}
