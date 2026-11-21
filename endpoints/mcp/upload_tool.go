package mcp

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const downloadTimeout = 30 * time.Minute

type UploadBundleArgs struct {
	AppID uint   `json:"appId" jsonschema:"app id from list_apps"`
	URL   string `json:"url" jsonschema:"http(s) URL of the .apk, .ipa or app archive"`
}

type BeginUploadArgs struct {
	AppID    uint   `json:"appId" jsonschema:"app id from list_apps"`
	FileName string `json:"fileName" jsonschema:"file name of the build, e.g. Client-1.58.0.apk"`
}

func (t *tools) registerUploads(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "upload_bundle", Description: "Download a build from an http(s) URL and add it as a bundle of the app. For a local file use begin_upload.", Annotations: writeSafe()}, t.uploadBundle)
	mcp.AddTool(s, &mcp.Tool{Name: "begin_upload", Description: "Start the upload of a local build file (any size). Returns a single-use uploadUrl, valid for 30 minutes: PUT the raw file to it, e.g. curl --fail-with-body -T <file> <uploadUrl>. The PUT response is the new bundle.", Annotations: writeSafe()}, t.beginUpload)
}

// buildSource checks the URL of upload_bundle and names the file it serves.
func buildSource(raw string) (*url.URL, string, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, "", fmt.Errorf("url must be http or https")
	}
	name, err := uploadFileName(path.Base(u.Path))
	if err != nil {
		return nil, "", fmt.Errorf("url has no file name")
	}
	return u, name, nil
}

// download starts the GET of the build; the caller closes the body.
func download(ctx context.Context, u *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download failed: %s", resp.Status)
	}
	return resp, nil
}

// bundleResult is what a stored upload answers: the bundle, or the bundle
// that exists already.
func bundleResult(bundle bundleInfo, exists bool) any {
	if exists {
		return map[string]any{"exists": true, "bundle": bundle}
	}
	return bundle
}

func (t *tools) uploadBundle(ctx context.Context, req *mcp.CallToolRequest, in UploadBundleArgs) (*mcp.CallToolResult, any, error) {
	application, err := t.app(ctx, in.AppID)
	if err != nil {
		return nil, nil, err
	}
	u, name, err := buildSource(in.URL)
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()
	resp, err := download(ctx, u)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	scope := appsvc.Scope{ProjectID: t.project.ID, AppID: application.ID}
	bundle, exists, err := t.storeUpload(ctx, uploadFile{scope: scope, name: name, body: resp.Body})
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(bundleResult(toBundleInfo(bundle), exists))
}

// requestHeader is the header of the request of a tool call; empty when the
// call has none.
func requestHeader(req *mcp.CallToolRequest) http.Header {
	if req != nil && req.Extra != nil && req.Extra.Header != nil {
		return req.Extra.Header
	}
	return http.Header{}
}

func (t *tools) beginUpload(ctx context.Context, req *mcp.CallToolRequest, in BeginUploadArgs) (*mcp.CallToolResult, any, error) {
	application, err := t.app(ctx, in.AppID)
	if err != nil {
		return nil, nil, err
	}
	name, err := uploadFileName(in.FileName)
	if err != nil {
		return nil, nil, err
	}
	id, expires := t.uploads.create(pendingUpload{projectID: t.project.ID, appID: application.ID, fileName: name, caller: t.caller})
	u := newUploadAddress(requestHeader(req)).url(t.project, id)
	return jsonResult(map[string]any{
		"uploadUrl": u,
		"method":    "PUT",
		"expiresAt": expires,
		"maxBytes":  t.MaxUpload,
		"example":   fmt.Sprintf("curl --fail-with-body -T '<path to %s>' '%s'", name, u),
		"note": "Send the raw file as the PUT body, once. Through the AutomationHub host proxy no auth header is needed; " +
			"when connected to the master directly, send the same X-Auth-Token as for MCP. The response is the stored bundle.",
	})
}
