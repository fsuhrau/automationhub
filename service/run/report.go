package run

import (
	"context"
	"slices"

	"github.com/fsuhrau/automationhub/storage/models"
)

// ReportLogLines is how many of the newest run log lines a report has.
const ReportLogLines = 20

// Report is a run of the project as a summary shows it: the run with its
// test, bundle and protocols, and the newest lines of its log.
type Report struct {
	Run *models.TestRun
	// Log is the newest lines of the run log, oldest first.
	Log []models.TestRunLogEntry
}

// Progress counts the protocols of a run and how many of them ended.
type Progress struct {
	Total int64
	Ended int64
}

// ReportRepo is what the run report and its progress need.
type ReportRepo interface {
	// ReportRun loads a run with its test, bundle and, by id, its protocols
	// with their device and keys.
	ReportRun(ctx context.Context, runID uint) (*models.TestRun, error)
	// LogTail lists the newest log entries of the run, newest first.
	LogTail(ctx context.Context, runID uint, limit int) ([]models.TestRunLogEntry, error)
	// ProtocolProgress counts the protocols of the run and those that ended.
	ProtocolProgress(ctx context.Context, runID uint) (Progress, error)
}

// Report loads a run of the project with the newest lines of its log. A run
// of another project is not found.
func (s *Service) Report(ctx context.Context, projectID, runID uint) (*Report, error) {
	if err := s.deps.Repo.RunInProject(ctx, projectID, runID); err != nil {
		return nil, err
	}
	loaded, err := s.deps.Repo.ReportRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	tail, err := s.deps.Repo.LogTail(ctx, runID, ReportLogLines)
	if err != nil {
		return nil, err
	}
	slices.Reverse(tail)
	return &Report{Run: loaded, Log: tail}, nil
}

// Progress tells how many protocols of a run of the project ended.
func (s *Service) Progress(ctx context.Context, projectID, runID uint) (Progress, error) {
	if err := s.deps.Repo.RunInProject(ctx, projectID, runID); err != nil {
		return Progress{}, err
	}
	return s.deps.Repo.ProtocolProgress(ctx, runID)
}
