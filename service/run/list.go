package run

import (
	"context"
	"time"

	testsvc "github.com/fsuhrau/automationhub/service/test"
	"github.com/fsuhrau/automationhub/storage/models"
)

// How many runs a page of the runs of a test has by default and at most.
const (
	PageDefault = 25
	PageMax     = 100
)

// ListQuery asks for the runs of a test: at most Limit runs older than the
// run id Before, when it is not 0.
type ListQuery struct {
	testsvc.TestRef
	Before uint
	Limit  int
}

// BundleRef is the bundle a run ran with.
type BundleRef struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ListEntry is a run on the test page: its summary, what it ran with, and
// when its last test ended.
type ListEntry struct {
	testsvc.RunSummary
	Parameter string     `json:"parameter"`
	Bundle    *BundleRef `json:"bundle"`
	EndedAt   *time.Time `json:"endedAt"`
}

// RunsPage is a page of the runs of a test, newest first.
type RunsPage struct {
	Runs []ListEntry `json:"runs"`
	// more older runs exist; ask again with before = the last id
	More bool `json:"more"`
}

// List pages through the runs of a test of the app, newest first.
func (s *Service) List(ctx context.Context, q ListQuery) (*RunsPage, error) {
	if err := s.deps.Repo.TestInScope(ctx, q.TestRef); err != nil {
		return nil, err
	}
	found, err := s.deps.Repo.TestRuns(ctx, PageQuery{TestID: q.TestID, Before: q.Before, Limit: q.Limit + 1})
	if err != nil {
		return nil, err
	}
	page := &RunsPage{More: len(found) > q.Limit}
	if page.More {
		found = found[:q.Limit]
	}
	summaries, err := s.deps.Summarizer.Summaries(ctx, found)
	if err != nil {
		return nil, err
	}
	ends, err := s.endsOf(ctx, found)
	if err != nil {
		return nil, err
	}
	page.Runs = make([]ListEntry, len(found))
	for i := range found {
		page.Runs[i] = ListEntry{RunSummary: summaries[i], Parameter: found[i].Parameter, EndedAt: ends[found[i].ID]}
		if b := found[i].AppBinary; b != nil {
			page.Runs[i].Bundle = &BundleRef{ID: b.ID, Name: b.Name, Version: b.Version}
		}
	}
	return page, nil
}

// endsOf gives each run the latest end of its protocols; a run without an
// ended protocol has none.
func (s *Service) endsOf(ctx context.Context, found []models.TestRun) (map[uint]*time.Time, error) {
	ids := make([]uint, len(found))
	for i := range found {
		ids[i] = found[i].ID
	}
	ends := map[uint]*time.Time{}
	if len(ids) == 0 {
		return ends, nil
	}
	protocolEnds, err := s.deps.Repo.RunEnds(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range protocolEnds {
		e := protocolEnds[i]
		if latest := ends[e.TestRunID]; latest == nil || e.EndedAt.After(*latest) {
			ends[e.TestRunID] = &e.EndedAt
		}
	}
	return ends, nil
}
