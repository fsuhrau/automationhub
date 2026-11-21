package access_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/service/access"
	"github.com/fsuhrau/automationhub/storage/models"
)

func ctx() context.Context { return context.Background() }

func ptr[T any](v T) *T { return &v }

func wantKind(t *testing.T, err error, kind apperr.Kind) {
	t.Helper()
	if apperr.KindOf(err) != kind {
		t.Fatalf("err %v has kind %v, want %v", err, apperr.KindOf(err), kind)
	}
}

func TestUsers(t *testing.T) {
	repo := &fakeRepo{users: []models.User{{Name: "Alice"}}}
	got, err := newService(repo, &counter{}).Users(ctx())
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v, %v", got, err)
	}
	repo.fail = "Users"
	if _, err := newService(repo, &counter{}).Users(ctx()); !errors.Is(err, errBoom) {
		t.Fatalf("err %v", err)
	}
}

func TestNewUsesTheRealGeneratorsByDefault(t *testing.T) {
	repo := &fakeRepo{}
	svc := access.New(access.Deps{Repo: repo})
	created, err := svc.CreateProjectToken(ctx(), 1, access.NewToken{Name: "ci"})
	if err != nil || len(created.Token) != 36 {
		t.Fatalf("project token %+v, %v", created, err)
	}
	repo.users = []models.User{{Model: models.Model{ID: 1}, Name: "Alice"}}
	future := time.Now().Add(time.Hour)
	personal, err := svc.CreatePersonalToken(ctx(), authentication.User{Name: "Alice"}, access.NewToken{Name: "x", ExpiresAt: &future})
	if err != nil || !strings.HasPrefix(personal.Plain, "ahp_") {
		t.Fatalf("personal token %+v, %v", personal, err)
	}
}

func TestProjectTokens(t *testing.T) {
	repo := &fakeRepo{}
	svc := newService(repo, &counter{})
	expires := now().Add(time.Hour)
	created, err := svc.CreateProjectToken(ctx(), 1, access.NewToken{Name: " ci ", ExpiresAt: &expires})
	if err != nil || created.Name != "ci" || created.Token != "uuid-1" || *created.ProjectID != 1 || created.ExpiresAt != &expires {
		t.Fatalf("created %+v, %v", created, err)
	}
	if _, err := svc.CreateProjectToken(ctx(), 2, access.NewToken{Name: "other"}); err != nil {
		t.Fatal(err)
	}
	list, err := svc.ProjectTokens(ctx(), 1)
	if err != nil || len(list) != 1 || list[0].Name != "ci" {
		t.Fatalf("list %+v, %v", list, err)
	}
}

func TestCreateProjectTokenErrors(t *testing.T) {
	tests := []struct {
		name string
		repo *fakeRepo
		ids  *counter
		in   string
		kind apperr.Kind
	}{
		{"blank name", &fakeRepo{}, &counter{}, "  ", apperr.KindInvalid},
		{"generator fails", &fakeRepo{}, &counter{failAt: 1}, "x", apperr.KindInternal},
		{"store fails", &fakeRepo{fail: "CreateToken"}, &counter{}, "x", apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newService(tc.repo, tc.ids).CreateProjectToken(ctx(), 1, access.NewToken{Name: tc.in})
			wantKind(t, err, tc.kind)
		})
	}
}

func TestProjectTokensStoreError(t *testing.T) {
	_, err := newService(&fakeRepo{fail: "ProjectTokens"}, &counter{}).ProjectTokens(ctx(), 1)
	wantKind(t, err, apperr.KindInternal)
}

// A project can only revoke its own tokens: those of other projects,
// nodes and users are not found.
func TestDeleteProjectTokenIsScopedToTheProject(t *testing.T) {
	repo := &fakeRepo{tokens: []models.AccessToken{
		{Model: models.Model{ID: 1}, ProjectID: ptr(uint(1))},
		{Model: models.Model{ID: 2}, ProjectID: ptr(uint(2))},
		{Model: models.Model{ID: 3}, UserID: ptr(uint(1))},
		{Model: models.Model{ID: 4}, NodeID: ptr(uint(1))},
	}}
	svc := newService(repo, &counter{})
	for _, id := range []uint{2, 3, 4, 99} {
		wantKind(t, svc.DeleteProjectToken(ctx(), 1, id), apperr.KindNotFound)
	}
	if len(repo.tokens) != 4 {
		t.Fatalf("a foreign token was deleted: %+v", repo.tokens)
	}
	if err := svc.DeleteProjectToken(ctx(), 1, 1); err != nil || len(repo.tokens) != 3 {
		t.Fatalf("own token: %v, %d left", err, len(repo.tokens))
	}
}

func TestCreateNode(t *testing.T) {
	repo := &fakeRepo{}
	created, err := newService(repo, &counter{}).CreateNode(ctx(), " Mac mini ")
	if err != nil || created.Node.Name != "Mac mini" || created.Node.Identifier != "uuid-1" || created.Token != "uuid-2" || repo.replaced != "uuid-2" {
		t.Fatalf("created %+v, %v", created, err)
	}
}

func TestCreateNodeErrors(t *testing.T) {
	tests := []struct {
		name string
		repo *fakeRepo
		ids  *counter
		in   string
		kind apperr.Kind
	}{
		{"blank name", &fakeRepo{}, &counter{}, "", apperr.KindInvalid},
		{"identifier fails", &fakeRepo{}, &counter{failAt: 1}, "x", apperr.KindInternal},
		{"token fails", &fakeRepo{}, &counter{failAt: 2}, "x", apperr.KindInternal},
		{"store fails", &fakeRepo{fail: "CreateNode"}, &counter{}, "x", apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newService(tc.repo, tc.ids).CreateNode(ctx(), tc.in)
			wantKind(t, err, tc.kind)
		})
	}
}

func TestNodesCarryTheirStatus(t *testing.T) {
	repo := &fakeRepo{nodes: []models.Node{{Identifier: "a"}, {Identifier: "b"}}}
	svc := newService(repo, &counter{})
	got, err := svc.Nodes(ctx())
	if err != nil || len(got) != 2 || got[0].Status != 2 || got[1].Status != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	repo.fail = "Nodes"
	_, err = svc.Nodes(ctx())
	wantKind(t, err, apperr.KindInternal)
}

func TestNode(t *testing.T) {
	repo := &fakeRepo{nodes: []models.Node{{Model: models.Model{ID: 1}, Name: "n"}}}
	svc := newService(repo, &counter{})
	if n, err := svc.Node(ctx(), 1); err != nil || n.Name != "n" {
		t.Fatalf("got %+v, %v", n, err)
	}
	_, err := svc.Node(ctx(), 2)
	wantKind(t, err, apperr.KindNotFound)
}

func TestIssueNodeToken(t *testing.T) {
	repo := &fakeRepo{nodes: []models.Node{{Model: models.Model{ID: 1}, Name: "n"}}}
	got, err := newService(repo, &counter{}).IssueNodeToken(ctx(), 1)
	if err != nil || got.Token != "uuid-1" || got.Node.Name != "n" || repo.replaced != "uuid-1" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestIssueNodeTokenErrors(t *testing.T) {
	nodes := []models.Node{{Model: models.Model{ID: 1}}}
	tests := []struct {
		name string
		repo *fakeRepo
		ids  *counter
		id   uint
		kind apperr.Kind
	}{
		{"unknown node", &fakeRepo{nodes: nodes}, &counter{}, 9, apperr.KindNotFound},
		{"generator fails", &fakeRepo{nodes: nodes}, &counter{failAt: 1}, 1, apperr.KindInternal},
		{"store fails", &fakeRepo{nodes: nodes, fail: "ReplaceNodeToken"}, &counter{}, 1, apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newService(tc.repo, tc.ids).IssueNodeToken(ctx(), tc.id)
			wantKind(t, err, tc.kind)
		})
	}
}

func TestDeleteNode(t *testing.T) {
	repo := &fakeRepo{nodes: []models.Node{{Model: models.Model{ID: 1}}}}
	svc := newService(repo, &counter{})
	wantKind(t, svc.DeleteNode(ctx(), 2), apperr.KindNotFound)
	if err := svc.DeleteNode(ctx(), 1); err != nil || len(repo.nodes) != 0 {
		t.Fatalf("err %v, %d nodes left", err, len(repo.nodes))
	}
}
