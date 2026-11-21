package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/fsuhrau/automationhub/endpoints/httpx"
	appsvc "github.com/fsuhrau/automationhub/service/app"
	"github.com/fsuhrau/automationhub/storage/models"
)

// The master's own address, set by ServeHTTP for URLs without a proxy.
const (
	originalHostHeader = "X-Automationhub-Host"
	originalTLSHeader  = "X-Automationhub-Tls"
)

// uploadAddress is where a client reached the master: the headers of its
// request and the host and scheme the master saw.
type uploadAddress struct {
	header http.Header
	host   string
	tls    bool
}

// uploadAddress reads the master's own address from the headers ServeHTTP set.
func newUploadAddress(header http.Header) uploadAddress {
	return uploadAddress{header: header, host: header.Get(originalHostHeader), tls: header.Get(originalTLSHeader) == "1"}
}

// url is the upload address as the client sees the MCP endpoint.
func (a uploadAddress) url(project *models.Project, id string) string {
	scheme, host := "http", a.host
	if a.tls {
		scheme = "https"
	}
	prefix := "/api/" + project.Identifier + "/mcp"
	if v := a.header.Get("X-Forwarded-Proto"); v == "http" || v == "https" {
		scheme = v
	}
	if v := a.header.Get("X-Forwarded-Host"); v != "" {
		host = v
	}
	if v := a.header.Get("X-Forwarded-Prefix"); strings.HasPrefix(v, "/") {
		prefix = strings.TrimSuffix(v, "/")
	}
	return fmt.Sprintf("%s://%s%s/uploads/%s", scheme, host, prefix, id)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"message": err.Error()})
}

// claimStatus is the HTTP status of a refused claim of an upload.
func claimStatus(err error) int {
	switch {
	case errors.Is(err, errUploadNotFound):
		return http.StatusNotFound
	case errors.Is(err, errUploadCaller):
		return http.StatusForbidden
	}
	return http.StatusConflict
}

// ServeUpload receives the PUT of an upload started with begin_upload.
//
//nolint:revive // called by the upload route of endpoints/api with these arguments
func (h *Handler) ServeUpload(w http.ResponseWriter, r *http.Request, project *models.Project, caller, id string) {
	if caller == "" {
		caller = "anonymous"
	}
	if r.ContentLength > h.deps.MaxUpload {
		writeError(w, http.StatusRequestEntityTooLarge, tooLargeError{h.deps.MaxUpload})
		return
	}
	upload, err := h.uploads.claim(id, project.ID, caller)
	if err != nil {
		writeError(w, claimStatus(err), err)
		return
	}
	stored := false
	defer func() { h.uploads.finish(id, stored) }()
	stored = h.receiveUpload(w, r, upload)
}

// receiveUpload stores the body of the request as the bundle the upload is
// for and answers; it tells whether the upload is done.
func (h *Handler) receiveUpload(w http.ResponseWriter, r *http.Request, upload pendingUpload) bool {
	scope := appsvc.Scope{ProjectID: upload.projectID, AppID: upload.appID}
	if _, err := h.deps.Apps.App(r.Context(), scope); err != nil {
		writeError(w, httpx.Status(err), err)
		return false
	}
	bundle, exists, err := h.deps.storeUpload(r.Context(), uploadFile{scope: scope, name: upload.fileName, body: r.Body})
	var tooLarge tooLargeError
	if errors.As(err, &tooLarge) {
		writeError(w, http.StatusRequestEntityTooLarge, err)
		return false
	}
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return false
	}
	if exists {
		writeJSON(w, http.StatusOK, bundleResult(toBundleInfo(bundle), true))
		return true
	}
	writeJSON(w, http.StatusCreated, toBundleInfo(bundle))
	return true
}
