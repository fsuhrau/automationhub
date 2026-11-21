package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	hubapp "github.com/fsuhrau/automationhub/app"
	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
)

var errBoom = errors.New("boom")

// fakeRepo is an in-memory Repo that honours the project scope. fail makes a
// method return the error.
type fakeRepo struct {
	apps    []models.App
	params  []models.AppParameter
	bundles []models.AppBinary
	runs    []models.TestRun
	fail    map[string]error
	nextID  uint
}

func (r *fakeRepo) failed(method string) error { return r.fail[method] }

func (r *fakeRepo) id() uint {
	r.nextID++
	return 100 + r.nextID
}

func (r *fakeRepo) Apps(_ context.Context, projectID uint) ([]models.App, error) {
	var out []models.App
	for _, a := range r.apps {
		if a.ProjectID == projectID {
			out = append(out, a)
		}
	}
	return out, r.failed("Apps")
}

func (r *fakeRepo) App(_ context.Context, sc Scope) (*models.App, error) {
	if err := r.failed("App"); err != nil {
		return nil, err
	}
	for i := range r.apps {
		if r.apps[i].ID == sc.AppID && r.apps[i].ProjectID == sc.ProjectID {
			a := r.apps[i]
			return &a, nil
		}
	}
	return nil, apperr.NotFound("app %d not found", sc.AppID)
}

func (r *fakeRepo) CreateApp(_ context.Context, app *models.App) error {
	app.ID = r.id()
	r.apps = append(r.apps, *app)
	return r.failed("CreateApp")
}

func (r *fakeRepo) SaveApp(_ context.Context, app *models.App) error {
	for i := range r.apps {
		if r.apps[i].ID == app.ID {
			r.apps[i] = *app
		}
	}
	return r.failed("SaveApp")
}

func (r *fakeRepo) Parameters(_ context.Context, sc Scope) ([]models.AppParameter, error) {
	var out []models.AppParameter
	for _, p := range r.params {
		if p.AppID == sc.AppID {
			out = append(out, p)
		}
	}
	return out, r.failed("Parameters")
}

func (r *fakeRepo) Parameter(_ context.Context, ref ParameterRef) (*models.AppParameter, error) {
	if err := r.failed("Parameter"); err != nil {
		return nil, err
	}
	for i := range r.params {
		if r.params[i].ID == ref.ID && r.params[i].AppID == ref.AppID && r.owns(ref.Scope) {
			p := r.params[i]
			return &p, nil
		}
	}
	return nil, apperr.NotFound("parameter %d not found", ref.ID)
}

func (r *fakeRepo) owns(sc Scope) bool {
	_, err := r.App(context.Background(), sc)
	return err == nil
}

func (r *fakeRepo) CreateParameter(_ context.Context, p *models.AppParameter) error {
	p.ID = r.id()
	r.params = append(r.params, *p)
	return r.failed("CreateParameter")
}

func (r *fakeRepo) SaveParameter(_ context.Context, p *models.AppParameter) error {
	for i := range r.params {
		if r.params[i].ID == p.ID {
			r.params[i] = *p
		}
	}
	return r.failed("SaveParameter")
}

func (r *fakeRepo) DeleteParameter(_ context.Context, ref ParameterRef) error {
	if !r.owns(ref.Scope) {
		return r.failed("DeleteParameter")
	}
	for i := range r.params {
		if r.params[i].ID == ref.ID && r.params[i].AppID == ref.AppID {
			r.params = append(r.params[:i], r.params[i+1:]...)
			break
		}
	}
	return r.failed("DeleteParameter")
}

func (r *fakeRepo) Bundles(_ context.Context, sc Scope) ([]models.AppBinary, error) {
	var out []models.AppBinary
	for i := len(r.bundles) - 1; i >= 0 && r.owns(sc); i-- {
		if r.bundles[i].AppID == sc.AppID {
			out = append(out, r.bundles[i])
		}
	}
	return out, r.failed("Bundles")
}

func (r *fakeRepo) Bundle(_ context.Context, ref BundleRef) (*models.AppBinary, error) {
	if err := r.failed("Bundle"); err != nil {
		return nil, err
	}
	for i := range r.bundles {
		if r.bundles[i].ID == ref.ID && r.bundles[i].AppID == ref.AppID && r.owns(ref.Scope) {
			b := r.bundles[i]
			return &b, nil
		}
	}
	return nil, apperr.NotFound("bundle %d not found", ref.ID)
}

func (r *fakeRepo) BundleByHash(_ context.Context, sc Scope, hash string) (*models.AppBinary, error) {
	if err := r.failed("BundleByHash"); err != nil {
		return nil, err
	}
	for i := range r.bundles {
		if r.bundles[i].Hash == hash && r.bundles[i].AppID == sc.AppID {
			b := r.bundles[i]
			return &b, nil
		}
	}
	return nil, apperr.NotFound("bundle %s not found", hash)
}

func (r *fakeRepo) CreateBundle(_ context.Context, b *models.AppBinary) error {
	if err := r.failed("CreateBundle"); err != nil {
		return err
	}
	b.ID = r.id()
	r.bundles = append(r.bundles, *b)
	return nil
}

func (r *fakeRepo) replace(b *models.AppBinary) {
	for i := range r.bundles {
		if r.bundles[i].ID == b.ID {
			r.bundles[i] = *b
		}
	}
}

func (r *fakeRepo) DeleteBundle(_ context.Context, b *models.AppBinary) error {
	for i := range r.bundles {
		if r.bundles[i].ID == b.ID {
			r.bundles = append(r.bundles[:i], r.bundles[i+1:]...)
			break
		}
	}
	return r.failed("DeleteBundle")
}

func (r *fakeRepo) SaveTags(_ context.Context, b *models.AppBinary) error {
	r.replace(b)
	return r.failed("SaveTags")
}

func (r *fakeRepo) SaveArchivedAt(_ context.Context, b *models.AppBinary) error {
	r.replace(b)
	return r.failed("SaveArchivedAt")
}

func (r *fakeRepo) SaveRead(_ context.Context, b *models.AppBinary, _ []string) error {
	r.replace(b)
	return r.failed("SaveRead")
}

func (r *fakeRepo) RunCounts(_ context.Context, ids []uint) (map[uint]int64, error) {
	counts := map[uint]int64{}
	for _, run := range r.runs {
		for _, id := range ids {
			if run.AppBinaryID == id {
				counts[id]++
			}
		}
	}
	return counts, r.failed("RunCounts")
}

func (r *fakeRepo) RunCount(_ context.Context, id uint) (int64, error) {
	var n int64
	for _, run := range r.runs {
		if run.AppBinaryID == id {
			n++
		}
	}
	return n, r.failed("RunCount")
}

func (r *fakeRepo) RecentRuns(_ context.Context, id uint, limit int) ([]models.TestRun, error) {
	var out []models.TestRun
	for i := len(r.runs) - 1; i >= 0 && len(out) < limit; i-- {
		if r.runs[i].AppBinaryID == id {
			out = append(out, r.runs[i])
		}
	}
	return out, r.failed("RecentRuns")
}

func (r *fakeRepo) RecentRunIDs(ctx context.Context, id uint, limit int) ([]uint, error) {
	runs, _ := r.RecentRuns(ctx, id, limit)
	ids := make([]uint, len(runs))
	for i, run := range runs {
		ids[i] = run.ID
	}
	return ids, r.failed("RecentRunIDs")
}

// fakeFiles keeps the files in a map below /store.
type fakeFiles struct {
	data map[string][]byte
	fail map[string]error
}

func newFakeFiles() *fakeFiles {
	return &fakeFiles{data: map[string][]byte{}, fail: map[string]error{}}
}

func (f *fakeFiles) Path(name string) string {
	if name == "" || strings.HasPrefix(name, "..") {
		return ""
	}
	return "/store/" + name
}

func (f *fakeFiles) Save(name string, src io.Reader) (string, error) {
	if err := f.fail["Save"]; err != nil {
		return "", err
	}
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(src)
	path := f.Path(name)
	f.data[path] = buf.Bytes()
	return path, nil
}

func (f *fakeFiles) Move(from, to string) error {
	if err := f.fail["Move"]; err != nil {
		return err
	}
	f.data[to] = f.data[from]
	delete(f.data, from)
	return nil
}

func (f *fakeFiles) Remove(path string) error {
	if err := f.fail["Remove:"+path]; err != nil {
		return err
	}
	delete(f.data, path)
	return nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	if err := f.fail["Read"]; err != nil {
		return nil, err
	}
	data, ok := f.data[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.data[path] = data
	return nil
}

func (f *fakeFiles) has(path string) bool { _, ok := f.data[path]; return ok }

// fakeAnalyser answers every file with the parameter, or fails.
type fakeAnalyser struct {
	params     hubapp.Parameter
	analyseErr error
	readErr    error
	icon       []byte
	iconType   string
	iconErr    error
	analysed   []string
}

func (a *fakeAnalyser) Analyse(path string) (*hubapp.Parameter, error) {
	a.analysed = append(a.analysed, path)
	p := a.params
	return &p, a.analyseErr
}

func (a *fakeAnalyser) Read(string) (*hubapp.Parameter, error) {
	p := a.params
	return &p, a.readErr
}

func (a *fakeAnalyser) Icon(string, string) ([]byte, string, error) {
	return a.icon, a.iconType, a.iconErr
}

// fixture is a service on the fakes with project 1, app 1 and, for project 2,
// app 2.
type fixture struct {
	svc      *Service
	repo     *fakeRepo
	files    *fakeFiles
	analyser *fakeAnalyser
	running  map[uint]bool
	sc       Scope
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		repo: &fakeRepo{
			apps: []models.App{
				{Model: models.Model{ID: 1}, ProjectID: 1, Name: "App", Identifier: "com.app"},
				{Model: models.Model{ID: 2}, ProjectID: 2, Name: "Other", Identifier: "com.other"},
			},
			fail: map[string]error{},
		},
		files:    newFakeFiles(),
		analyser: &fakeAnalyser{},
		running:  map[uint]bool{},
		sc:       Scope{ProjectID: 1, AppID: 1},
	}
	f.svc = New(Deps{
		Repo:     f.repo,
		Files:    f.files,
		Analyser: f.analyser,
		Running:  func(id uint) bool { return f.running[id] },
		Summarize: func(runs []models.TestRun) ([]RunSummary, error) {
			out := make([]RunSummary, len(runs))
			for i, r := range runs {
				out[i] = RunSummary{ID: r.ID, Status: "passed"}
			}
			return out, f.repo.fail["Summarize"]
		},
	})
	return f
}

// addBundle stores a bundle of the fixture's app and returns its reference.
func (f *fixture) addBundle(b models.AppBinary) BundleRef {
	b.AppID = f.sc.AppID
	b.ID = f.repo.id()
	f.repo.bundles = append(f.repo.bundles, b)
	return BundleRef{Scope: f.sc, ID: b.ID}
}
