package app

import (
	"context"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

const optionType = `{"type":"option","defaultValue":"de","options":[" de","en","de",""]}`

func TestAddParameter(t *testing.T) {
	f := newFixture(t)
	got, err := f.svc.AddParameter(context.Background(), f.sc, ParameterInput{Name: " Market ", Type: optionType})
	if err != nil || got.AppID != 1 || got.Name != "Market" || !strings.Contains(got.Type, `"options":["de","en"]`) {
		t.Fatalf("added %+v, %v", got, err)
	}
	if len(f.repo.params) != 1 {
		t.Errorf("%d parameters stored, want 1", len(f.repo.params))
	}
}

func TestParameterValidation(t *testing.T) {
	tests := []struct {
		name string
		in   ParameterInput
	}{
		{"no name", ParameterInput{Name: " ", Type: `{"type":"string"}`}},
		{"taken name, ignoring case", ParameterInput{Name: "market", Type: `{"type":"string"}`}},
		{"type is no json", ParameterInput{Name: "X", Type: "string"}},
		{"no options", ParameterInput{Name: "X", Type: `{"type":"option","options":[" "]}`}},
		{"default not an option", ParameterInput{Name: "X", Type: `{"type":"option","defaultValue":"fr","options":["de"]}`}},
		{"unknown type", ParameterInput{Name: "X", Type: `{"type":"number"}`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.repo.params = []models.AppParameter{{Model: models.Model{ID: 7}, AppID: 1, Name: "Market"}}
			_, err := f.svc.AddParameter(context.Background(), f.sc, tt.in)
			wantKind(t, err, apperr.KindInvalid)
		})
	}
}

func TestStringParameterDropsOptions(t *testing.T) {
	f := newFixture(t)
	got, err := f.svc.AddParameter(context.Background(), f.sc, ParameterInput{Name: "P", Type: `{"type":"string","defaultValue":"x","options":["a"]}`})
	if err != nil || strings.Contains(got.Type, "options") {
		t.Fatalf("type %q, %v; want no options", got.Type, err)
	}
}

func TestAddParameterErrors(t *testing.T) {
	f := newFixture(t)
	in := ParameterInput{Name: "P", Type: `{"type":"string"}`}
	_, err := f.svc.AddParameter(context.Background(), Scope{ProjectID: 2, AppID: 1}, in)
	wantKind(t, err, apperr.KindNotFound)
	for _, method := range []string{"Parameters", "CreateParameter"} {
		f.repo.fail[method] = errBoom
		if _, err := f.svc.AddParameter(context.Background(), f.sc, in); err == nil {
			t.Errorf("%s: a store error must come back", method)
		}
		delete(f.repo.fail, method)
	}
}

func TestUpdateParameter(t *testing.T) {
	f := newFixture(t)
	f.repo.params = []models.AppParameter{
		{Model: models.Model{ID: 7}, AppID: 1, Name: "Market"},
		{Model: models.Model{ID: 8}, AppID: 1, Name: "Proxy"},
	}
	ref := ParameterRef{Scope: f.sc, ID: 7}
	got, err := f.svc.UpdateParameter(context.Background(), ref, ParameterInput{Name: "Market", Type: optionType})
	if err != nil || got.ID != 7 || !strings.Contains(got.Type, `"defaultValue":"de"`) {
		t.Fatalf("updated %+v, %v; an update keeps its own name", got, err)
	}
	if f.repo.params[0].Type != got.Type {
		t.Error("the change was not saved")
	}
	_, err = f.svc.UpdateParameter(context.Background(), ref, ParameterInput{Name: "Proxy", Type: `{"type":"string"}`})
	wantKind(t, err, apperr.KindInvalid)
}

func TestUpdateParameterErrors(t *testing.T) {
	f := newFixture(t)
	f.repo.params = []models.AppParameter{{Model: models.Model{ID: 7}, AppID: 1, Name: "Market"}}
	in := ParameterInput{Name: "M", Type: `{"type":"string"}`}
	for _, ref := range []ParameterRef{{Scope: f.sc, ID: 99}, {Scope: Scope{ProjectID: 2, AppID: 1}, ID: 7}} {
		_, err := f.svc.UpdateParameter(context.Background(), ref, in)
		wantKind(t, err, apperr.KindNotFound)
	}
	f.repo.fail["SaveParameter"] = errBoom
	if _, err := f.svc.UpdateParameter(context.Background(), ParameterRef{Scope: f.sc, ID: 7}, in); err == nil {
		t.Fatal("a store error must come back")
	}
}

func TestRemoveParameter(t *testing.T) {
	f := newFixture(t)
	f.repo.params = []models.AppParameter{{Model: models.Model{ID: 7}, AppID: 1, Name: "Market"}}
	ctx := context.Background()
	if err := f.svc.RemoveParameter(ctx, ParameterRef{Scope: Scope{ProjectID: 2, AppID: 1}, ID: 7}); err != nil || len(f.repo.params) != 1 {
		t.Fatalf("another project's remove: %v, %d parameters, want 1", err, len(f.repo.params))
	}
	if err := f.svc.RemoveParameter(ctx, ParameterRef{Scope: f.sc, ID: 99}); err != nil {
		t.Fatalf("an unknown parameter is not an error: %v", err)
	}
	if err := f.svc.RemoveParameter(ctx, ParameterRef{Scope: f.sc, ID: 7}); err != nil || len(f.repo.params) != 0 {
		t.Fatalf("remove: %v, %d parameters, want 0", err, len(f.repo.params))
	}
}
