package device

import (
	"context"
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Param is a custom parameter of a device; ID is 0 for a new one.
type Param struct {
	ID    uint   `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Update is what the device edit page changes: the alias and the custom
// parameters of the device with DeviceID. Everything else is reported by the
// device or its node.
type Update struct {
	DeviceID        string  `json:"-"`
	Alias           string  `json:"alias"`
	CustomParameter []Param `json:"customParameter"`
}

// Change is an Update turned into the writes the storage has to make.
type Change struct {
	Alias  string
	Update []Param
	Add    []Param
	Remove []uint
}

// Update sets the alias (an empty one clears it) and replaces the custom
// parameters: submitted ones with the id of an existing parameter are
// updated, others are added, and existing ones not submitted are removed.
// Keys must be set and unique.
func (s *Service) Update(ctx context.Context, project *models.Project, req Update) (*models.Device, error) {
	req, err := req.normalized()
	if err != nil {
		return nil, err
	}
	d, err := s.deps.Repo.FindDetail(ctx, project, req.DeviceID)
	if err != nil {
		return nil, err
	}
	if err := s.deps.Repo.Apply(ctx, d.ID, planChange(d.CustomParameter, req)); err != nil {
		return nil, err
	}
	return s.deps.Repo.FindDetail(ctx, project, req.DeviceID)
}

// normalized trims the alias and the keys and rejects empty and repeated
// keys. The receiver is not changed.
func (u Update) normalized() (Update, error) {
	out := Update{DeviceID: u.DeviceID, Alias: strings.TrimSpace(u.Alias), CustomParameter: make([]Param, len(u.CustomParameter))}
	seen := map[string]bool{}
	for i, p := range u.CustomParameter {
		p.Key = strings.TrimSpace(p.Key)
		if p.Key == "" {
			return Update{}, apperr.Invalid("custom parameter %d has no key", i+1)
		}
		if seen[p.Key] {
			return Update{}, apperr.Invalid("custom parameter %q is there twice", p.Key)
		}
		seen[p.Key] = true
		out.CustomParameter[i] = p
	}
	return out, nil
}

// planChange decides what happens to each parameter. An id of a parameter
// the device doesn't have (another device's) is treated as new.
func planChange(existing []models.CustomParameter, req Update) Change {
	known := map[uint]bool{}
	for _, p := range existing {
		known[p.ID] = true
	}
	change := Change{Alias: req.Alias}
	kept := map[uint]bool{}
	for _, p := range req.CustomParameter {
		if p.ID != 0 && known[p.ID] {
			kept[p.ID] = true
			change.Update = append(change.Update, p)
			continue
		}
		change.Add = append(change.Add, p)
	}
	for _, p := range existing {
		if !kept[p.ID] {
			change.Remove = append(change.Remove, p.ID)
		}
	}
	return change
}
