package run

import (
	"context"
	"strconv"

	"github.com/fsuhrau/automationhub/apperr"
)

// Cancel cancels a run of the project that this master started and that is
// still active.
func (s *Service) Cancel(ctx context.Context, projectID, runID uint) error {
	if err := s.deps.Repo.RunInProject(ctx, projectID, runID); err != nil {
		return err
	}
	runner, ok := s.deps.Runners.Get(runID)
	if !ok {
		return apperr.NotFound("not found: no active runner for test run %d", runID)
	}
	if err := runner.Cancel(strconv.FormatUint(uint64(runID), 10)); err != nil {
		return err
	}
	s.deps.Runners.Remove(runID)
	return nil
}
