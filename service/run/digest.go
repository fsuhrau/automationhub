package run

import (
	"context"

	"github.com/fsuhrau/automationhub/storage/models"
)

// DigestErrors is how many error entries a protocol digest lists.
const DigestErrors = 50

// Performance sums up the performance samples of a protocol.
type Performance struct {
	Samples int64
	AvgFPS  float64
	MinFPS  float64
	AvgMEM  float64
	MaxMEM  float64
	AvgCPU  float64
}

// Digest is what a protocol holds in short: how many entries it has, the
// first errors, its performance, its child protocols and the entries that
// refer to data such as screenshots.
type Digest struct {
	Entries     int64
	Errors      []models.ProtocolEntry
	Performance Performance
	Children    []models.TestProtocol
	// Data is every entry with data, oldest first.
	Data []models.ProtocolEntry
}

// ProtocolDetail is a protocol of the project with its digest.
type ProtocolDetail struct {
	Protocol *models.TestProtocol
	Digest
}

// DigestRepo is what the protocol digest needs.
type DigestRepo interface {
	// ProtocolDigest summarises the protocol; it lists at most limit errors.
	ProtocolDigest(ctx context.Context, protocolID uint, limit int) (Digest, error)
	// DataEntries lists the entries of the protocol that carry data, by id.
	DataEntries(ctx context.Context, protocolID uint) ([]models.ProtocolEntry, error)
}

// ProtocolDetail loads a protocol of the project with its digest.
func (s *Service) ProtocolDetail(ctx context.Context, projectID, protocolID uint) (*ProtocolDetail, error) {
	p, err := s.deps.Repo.ProjectProtocol(ctx, projectID, protocolID)
	if err != nil {
		return nil, err
	}
	digest, err := s.deps.Repo.ProtocolDigest(ctx, p.ID, DigestErrors)
	if err != nil {
		return nil, err
	}
	return &ProtocolDetail{Protocol: p, Digest: digest}, nil
}

// ProtocolData lists the entries of a protocol of the project that carry
// data, such as screenshot files, oldest first.
func (s *Service) ProtocolData(ctx context.Context, projectID, protocolID uint) ([]models.ProtocolEntry, error) {
	if _, err := s.deps.Repo.ProjectProtocol(ctx, projectID, protocolID); err != nil {
		return nil, err
	}
	return s.deps.Repo.DataEntries(ctx, protocolID)
}
