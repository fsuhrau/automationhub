package app

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// ParameterInput is what a request may set on an app parameter.
type ParameterInput struct {
	Name string
	// Type is the JSON of the parameter's type.
	Type string
}

// parameterType is the JSON the parameter's type is stored as.
type parameterType struct {
	Type         string   `json:"type"`
	DefaultValue string   `json:"defaultValue"`
	Options      []string `json:"options,omitempty"`
}

// AddParameter adds a parameter to the app.
func (s *Service) AddParameter(ctx context.Context, sc Scope, in ParameterInput) (*models.AppParameter, error) {
	if _, err := s.repo.App(ctx, sc); err != nil {
		return nil, err
	}
	tidy, err := s.validate(ctx, ParameterRef{Scope: sc}, in)
	if err != nil {
		return nil, err
	}
	param := models.AppParameter{AppID: sc.AppID, Name: tidy.Name, Type: tidy.Type}
	if err := s.repo.CreateParameter(ctx, &param); err != nil {
		return nil, err
	}
	return &param, nil
}

// UpdateParameter changes the name and type of a parameter of the app.
func (s *Service) UpdateParameter(ctx context.Context, ref ParameterRef, in ParameterInput) (*models.AppParameter, error) {
	param, err := s.repo.Parameter(ctx, ref)
	if err != nil {
		return nil, err
	}
	tidy, err := s.validate(ctx, ref, in)
	if err != nil {
		return nil, err
	}
	param.Name, param.Type = tidy.Name, tidy.Type
	if err := s.repo.SaveParameter(ctx, param); err != nil {
		return nil, err
	}
	return param, nil
}

// RemoveParameter deletes a parameter of the app; one that is not there is
// not an error.
func (s *Service) RemoveParameter(ctx context.Context, ref ParameterRef) error {
	return s.repo.DeleteParameter(ctx, ref)
}

// validate checks the input against the other parameters of the app, the one
// the ref names excluded (ID 0 names none), and returns it with the tidied
// name and type.
func (s *Service) validate(ctx context.Context, ref ParameterRef, in ParameterInput) (ParameterInput, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return in, apperr.Invalid("the parameter needs a name")
	}
	others, err := s.repo.Parameters(ctx, ref.Scope)
	if err != nil {
		return in, err
	}
	if clash := nameClash(others, ref.ID, name); clash != "" {
		return in, apperr.Invalid("the app has a parameter %s already", clash)
	}
	tidy, err := tidyType(in.Type)
	return ParameterInput{Name: name, Type: tidy}, err
}

// nameClash is the name of the parameter, other than the one with the id,
// that has the name ignoring case, or "".
func nameClash(params []models.AppParameter, id uint, name string) string {
	for _, p := range params {
		if p.ID != id && strings.EqualFold(p.Name, name) {
			return p.Name
		}
	}
	return ""
}

// tidyType checks the type of a parameter, text or a choice with options and
// a default among them, and returns it as the JSON to store.
func tidyType(raw string) (string, error) {
	var t parameterType
	if err := json.Unmarshal([]byte(raw), &t); err != nil {
		return "", apperr.Invalid("the parameter's type isn't valid")
	}
	switch t.Type {
	case "string":
		t.Options = nil
	case "option":
		if err := tidyOptions(&t); err != nil {
			return "", err
		}
	default:
		return "", apperr.Invalid("unknown parameter type %q", t.Type)
	}
	tidy, _ := json.Marshal(t) // a struct of strings always encodes
	return string(tidy), nil
}

func tidyOptions(t *parameterType) error {
	seen := map[string]bool{}
	var options []string
	for _, o := range t.Options {
		o = strings.TrimSpace(o)
		if o != "" && !seen[o] {
			seen[o] = true
			options = append(options, o)
		}
	}
	if len(options) == 0 {
		return apperr.Invalid("a choice needs at least one option")
	}
	if t.DefaultValue != "" && !seen[t.DefaultValue] {
		return apperr.Invalid("the default %s isn't one of the options", t.DefaultValue)
	}
	t.Options = options
	return nil
}
