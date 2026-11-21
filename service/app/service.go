// Package app is the use cases of apps and their bundles: creating and
// editing apps and their parameters, listing, reading, archiving and
// deleting bundles, and storing an uploaded build. Every method enforces the
// project scope: an app is only reachable through its project, and a bundle
// only through its app.
package app

import (
	"context"
	"io"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

// ErrBundleExists is returned, together with the stored bundle, when the same
// build was uploaded before and is not archived.
var ErrBundleExists = apperr.Conflict("app bundle exists already")

// Scope is where an app lives: the project that owns it.
type Scope struct {
	ProjectID uint
	AppID     uint
}

// BundleRef names one bundle of an app.
type BundleRef struct {
	Scope
	ID uint
}

// ParameterRef names one parameter of an app.
type ParameterRef struct {
	Scope
	ID uint
}

// AppRepo is the storage of apps and their parameters. Lookups that find
// nothing return an apperr.NotFound error.
type AppRepo interface {
	Apps(ctx context.Context, projectID uint) ([]models.App, error)
	App(ctx context.Context, sc Scope) (*models.App, error)
	CreateApp(ctx context.Context, app *models.App) error
	SaveApp(ctx context.Context, app *models.App) error
	Parameters(ctx context.Context, sc Scope) ([]models.AppParameter, error)
	Parameter(ctx context.Context, ref ParameterRef) (*models.AppParameter, error)
	CreateParameter(ctx context.Context, p *models.AppParameter) error
	SaveParameter(ctx context.Context, p *models.AppParameter) error
	// DeleteParameter removes the parameter if the scope has it.
	DeleteParameter(ctx context.Context, ref ParameterRef) error
}

// BundleRepo is the storage of bundles. Lookups that find nothing return an
// apperr.NotFound error.
type BundleRepo interface {
	Bundles(ctx context.Context, sc Scope) ([]models.AppBinary, error)
	Bundle(ctx context.Context, ref BundleRef) (*models.AppBinary, error)
	BundleByHash(ctx context.Context, sc Scope, hash string) (*models.AppBinary, error)
	CreateBundle(ctx context.Context, b *models.AppBinary) error
	DeleteBundle(ctx context.Context, b *models.AppBinary) error
	SaveTags(ctx context.Context, b *models.AppBinary) error
	SaveArchivedAt(ctx context.Context, b *models.AppBinary) error
	// SaveRead stores the columns of the bundle that reading its file sets.
	SaveRead(ctx context.Context, b *models.AppBinary, columns []string) error
}

// RunRepo is what the service needs to know about the runs of bundles.
type RunRepo interface {
	RunCounts(ctx context.Context, bundleIDs []uint) (map[uint]int64, error)
	RunCount(ctx context.Context, bundleID uint) (int64, error)
	// RecentRuns lists the newest runs that used the bundle, with their test.
	RecentRuns(ctx context.Context, bundleID uint, limit int) ([]models.TestRun, error)
	RecentRunIDs(ctx context.Context, bundleID uint, limit int) ([]uint, error)
}

// Repo is the whole storage of the service.
type Repo interface {
	AppRepo
	BundleRepo
	RunRepo
}

// Files is the folder the bundle files live in.
type Files interface {
	// Path is the absolute path of a stored file, or "" when name is empty
	// or would leave the folder.
	Path(name string) string
	// Save stores the upload under the base name and returns its path.
	Save(name string, src io.Reader) (string, error)
	Move(from, to string) error
	// Remove deletes a file or folder; a missing one is not an error.
	Remove(path string) error
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
}

// Analyser reads what a bundle file says about the app.
type Analyser interface {
	// Analyse reads and hashes the file.
	Analyse(path string) (*hubapp.Parameter, error)
	// Read reads the file without hashing it.
	Read(path string) (*hubapp.Parameter, error)
	// Icon extracts the icon of the bundle at the path inside it.
	Icon(path, iconPath string) (data []byte, contentType string, err error)
}

// RunningFunc tells whether this master is running the run right now.
type RunningFunc func(runID uint) bool

// SummarizeFunc gives each run its status and result counts, in order.
type SummarizeFunc func(testRuns []models.TestRun) ([]RunSummary, error)

// Deps are what New needs.
type Deps struct {
	Repo      Repo
	Files     Files
	Analyser  Analyser
	Running   RunningFunc
	Summarize SummarizeFunc
}

// Service holds the app and bundle use cases.
type Service struct {
	repo      Repo
	files     Files
	analyser  Analyser
	running   RunningFunc
	summarize SummarizeFunc
}

// New builds the service.
func New(deps Deps) *Service {
	return &Service{
		repo: deps.Repo, files: deps.Files, analyser: deps.Analyser,
		running: deps.Running, summarize: deps.Summarize,
	}
}
