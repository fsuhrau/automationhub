package access_test

import (
	"testing"
	"time"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/service/access"
	"github.com/fsuhrau/automationhub/storage/models"
)

func alice() authentication.User { return authentication.User{ID: 1, Name: "Alice"} }

func personalRepo() *fakeRepo {
	return &fakeRepo{
		users: []models.User{{Model: models.Model{ID: 1}, Name: "Alice"}, {Model: models.Model{ID: 2}, Name: "Bob"}},
		tokens: []models.AccessToken{
			{Model: models.Model{ID: 1}, UserID: ptr(uint(1)), Name: "mine"},
			{Model: models.Model{ID: 2}, UserID: ptr(uint(2)), Name: "bobs"},
			{Model: models.Model{ID: 3}, ProjectID: ptr(uint(1))},
		},
	}
}

func TestMe(t *testing.T) {
	svc := newService(personalRepo(), &counter{})
	if u, err := svc.Me(ctx(), alice()); err != nil || u.ID != 1 {
		t.Fatalf("got %+v, %v", u, err)
	}
	_, err := svc.Me(ctx(), authentication.User{})
	wantKind(t, err, apperr.KindNotFound)
	if err.Error() != noUserMessage {
		t.Fatalf("message %q", err)
	}
}

const noUserMessage = "no signed in user; personal access tokens need a login (configure authentication on the master)"

func TestMeStoreError(t *testing.T) {
	repo := personalRepo()
	repo.fail = "FindUser"
	_, err := newService(repo, &counter{}).Me(ctx(), alice())
	wantKind(t, err, apperr.KindInternal)
}

func TestPersonalTokensListOnlyTheUsersOwn(t *testing.T) {
	svc := newService(personalRepo(), &counter{})
	got, err := svc.PersonalTokens(ctx(), alice())
	if err != nil || len(got) != 1 || got[0].Name != "mine" {
		t.Fatalf("got %+v, %v", got, err)
	}
	_, err = svc.PersonalTokens(ctx(), authentication.User{})
	wantKind(t, err, apperr.KindNotFound)
	repo := personalRepo()
	repo.fail = "PersonalTokens"
	_, err = newService(repo, &counter{}).PersonalTokens(ctx(), alice())
	wantKind(t, err, apperr.KindInternal)
}

func TestCreatePersonalToken(t *testing.T) {
	repo := personalRepo()
	expires := now().Add(time.Hour)
	got, err := newService(repo, &counter{}).CreatePersonalToken(ctx(), alice(), access.NewToken{Name: " laptop ", ExpiresAt: &expires})
	if err != nil || got.Plain != "ahp_plain" || got.Token != "sha256:hash" || got.Hint != "lain" || got.Name != "laptop" || *got.UserID != 1 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestCreatePersonalTokenErrors(t *testing.T) {
	past := now().Add(-time.Second)
	tests := []struct {
		name string
		who  authentication.User
		in   access.NewToken
		fail string
		kind apperr.Kind
	}{
		{"no user", authentication.User{}, access.NewToken{Name: "x"}, "", apperr.KindNotFound},
		{"blank name", alice(), access.NewToken{Name: " "}, "", apperr.KindInvalid},
		{"expired", alice(), access.NewToken{Name: "x", ExpiresAt: &past}, "", apperr.KindInvalid},
		{"store fails", alice(), access.NewToken{Name: "x"}, "CreateToken", apperr.KindInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := personalRepo()
			repo.fail = tc.fail
			_, err := newService(repo, &counter{}).CreatePersonalToken(ctx(), tc.who, tc.in)
			wantKind(t, err, tc.kind)
		})
	}
}

func TestCreatePersonalTokenGeneratorFails(t *testing.T) {
	svc := access.New(access.Deps{
		Repo:        personalRepo(),
		NewPersonal: func() (string, string, string, error) { return "", "", "", errBoom },
	})
	_, err := svc.CreatePersonalToken(ctx(), alice(), access.NewToken{Name: "x"})
	wantKind(t, err, apperr.KindInternal)
}

// Only the owner can revoke a personal token.
func TestDeletePersonalToken(t *testing.T) {
	repo := personalRepo()
	svc := newService(repo, &counter{})
	wantKind(t, svc.DeletePersonalToken(ctx(), alice(), 2), apperr.KindNotFound)
	wantKind(t, svc.DeletePersonalToken(ctx(), alice(), 3), apperr.KindNotFound)
	wantKind(t, svc.DeletePersonalToken(ctx(), authentication.User{}, 1), apperr.KindNotFound)
	if err := svc.DeletePersonalToken(ctx(), alice(), 1); err != nil || len(repo.tokens) != 2 {
		t.Fatalf("err %v, %d tokens left", err, len(repo.tokens))
	}
}
