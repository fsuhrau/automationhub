// Package appfixture is the test setup of the app and bundle routes: the api
// with two projects that have an app each, a temporary bundle folder, and
// helpers to store bundles and send raw and multipart requests.
package appfixture

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/endpoints/api/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
)

// Env is the api with the project "mine" and its app, the project "other"
// with an app of its own, and an empty bundle folder.
type Env struct {
	*apitest.Env
	Mine, Other *models.App
	Store       string

	t testing.TB
}

// New builds the environment; the bundle folder is restored when the test
// ends, so tests that use it must not run in parallel.
func New(t testing.TB) *Env {
	t.Helper()
	store := t.TempDir()
	old := apps.AppStoragePath
	apps.AppStoragePath = store
	t.Cleanup(func() { apps.AppStoragePath = old })

	env := apitest.New(t)
	project := func(identifier string) uint {
		return testkit.Project(env.DB, func(p *models.Project) { p.Identifier = identifier }).ID
	}
	mine := testkit.App(env.DB, func(a *models.App) { a.ProjectID = project("mine") })
	other := testkit.App(env.DB, func(a *models.App) { a.ProjectID = project("other") })
	return &Env{Env: env, Mine: mine, Other: other, Store: store, t: t}
}

// AppPath is the route below the app of the project "mine".
func (e *Env) AppPath(suffix string) string {
	return fmt.Sprintf("/api/mine/app/%d%s", e.Mine.ID, suffix)
}

// BundlePath is the route below a bundle of the app of the project "mine".
func (e *Env) BundlePath(id uint, suffix string) string {
	return e.AppPath(fmt.Sprintf("/bundle/%d%s", id, suffix))
}

// AddBundle stores the bundle for the app of the project "mine".
func (e *Env) AddBundle(b models.AppBinary) *models.AppBinary {
	b.AppID = e.Mine.ID
	e.DB.Create(&b)
	return &b
}

// Raw sends the text as the JSON body as it is, valid or not.
func (e *Env) Raw(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return e.Request(req)
}

// IPA is a minimal ipa with an icon.
func IPA(t testing.TB) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	plist, _ := w.Create("Payload/Game.app/Info.plist")
	_, _ = plist.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>com.example.game</string>
<key>CFBundleShortVersionString</key><string>1.2.3</string>
<key>CFBundleVersion</key><string>45</string>
<key>CFBundleDisplayName</key><string>Game</string>
<key>MinimumOSVersion</key><string>15.0</string>
<key>CFBundleIconFiles</key><array><string>Icon</string></array>
</dict></plist>`))
	icon, _ := w.Create("Payload/Game.app/Icon@2x.png")
	_, _ = icon.Write([]byte("\x89PNG\r\n\x1a\nnot crushed"))
	must(t, w.Close())
	return buf.Bytes()
}

// WriteFile stores the content in the bundle folder and returns its path.
func (e *Env) WriteFile(name string, content []byte) string {
	e.t.Helper()
	path := filepath.Join(e.Store, name)
	must(e.t, os.WriteFile(path, content, 0o600))
	return path
}

// WriteIPA stores a minimal ipa in the bundle folder and returns its path.
func (e *Env) WriteIPA(name string) string {
	e.t.Helper()
	return e.WriteFile(name, IPA(e.t))
}

// Upload sends the content as the form file of the upload route.
func (e *Env) Upload(field, name string, content []byte) *httptest.ResponseRecorder {
	e.t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile(field, name)
	must(e.t, err)
	_, _ = io.Copy(part, bytes.NewReader(content))
	_ = form.Close()
	req := httptest.NewRequest("POST", e.AppPath("/upload"), &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	return e.Request(req)
}

// must stops the test when err is set.
func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
