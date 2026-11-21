package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/service/project"
	"github.com/fsuhrau/automationhub/storage/models"
)

var errBoom = errors.New("boom")

type fakeRepo struct {
	projects []models.Project
	saved    *models.Project
	created  *models.Project
	base     string
	deleted  uint
	err      error
	saveErr  error
}

func (f *fakeRepo) List(context.Context) ([]models.Project, error) { return f.projects, f.err }

func (f *fakeRepo) CreateUnique(_ context.Context, p *models.Project, base string) error {
	f.created, f.base = p, base
	return f.err
}

func (f *fakeRepo) ByIdentifier(_ context.Context, identifier string) (*models.Project, error) {
	if f.err != nil {
		return nil, f.err
	}
	for i := range f.projects {
		if f.projects[i].Identifier == identifier {
			p := f.projects[i]
			return &p, nil
		}
	}
	return nil, apperr.NotFound("project %s not found", identifier)
}

func (f *fakeRepo) Save(_ context.Context, p *models.Project) error {
	f.saved = p
	return f.saveErr
}

func (f *fakeRepo) Delete(_ context.Context, id uint) error {
	f.deleted = id
	return f.err
}

func newService(repo *fakeRepo) *project.Service { return project.New(project.Deps{Repo: repo}) }

func TestList(t *testing.T) {
	repo := &fakeRepo{projects: []models.Project{{Name: "a"}}}
	got, err := newService(repo).List(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("got %v, %v", got, err)
	}
	repo.err = errBoom
	if _, err := newService(repo).List(context.Background()); !errors.Is(err, errBoom) {
		t.Fatalf("err %v", err)
	}
}

func TestCreate(t *testing.T) {
	repo := &fakeRepo{}
	p, err := newService(repo).Create(context.Background(), "  Heroes Game ")
	if err != nil || p.Name != "Heroes Game" || repo.base != "heroes-game" || repo.created != p {
		t.Fatalf("got %+v %q, %v", p, repo.base, err)
	}
}

func TestCreateErrors(t *testing.T) {
	if _, err := newService(&fakeRepo{}).Create(context.Background(), " "); apperr.KindOf(err) != apperr.KindInvalid {
		t.Errorf("blank name: %v", err)
	}
	if _, err := newService(&fakeRepo{err: errBoom}).Create(context.Background(), "x"); !errors.Is(err, errBoom) {
		t.Errorf("repo error: %v", err)
	}
}

func TestRename(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		change     project.Change
		wantName   string
		wantID     string
	}{
		{"name only", "proj", project.Change{Name: " New ", Identifier: "other"}, "New", "proj"},
		{"default project takes the identifier", "default_project", project.Change{Name: "New", Identifier: "main"}, "New", "main"},
		{"default project keeps it without one", "default_project", project.Change{Name: "New"}, "New", "default_project"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{projects: []models.Project{{Identifier: tc.identifier, Name: "Old"}}}
			p, err := newService(repo).Rename(context.Background(), tc.identifier, tc.change)
			if err != nil || p.Name != tc.wantName || p.Identifier != tc.wantID || repo.saved != p {
				t.Fatalf("got %+v, %v", p, err)
			}
		})
	}
}

func TestRenameErrors(t *testing.T) {
	existing := []models.Project{{Identifier: "proj"}}
	tests := []struct {
		name string
		repo *fakeRepo
		id   string
		in   string
		kind apperr.Kind
	}{
		{"blank name", &fakeRepo{projects: existing}, "proj", " ", apperr.KindInvalid},
		{"unknown", &fakeRepo{projects: existing}, "nope", "x", apperr.KindNotFound},
		{"save fails", &fakeRepo{projects: existing, saveErr: apperr.Conflict("taken")}, "proj", "x", apperr.KindConflict},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newService(tc.repo).Rename(context.Background(), tc.id, project.Change{Name: tc.in})
			if apperr.KindOf(err) != tc.kind {
				t.Fatalf("err %v", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	repo := &fakeRepo{}
	if err := newService(repo).Delete(context.Background(), 7); err != nil || repo.deleted != 7 {
		t.Fatalf("deleted %d, %v", repo.deleted, err)
	}
	repo.err = apperr.NotFound("project 7 not found")
	if err := newService(repo).Delete(context.Background(), 7); apperr.KindOf(err) != apperr.KindNotFound {
		t.Fatalf("err %v", err)
	}
}
