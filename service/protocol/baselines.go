package protocol

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/fsuhrau/automationhub/storage/models"
)

// A protocol's performance is compared with a baseline: another run of the
// same test that measured performance. By default that is the latest earlier
// run on the same device; the filters let the user pick one by bundle name,
// version, tag or run, on the same device or on any. A baseline is chosen by
// run, so it stays while moving through the tests of a run.

// How many candidates are listed by default and at most.
const (
	DefaultCandidates = 50
	MaxCandidates     = 200
)

// Filter narrows the candidates: Query matches the bundle's name, version
// or tags, the run's parameters or a run id; Tag matches one tag exactly;
// AnyDevice drops the same device rule.
type Filter struct {
	Tag       string
	Query     string
	AnyDevice bool
	Limit     int
}

// CandidateQuery asks the repo for the candidates of the current protocol.
type CandidateQuery struct {
	AppID   uint
	Current *models.TestProtocol
	Filter  Filter
}

// BaselineQuery asks the repo for the baseline in a run: on the device
// DeviceID, or else preferably the device of the current protocol.
type BaselineQuery struct {
	AppID    uint
	Current  *models.TestProtocol
	RunID    uint
	DeviceID *uint
}

// Bundle is the bundle a candidate ran with.
type Bundle struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Tags     string `json:"tags"`
}

// Candidate is a protocol the current one can be compared with.
type Candidate struct {
	ProtocolID uint                   `json:"protocolId"`
	TestRunID  uint                   `json:"testRunId"`
	TestID     uint                   `json:"testId"`
	StartedAt  time.Time              `json:"startedAt"`
	TestResult models.TestResultState `json:"testResult"`
	DeviceID   *uint                  `json:"deviceId,omitempty"`
	Device     string                 `json:"device"`
	Bundle     *Bundle                `json:"bundle"`
	Parameter  string                 `json:"parameter"`
	Averages
}

// Baselines is the list of candidates and the tags to filter them by.
type Baselines struct {
	Candidates []Candidate `json:"candidates"`
	// the tags of the app's bundles, to filter by
	Tags []string `json:"tags"`
}

// Baseline is one candidate with its checkpoints.
type Baseline struct {
	Candidate
	Performance []models.ProtocolPerformanceEntry `json:"performance"`
}

// Baselines lists what the protocol can be compared with, newest first.
func (s *Service) Baselines(ctx context.Context, key Key, f Filter) (*Baselines, error) {
	current, err := s.repo.Protocol(ctx, key)
	if err != nil {
		return nil, err
	}
	f.Tag, f.Query = strings.TrimSpace(f.Tag), strings.ToLower(strings.TrimSpace(f.Query))
	found, err := s.repo.Candidates(ctx, CandidateQuery{AppID: key.AppID, Current: current, Filter: f})
	if err != nil {
		return nil, err
	}
	lists, err := s.repo.BundleTags(ctx, key.AppID)
	if err != nil {
		return nil, err
	}
	out := &Baselines{Candidates: make([]Candidate, len(found)), Tags: splitTags(lists)}
	for i := range found {
		out.Candidates[i] = s.candidate(&found[i])
	}
	return out, nil
}

// Choice picks the baseline: the run it is in, and optionally the device it
// ran on.
type Choice struct {
	RunID    uint
	DeviceID *uint
}

// Baseline loads the baseline of the choice with its checkpoints: the
// protocol of the same test in the chosen run. It has to be one the
// protocol can be compared with.
func (s *Service) Baseline(ctx context.Context, key Key, choice Choice) (*Baseline, error) {
	current, err := s.repo.Protocol(ctx, key)
	if err != nil {
		return nil, err
	}
	found, err := s.repo.Baseline(ctx, BaselineQuery{AppID: key.AppID, Current: current, RunID: choice.RunID, DeviceID: choice.DeviceID})
	if err != nil {
		return nil, err
	}
	return &Baseline{Candidate: s.candidate(found), Performance: found.Performance}, nil
}

func (s *Service) candidate(p *models.TestProtocol) Candidate {
	c := Candidate{
		ProtocolID: p.ID, TestRunID: p.TestRunID, StartedAt: p.StartedAt, TestResult: p.TestResult,
		DeviceID: p.DeviceID, Averages: AveragesOf(p),
	}
	if p.Device != nil {
		c.Device = s.deviceName(p.Device)
	}
	if p.TestRun != nil {
		c.TestID, c.Parameter = p.TestRun.TestID, p.TestRun.Parameter
		c.Bundle = bundleOf(p.TestRun.AppBinary)
	}
	return c
}

func bundleOf(b *models.AppBinary) *Bundle {
	if b == nil {
		return nil
	}
	return &Bundle{ID: b.ID, Name: b.Name, Version: b.Version, Platform: b.Platform, Tags: b.Tags}
}

// splitTags merges comma separated tag lists into the sorted distinct tags.
func splitTags(lists []string) []string {
	seen := map[string]bool{}
	tags := []string{}
	for _, list := range lists {
		for _, t := range strings.Split(list, ",") {
			if t = strings.TrimSpace(t); t != "" && !seen[t] {
				seen[t] = true
				tags = append(tags, t)
			}
		}
	}
	sort.Strings(tags)
	return tags
}
