package app

import (
	"time"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/hub/runs"
	"github.com/fsuhrau/automationhub/storage/models"
)

const bundleRunsLimit = 25

// RunSummary is one run on the bundle page: when it ran and how it went.
type RunSummary struct {
	ID        uint          `json:"id"`
	CreatedAt time.Time     `json:"createdAt"`
	State     runs.RunState `json:"state"`
	Status    string        `json:"status"`
	Passed    int           `json:"passed"`
	Failed    int           `json:"failed"`
	Unstable  int           `json:"unstable"`
	Open      int           `json:"open"`
}

// BundleEntry is a bundle with how many test runs used it, so the bundles
// page can warn before deleting one that runs point to.
type BundleEntry struct {
	models.AppBinary
	RunCount int64 `json:"runCount"`
}

// BundleRun is a test run that used a bundle.
type BundleRun struct {
	RunSummary
	TestID   uint   `json:"testId"`
	TestName string `json:"testName"`
}

// BundleDetail is a bundle with all read from its file and the runs that
// used it, for the bundle page.
type BundleDetail struct {
	models.AppBinary
	RunCount int64       `json:"runCount"`
	Runs     []BundleRun `json:"runs"`
}

// Icon is the icon of a bundle.
type Icon struct {
	Data        []byte
	ContentType string
}

// summaryInfo is the part of a bundle's info the bundle list shows.
func summaryInfo(info *hubapp.BundleInfo) *hubapp.BundleInfo {
	if info == nil {
		return nil
	}
	return &hubapp.BundleInfo{
		Format:      info.Format,
		DisplayName: info.DisplayName,
		Identifier:  info.Identifier,
		Version:     info.Version,
		Build:       info.Build,
		Icon:        info.Icon,
		Engine:      info.Engine,
		ReadAt:      info.ReadAt,
	}
}
