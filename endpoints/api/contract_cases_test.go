package api_test

import (
	"github.com/fsuhrau/automationhub/authentication"
	"github.com/fsuhrau/automationhub/endpoints/api/apitest"
	"github.com/fsuhrau/automationhub/internal/testkit"
)

const (
	p   = "/api/proj"
	app = p + "/app/1"
	dev = p + "/device/1"
	tst = app + "/test/1"
	run = tst + "/run/1"
)

// contractGroups lists the snapshot cases per resource; each group has its
// golden file testdata/contract/<group>.golden.
func contractGroups() map[string][]contractCase {
	return map[string][]contractCase{
		"projects":       projectCases(),
		"settings":       settingsCases(),
		"profile":        profileCases(),
		"apps":           appCases(),
		"bundles":        bundleCases(),
		"tests":          testCases(),
		"test_runs":      testRunCases(),
		"protocols":      protocolCases(),
		"devices":        deviceCases(),
		"device_control": deviceControlCases(),
		"mcp":            mcpCases(),
		"sse":            sseCases(),
	}
}

func projectCases() []contractCase {
	return []contractCase{
		{name: "stats", route: "GET /api/stats", path: "/api/stats"},
		{name: "list", route: "GET /api/projects", path: "/api/projects"},
		{name: "create", route: "POST /api/project", path: "/api/project", body: `{"name":"New Project"}`},
		{name: "create without name", route: "POST /api/project", path: "/api/project", body: `{"name":" "}`},
		{name: "create malformed", route: "POST /api/project", path: "/api/project", body: `{`},
		{name: "update", route: "PUT /api/project/:project_id", path: "/api/project/proj", body: `{"name":"Renamed"}`},
		{name: "update unknown", route: "PUT /api/project/:project_id", path: "/api/project/nope", body: `{"name":"x"}`},
		{name: "delete", route: "DELETE /api/project/:project_id", path: "/api/project/1"},
		{name: "delete unknown", route: "DELETE /api/project/:project_id", path: "/api/project/99"},
		{name: "delete by identifier", route: "DELETE /api/project/:project_id", path: "/api/project/proj"},
		{name: "data traversal", route: "GET /api/data/:name", path: "/api/data/..%2Fsecret"},
		{name: "data missing", route: "GET /api/data/:name", path: "/api/data/missing.txt"},
		{name: "unknown project", route: "GET /api/:project_id/apps", path: "/api/nope/apps"},
	}
}

func settingsCases() []contractCase {
	s := p + "/settings"
	user := apitest.WithUser(authentication.User{Name: "Alice"})
	return []contractCase{
		{name: "users", route: "GET /api/:project_id/settings/users", path: s + "/users"},
		{name: "tokens", route: "GET /api/:project_id/settings/access_tokens", path: s + "/access_tokens"},
		{name: "token create", route: "POST /api/:project_id/settings/access_token", path: s + "/access_token", body: `{"name":"deploy"}`},
		{name: "token create without name", route: "POST /api/:project_id/settings/access_token", path: s + "/access_token", body: `{"name":""}`},
		{name: "token delete", route: "DELETE /api/:project_id/settings/access_token/:token_id", path: s + "/access_token/1"},
		{name: "token delete unknown", route: "DELETE /api/:project_id/settings/access_token/:token_id", path: s + "/access_token/99"},
		{name: "node create", route: "POST /api/:project_id/settings/node", path: s + "/node", body: `{"name":"Build node"}`},
		{name: "node create without name", route: "POST /api/:project_id/settings/node", path: s + "/node", body: `{}`},
		{name: "nodes", route: "GET /api/:project_id/settings/nodes", path: s + "/nodes"},
		{name: "node status", route: "GET /api/:project_id/settings/nodes/:node_id", path: s + "/nodes/1", opts: []apitest.Option{user}},
		{name: "node status unknown", route: "GET /api/:project_id/settings/nodes/:node_id", path: s + "/nodes/99"},
		{name: "node delete", route: "DELETE /api/:project_id/settings/nodes/:node_id", path: s + "/nodes/1"},
		{name: "node token", route: "POST /api/:project_id/settings/nodes/:node_id/token", path: s + "/nodes/1/token", body: `{"name":"agent"}`},
		{name: "node token without name", route: "POST /api/:project_id/settings/nodes/:node_id/token", path: s + "/nodes/1/token", body: `{}`},
		{name: "token auth refused", route: "GET /api/:project_id/settings/users", path: s + "/users", opts: []apitest.Option{apitest.WithAccessToken(1)}},
	}
}

func profileCases() []contractCase {
	user := apitest.WithUser(authentication.User{Name: "Alice"})
	return []contractCase{
		{name: "me", route: "GET /api/me", path: "/api/me", opts: []apitest.Option{user}},
		{name: "me signed out", route: "GET /api/me", path: "/api/me"},
		{name: "tokens", route: "GET /api/me/tokens", path: "/api/me/tokens", opts: []apitest.Option{user}},
		{name: "tokens signed out", route: "GET /api/me/tokens", path: "/api/me/tokens"},
		{name: "token create", route: "POST /api/me/tokens", path: "/api/me/tokens", body: `{"name":"laptop"}`, opts: []apitest.Option{user}},
		{name: "token create without name", route: "POST /api/me/tokens", path: "/api/me/tokens", body: `{}`, opts: []apitest.Option{user}},
		{name: "token delete unknown", route: "DELETE /api/me/tokens/:token_id", path: "/api/me/tokens/99", opts: []apitest.Option{user}},
		{name: "token delete refused", route: "DELETE /api/me/tokens/:token_id", path: "/api/me/tokens/1", opts: []apitest.Option{apitest.WithAccessToken(1)}},
	}
}

func appCases() []contractCase {
	return []contractCase{
		{name: "list", route: "GET /api/:project_id/apps", path: p + "/apps"},
		{name: "create", route: "POST /api/:project_id/app", path: p + "/app", body: `{"name":"Second","platform":2,"identifier":"com.example.two"}`},
		{name: "create without name", route: "POST /api/:project_id/app", path: p + "/app", body: `{"name":" "}`},
		{name: "update", route: "PUT /api/:project_id/app/:app_id/", path: app + "/", body: `{"name":"Renamed"}`},
		{name: "unknown app", route: "PUT /api/:project_id/app/:app_id/", path: p + "/app/99/", body: `{"name":"x"}`},
		{name: "parameter add", route: "POST /api/:project_id/app/:app_id/parameter", path: app + "/parameter", body: `{"name":"mode","type":"string"}`},
		{name: "parameter add malformed", route: "POST /api/:project_id/app/:app_id/parameter", path: app + "/parameter", body: `{`},
		{name: "parameter update", route: "PUT /api/:project_id/app/:app_id/parameter/:parameter_id", path: app + "/parameter/1", body: `{"name":"stage","type":"string"}`},
		{name: "parameter update unknown", route: "PUT /api/:project_id/app/:app_id/parameter/:parameter_id", path: app + "/parameter/99", body: `{"name":"x","type":"string"}`},
		{name: "parameter remove", route: "DELETE /api/:project_id/app/:app_id/parameter/:parameter_id", path: app + "/parameter/1"},
		{name: "parameter remove unknown", route: "DELETE /api/:project_id/app/:app_id/parameter/:parameter_id", path: app + "/parameter/99"},
	}
}

func bundleCases() []contractCase {
	b := app + "/bundle/1"
	return []contractCase{
		{name: "list", route: "GET /api/:project_id/app/:app_id/bundles", path: app + "/bundles"},
		{name: "get", route: "GET /api/:project_id/app/:app_id/bundle/:binary_id", path: b},
		{name: "get unknown", route: "GET /api/:project_id/app/:app_id/bundle/:binary_id", path: app + "/bundle/99"},
		{name: "icon", route: "GET /api/:project_id/app/:app_id/bundle/:binary_id/icon", path: b + "/icon"},
		{name: "read again", route: "POST /api/:project_id/app/:app_id/bundle/:binary_id/read", path: b + "/read"},
		{name: "archive", route: "POST /api/:project_id/app/:app_id/bundle/:binary_id/archive", path: b + "/archive"},
		{name: "update tags", route: "PUT /api/:project_id/app/:app_id/bundle/:binary_id", path: b, body: `{"tags":"beta"}`},
		{name: "update unknown", route: "PUT /api/:project_id/app/:app_id/bundle/:binary_id", path: app + "/bundle/99", body: `{"tags":"beta"}`},
		{name: "delete", route: "DELETE /api/:project_id/app/:app_id/bundle/:binary_id", path: b},
		{name: "delete unknown", route: "DELETE /api/:project_id/app/:app_id/bundle/:binary_id", path: app + "/bundle/99"},
		{name: "upload without file", route: "POST /api/:project_id/app/:app_id/upload", path: app + "/upload"},
	}
}

func testCases() []contractCase {
	return []contractCase{
		{name: "list", route: "GET /api/:project_id/app/:app_id/tests", path: app + "/tests"},
		{name: "get", route: "GET /api/:project_id/app/:app_id/test/:test_id", path: tst},
		{name: "get unknown", route: "GET /api/:project_id/app/:app_id/test/:test_id", path: app + "/test/99"},
		{name: "get invalid id", route: "GET /api/:project_id/app/:app_id/test/:test_id", path: app + "/test/abc"},
		{name: "create", route: "POST /api/:project_id/app/:app_id/test", path: app + "/test", body: `{"type":0,"name":"Regression","allDevices":true}`},
		{name: "create other type", route: "POST /api/:project_id/app/:app_id/test", path: app + "/test", body: `{"type":1,"name":"x"}`},
		{name: "update", route: "PUT /api/:project_id/app/:app_id/test/:test_id", path: tst, body: `{"name":"Smoke 2","allDevices":true}`},
		{name: "update without name", route: "PUT /api/:project_id/app/:app_id/test/:test_id", path: tst, body: `{"name":" "}`},
		{name: "run unknown test", route: "POST /api/:project_id/app/:app_id/test/:test_id/run", path: app + "/test/99/run", body: `{"appBinaryId":1}`},
		{name: "run invalid test id", route: "POST /api/:project_id/app/:app_id/test/:test_id/run", path: app + "/test/abc/run", body: `{}`},
		{name: "cancel finished run", route: "POST /api/:project_id/app/:app_id/test/:test_id/run/:run_id/cancel", path: run + "/cancel"},
		{name: "cancel invalid run id", route: "POST /api/:project_id/app/:app_id/test/:test_id/run/:run_id/cancel", path: tst + "/run/abc/cancel"},
	}
}

func testRunCases() []contractCase {
	return []contractCase{
		{name: "runs", route: "GET /api/:project_id/app/:app_id/test/:test_id/runs", path: tst + "/runs"},
		{name: "runs limited", route: "GET /api/:project_id/app/:app_id/test/:test_id/runs", path: tst + "/runs?limit=1&before=2"},
		{name: "runs unknown test", route: "GET /api/:project_id/app/:app_id/test/:test_id/runs", path: app + "/test/99/runs"},
		{name: "last", route: "GET /api/:project_id/app/:app_id/test/:test_id/runs/last", path: tst + "/runs/last"},
		{name: "last unknown test", route: "GET /api/:project_id/app/:app_id/test/:test_id/runs/last", path: app + "/test/99/runs/last"},
		{name: "run", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id", path: run},
		{name: "run unknown", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id", path: tst + "/run/99"},
		{name: "run invalid id", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id", path: tst + "/run/abc"},
	}
}

func protocolCases() []contractCase {
	return []contractCase{
		{name: "protocol", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id/:protocol_id", path: run + "/1"},
		{name: "protocol unknown", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id/:protocol_id", path: run + "/99"},
		{name: "baselines", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id/:protocol_id/baselines", path: run + "/1/baselines?anyDevice=true"},
		{name: "baselines unknown", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id/:protocol_id/baselines", path: run + "/99/baselines"},
		{name: "baseline", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id/:protocol_id/baseline/:baseline_run_id", path: run + "/1/baseline/2"},
		{name: "baseline unknown", route: "GET /api/:project_id/app/:app_id/test/:test_id/run/:run_id/:protocol_id/baseline/:baseline_run_id", path: run + "/1/baseline/99"},
	}
}

func deviceCases() []contractCase {
	return []contractCase{
		{name: "list", route: "GET /api/:project_id/devices", path: p + "/devices"},
		{name: "list by platform", route: "GET /api/:project_id/devices", path: p + "/devices?platform=android"},
		{name: "get", route: "GET /api/:project_id/device/:device_id", path: dev},
		{name: "get unknown", route: "GET /api/:project_id/device/:device_id", path: p + "/device/99"},
		{name: "update", route: "PUT /api/:project_id/device/:device_id", path: dev, body: `{"alias":"Lab","customParameter":[{"key":"a","value":"b"}]}`},
		{name: "update duplicate keys", route: "PUT /api/:project_id/device/:device_id", path: dev, body: `{"customParameter":[{"key":"a","value":"1"},{"key":"a","value":"2"}]}`},
		{name: "delete", route: "DELETE /api/:project_id/device/:device_id", path: dev},
		{name: "delete unknown", route: "DELETE /api/:project_id/device/:device_id", path: p + "/device/99"},
		{name: "unlock", route: "POST /api/:project_id/device/:device_id/unlock", path: dev + "/unlock"},
		{name: "run tests invalid signature", route: "POST /api/:project_id/device/:device_id/tests", path: dev + "/tests", body: `{"testName":"x"}`},
		{name: "protocols", route: "GET /api/:project_id/device/:device_id/protocols", path: dev + "/protocols?limit=5"},
		{name: "protocols unknown device", route: "GET /api/:project_id/device/:device_id/protocols", path: p + "/device/99/protocols"},
		{name: "overview", route: "GET /api/:project_id/overview", path: p + "/overview"},
	}
}

func connected(env *apitest.Env) { env.Devices.Add("android_device", testkit.NewFakeDevice("dev-1")) }

func deviceControlCases() []contractCase {
	c := func(name, method, suffix, body string) contractCase {
		streaming := suffix == "/stream" || suffix == "/video" || suffix == "/logs/stream"
		return contractCase{name: name, route: method + " /api/:project_id/device/:device_id" + suffix, path: dev + suffix, body: body, setup: connected, cancel: streaming, statusOnly: streaming}
	}
	return []contractCase{
		c("get lock", "GET", "/lock", ""),
		c("lock", "POST", "/lock", ""),
		c("release without lease", "DELETE", "/lock", ""),
		c("screenshot", "GET", "/screenshot", ""),
		c("stream", "GET", "/stream", ""),
		c("video", "GET", "/video", ""),
		c("hierarchy", "GET", "/hierarchy", ""),
		c("logs", "GET", "/logs", ""),
		c("logs stream", "GET", "/logs/stream", ""),
		c("tap", "POST", "/tap", `{"x":1,"y":2}`),
		c("swipe", "POST", "/swipe", `{}`),
		c("text", "POST", "/text", `{"text":"hi"}`),
		c("key", "POST", "/key", `{"key":"home"}`),
		c("element touch", "POST", "/element/touch", `{}`),
		c("element value", "POST", "/element/value", `{}`),
		c("app install", "POST", "/app/install", `{"bundleId":1}`),
		c("app start", "POST", "/app/start", `{}`),
		c("app stop", "POST", "/app/stop", `{}`),
		{name: "device not connected", route: "GET /api/:project_id/device/:device_id/lock", path: dev + "/lock"},
		{name: "unknown device", route: "POST /api/:project_id/device/:device_id/tap", path: p + "/device/99/tap", body: `{}`, setup: connected},
	}
}

func mcpCases() []contractCase {
	return []contractCase{
		{name: "post", route: "POST /api/:project_id/mcp", path: p + "/mcp", body: `{"jsonrpc":"2.0","id":1,"method":"ping"}`},
		{name: "get stream", route: "GET /api/:project_id/mcp", path: p + "/mcp", cancel: true, statusOnly: true},
		{name: "delete session", route: "DELETE /api/:project_id/mcp", path: p + "/mcp"},
		{name: "upload unknown", route: "PUT /api/:project_id/mcp/uploads/:upload_id", path: p + "/mcp/uploads/none", body: `data`},
	}
}

func sseCases() []contractCase {
	return []contractCase{
		{name: "stream", route: "GET /api/sse", path: "/api/sse", cancel: true, statusOnly: true},
		{name: "stream trailing slash", route: "GET /api/sse/", path: "/api/sse/", cancel: true, statusOnly: true},
	}
}
