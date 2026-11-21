package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/fsuhrau/automationhub/internal/apitest"
)

// The contract snapshots pin status and body of every route of the api, so
// refactorings of the handlers can't change what clients see. Refresh the
// golden files with: go test ./endpoints/api/ -run Contract -update-contract
var updateContract = flag.Bool("update-contract", false, "rewrite the contract golden files in testdata/contract") //nolint:gochecknoglobals // test flag

// contractCase is one request against a freshly seeded api.
type contractCase struct {
	name   string // unique within its group
	route  string // "METHOD /registered/:pattern" it exercises
	path   string // the concrete request path
	body   string // raw JSON request body
	setup  func(*apitest.Env)
	opts   []apitest.Option
	cancel bool // the client is gone already, for streaming routes
	// statusOnly pins just the status; for streams whose content depends on timing
	statusOnly bool
}

func (c contractCase) method() string { return strings.Fields(c.route)[0] }

func (c contractCase) request() *http.Request {
	req := httptest.NewRequest(c.method(), c.path, strings.NewReader(c.body))
	req.Header.Set("Content-Type", "application/json")
	if c.cancel {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		req = req.WithContext(ctx)
	}
	return req
}

// volatile matches the values that differ from run to run: timestamps,
// generated tokens and hints, and uuids.
var volatile = []struct { //nolint:gochecknoglobals // read-only lookup table of the test
	pattern *regexp.Regexp
	with    string
}{
	{regexp.MustCompile(`"\d{4}-\d\d-\d\dT[^"]*"`), `"<time>"`},
	{regexp.MustCompile(`"(token|hint)": "[^"]*"`), `"$1": "<generated>"`},
	{regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`), `<uuid>`},
}

// snapshot renders a response as stable text: ids are deterministic with
// the fixed seed, timestamps are replaced.
func snapshot(c contractCase, rec *httptest.ResponseRecorder) string {
	body := rec.Body.Bytes()
	var pretty bytes.Buffer
	if json.Indent(&pretty, body, "", "  ") == nil {
		body = pretty.Bytes()
	}
	text := string(body)
	for _, v := range volatile {
		text = v.pattern.ReplaceAllString(text, v.with)
	}
	if c.statusOnly {
		return fmt.Sprintf("== %s\n%s %s\n%d\n\n", c.name, c.method(), c.path, rec.Code)
	}
	return fmt.Sprintf("== %s\n%s %s\n%d %s\n%s\n\n", c.name, c.method(), c.path,
		rec.Code, rec.Header().Get("Content-Type"), strings.TrimSpace(text))
}

func runContractCase(t *testing.T, c contractCase) string {
	t.Helper()
	env := apitest.New(t, c.opts...)
	seedContract(env)
	if c.setup != nil {
		c.setup(env)
	}
	return snapshot(c, env.Request(c.request()))
}

func checkGolden(t *testing.T, group, got string) {
	t.Helper()
	file := filepath.Join("testdata", "contract", group+".golden")
	if *updateContract {
		if err := os.WriteFile(file, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("%v; create it with -update-contract", err)
	}
	if string(want) != got {
		t.Errorf("%s differs from the snapshot; rerun with -update-contract if the change is intended\n--- want\n%s\n--- got\n%s", file, want, got)
	}
}

func TestContractSnapshots(t *testing.T) {
	for group, cases := range contractGroups() {
		t.Run(group, func(t *testing.T) {
			var out strings.Builder
			for _, c := range cases {
				t.Run(c.name, func(t *testing.T) { out.WriteString(runContractCase(t, c)) })
			}
			checkGolden(t, group, out.String())
		})
	}
}

// TestContractCoversEveryRoute fails when a route of the router has no
// snapshot case, so a route can't be added or dropped unnoticed.
func TestContractCoversEveryRoute(t *testing.T) {
	covered := map[string]bool{}
	names := map[string]bool{}
	for group, cases := range contractGroups() {
		for _, c := range cases {
			covered[c.route] = true
			if key := group + "/" + c.name; names[key] {
				t.Errorf("duplicate case %s", key)
			}
			names[group+"/"+c.name] = true
		}
	}
	registered := map[string]bool{}
	for _, r := range apitest.New(t).Router.Routes() {
		registered[r.Method+" "+r.Path] = true
	}
	for _, route := range sortedKeys(registered) {
		if !covered[route] {
			t.Errorf("route %s has no contract snapshot case", route)
		}
	}
	for _, route := range sortedKeys(covered) {
		if !registered[route] {
			t.Errorf("contract case for %s, which is not registered", route)
		}
	}
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
