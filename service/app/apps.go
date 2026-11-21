package app

import (
	"context"
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// AppInput is what a request may set on an app.
type AppInput struct {
	Name       string              `json:"name"`
	Identifier string              `json:"identifier"`
	Platform   models.PlatformType `json:"platform"`
}

// defaultIdentifier is the one an app has until it is changed once.
const defaultIdentifier = "default_app"

// Apps lists the apps of the project with their parameters.
func (s *Service) Apps(ctx context.Context, projectID uint) ([]models.App, error) {
	return s.repo.Apps(ctx, projectID)
}

// CreateApp adds an app to the project; it needs a name, a platform an app
// can target, and an identifier for the platforms that install bundles.
func (s *Service) CreateApp(ctx context.Context, projectID uint, in AppInput) (*models.App, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Identifier = strings.TrimSpace(in.Identifier)
	if err := validateApp(in); err != nil {
		return nil, err
	}
	created := models.App{ProjectID: projectID, Name: in.Name, Identifier: in.Identifier, Platform: in.Platform}
	if err := s.repo.CreateApp(ctx, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// UpdateApp renames the app. Its identifier can be changed once, while it is
// still the generated default one.
func (s *Service) UpdateApp(ctx context.Context, sc Scope, in AppInput) (*models.App, error) {
	current, err := s.repo.App(ctx, sc)
	if err != nil {
		return nil, err
	}
	current.Name = in.Name
	if current.Identifier == defaultIdentifier {
		current.Identifier = in.Identifier
	}
	if err := s.repo.SaveApp(ctx, current); err != nil {
		return nil, err
	}
	return current, nil
}

func validateApp(in AppInput) error {
	if in.Name == "" {
		return apperr.Invalid("the app needs a name")
	}
	switch in.Platform {
	case models.PlatformTypeAndroid, models.PlatformTypeiOS, models.PlatformTypeMac:
		if in.Identifier == "" {
			return apperr.Invalid("the app needs its bundle identifier or package name")
		}
	case models.PlatformTypeWindows, models.PlatformTypeLinux, models.PlatformTypeWeb, models.PlatformTypeEditor:
	default:
		return apperr.Invalid("unknown platform %d", in.Platform)
	}
	return nil
}
