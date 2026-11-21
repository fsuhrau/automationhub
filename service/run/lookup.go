package run

import (
	"context"

	"github.com/fsuhrau/automationhub/storage/models"
)

// How many log entries a page has by default and at most.
const (
	LogsDefault = 100
	LogsMax     = 500
)

// LogPage is a page of the log entries of a protocol and how many entries
// match in all.
type LogPage struct {
	Total   int64
	Offset  int
	Entries []models.ProtocolEntry
}

// Devices lists the devices the project can see, with their node.
func (s *Service) Devices(ctx context.Context, project *models.Project) ([]models.Device, error) {
	return s.deps.Repo.Devices(ctx, DeviceQuery{Project: project, All: true})
}

// Device returns one device of the project, with its node.
func (s *Service) Device(ctx context.Context, project *models.Project, id uint) (*models.Device, error) {
	return s.deps.Repo.Device(ctx, project, id)
}

// Protocol returns a protocol of the project with its device and keys.
func (s *Service) Protocol(ctx context.Context, projectID, protocolID uint) (*models.TestProtocol, error) {
	return s.deps.Repo.ProjectProtocol(ctx, projectID, protocolID)
}

// Logs pages the log entries of a protocol of the project. A Limit outside
// 1..LogsMax is LogsDefault or LogsMax.
func (s *Service) Logs(ctx context.Context, q LogQuery) (*LogPage, error) {
	if _, err := s.deps.Repo.ProjectProtocol(ctx, q.ProjectID, q.ProtocolID); err != nil {
		return nil, err
	}
	q.Limit = min(max(q.Limit, 0), LogsMax)
	if q.Limit == 0 {
		q.Limit = LogsDefault
	}
	entries, total, err := s.deps.Repo.ProtocolLogs(ctx, q)
	if err != nil {
		return nil, err
	}
	return &LogPage{Total: total, Offset: q.Offset, Entries: entries}, nil
}
