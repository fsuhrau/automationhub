package appstore

import (
	"context"
	"fmt"

	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/fsuhrau/automationhub/storage/storeerr"
)

// Parameters lists the parameters of the app.
func (s *Store) Parameters(ctx context.Context, sc appsvc.Scope) ([]models.AppParameter, error) {
	var found []models.AppParameter
	err := inScope(s.tx(ctx), "app_id", sc).Find(&found).Error
	return found, storeerr.Map(err, "parameters")
}

// Parameter loads one parameter of the app.
func (s *Store) Parameter(ctx context.Context, ref appsvc.ParameterRef) (*models.AppParameter, error) {
	var found models.AppParameter
	err := inScope(s.tx(ctx), "app_id", ref.Scope).First(&found, "id = ?", ref.ID).Error
	return &found, lookup(err, fmt.Sprintf("parameter %d", ref.ID))
}

// CreateParameter inserts the parameter.
func (s *Store) CreateParameter(ctx context.Context, p *models.AppParameter) error {
	return storeerr.Map(s.tx(ctx).Create(p).Error, "parameter")
}

// SaveParameter writes the parameter.
func (s *Store) SaveParameter(ctx context.Context, p *models.AppParameter) error {
	return storeerr.Map(s.tx(ctx).Save(p).Error, fmt.Sprintf("parameter %d", p.ID))
}

// DeleteParameter removes the parameter if the scope has it.
func (s *Store) DeleteParameter(ctx context.Context, ref appsvc.ParameterRef) error {
	err := inScope(s.tx(ctx), "app_id", ref.Scope).Delete(&models.AppParameter{}, "id = ?", ref.ID).Error
	return storeerr.Map(err, fmt.Sprintf("parameter %d", ref.ID))
}
