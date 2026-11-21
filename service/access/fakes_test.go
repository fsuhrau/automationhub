package access_test

import (
	"context"
	"errors"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/hub/manager"
	"github.com/fsuhrau/automationhub/service/access"
	"github.com/fsuhrau/automationhub/storage/models"
)

var errBoom = errors.New("boom")

// fakeRepo keeps everything in memory. fail names the method that returns
// errBoom; project scope and user scope work like the real store's.
type fakeRepo struct {
	users  []models.User
	tokens []models.AccessToken
	nodes  []models.Node
	fail   string
	// replaced is the last token a node got
	replaced string
}

func (f *fakeRepo) err(method string) error {
	if f.fail == method {
		return errBoom
	}
	return nil
}

func (f *fakeRepo) Users(context.Context) ([]models.User, error) { return f.users, f.err("Users") }

func (f *fakeRepo) FindUser(_ context.Context, who authentication.User) (*models.User, error) {
	if err := f.err("FindUser"); err != nil {
		return nil, err
	}
	for i := range f.users {
		if f.users[i].ID == who.ID {
			return &f.users[i], nil
		}
	}
	return nil, apperr.NotFound("user not found")
}

func (f *fakeRepo) CreateToken(_ context.Context, t *models.AccessToken) error {
	t.ID = uint(len(f.tokens) + 1)
	f.tokens = append(f.tokens, *t)
	return f.err("CreateToken")
}

func (f *fakeRepo) ProjectTokens(_ context.Context, projectID uint) ([]models.AccessToken, error) {
	var out []models.AccessToken
	for _, t := range f.tokens {
		if t.ProjectID != nil && *t.ProjectID == projectID {
			out = append(out, t)
		}
	}
	return out, f.err("ProjectTokens")
}

func (f *fakeRepo) DeleteProjectToken(_ context.Context, projectID, tokenID uint) error {
	for i, t := range f.tokens {
		if t.ID == tokenID && t.ProjectID != nil && *t.ProjectID == projectID {
			f.tokens = append(f.tokens[:i], f.tokens[i+1:]...)
			return f.err("DeleteProjectToken")
		}
	}
	return apperr.NotFound("token %d not found", tokenID)
}

func (f *fakeRepo) PersonalTokens(_ context.Context, userID uint) ([]models.AccessToken, error) {
	var out []models.AccessToken
	for _, t := range f.tokens {
		if t.UserID != nil && *t.UserID == userID {
			out = append(out, t)
		}
	}
	return out, f.err("PersonalTokens")
}

func (f *fakeRepo) DeletePersonalToken(_ context.Context, userID, tokenID uint) error {
	for i, t := range f.tokens {
		if t.ID == tokenID && t.UserID != nil && *t.UserID == userID {
			f.tokens = append(f.tokens[:i], f.tokens[i+1:]...)
			return f.err("DeletePersonalToken")
		}
	}
	return apperr.NotFound("token not found")
}

func (f *fakeRepo) Nodes(context.Context) ([]models.Node, error) {
	return append([]models.Node(nil), f.nodes...), f.err("Nodes")
}

func (f *fakeRepo) Node(_ context.Context, id uint) (*models.Node, error) {
	if err := f.err("Node"); err != nil {
		return nil, err
	}
	for i := range f.nodes {
		if f.nodes[i].ID == id {
			n := f.nodes[i]
			return &n, nil
		}
	}
	return nil, apperr.NotFound("node %d not found", id)
}

func (f *fakeRepo) CreateNode(_ context.Context, n *models.Node, token string) error {
	n.ID = uint(len(f.nodes) + 1)
	f.nodes = append(f.nodes, *n)
	f.replaced = token
	return f.err("CreateNode")
}

func (f *fakeRepo) ReplaceNodeToken(_ context.Context, _ *models.Node, token string) error {
	f.replaced = token
	return f.err("ReplaceNodeToken")
}

func (f *fakeRepo) DeleteNode(_ context.Context, id uint) error {
	for i, n := range f.nodes {
		if n.ID == id {
			f.nodes = append(f.nodes[:i], f.nodes[i+1:]...)
			return f.err("DeleteNode")
		}
	}
	return apperr.NotFound("node %d not found", id)
}

// fakeStatus answers every node with a fixed status.
type fakeStatus struct{ status int }

func (f fakeStatus) GetStatus(manager.NodeIdentifier) (int, error) { return f.status, nil }

func now() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) }

// counter hands out uuid-1, uuid-2, ... and can fail on the nth call.
type counter struct{ n, failAt int }

func (c *counter) next() (string, error) {
	c.n++
	if c.n == c.failAt {
		return "", errBoom
	}
	return "uuid-" + string(rune('0'+c.n)), nil
}

func newService(repo *fakeRepo, ids *counter) *access.Service {
	return access.New(access.Deps{
		Repo:    repo,
		Status:  fakeStatus{status: 2},
		NewUUID: ids.next,
		NewPersonal: func() (string, string, string, error) {
			return "ahp_plain", "sha256:hash", "lain", nil
		},
		Now: now,
	})
}
