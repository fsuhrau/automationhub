package mcp

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsuhrau/automationhub/apperr"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ProtocolScreenshotArgs struct {
	ProtocolID uint   `json:"protocolId" jsonschema:"protocol id from get_test_run or get_protocol"`
	File       string `json:"file,omitempty" jsonschema:"file name of the screenshot, e.g. d5f7b6c4….png; default: the protocol's last screenshot"`
	MaxWidth   int    `json:"maxWidth,omitempty" jsonschema:"downscale to at most this width (default 1024, 0 keeps the default)"`
}

type screenshotInfo struct {
	File    string  `json:"file"`
	Message string  `json:"message,omitempty"`
	Runtime float64 `json:"runtime"`
}

// isScreenshotFile tells whether the file name is one of the screenshots,
// which are written to the test data folder and referenced from the data of a
// protocol entry.
func isScreenshotFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png", ".jpg", ".jpeg":
		return true
	}
	return false
}

// screenshotInfos lists the screenshots among the entries with data, in the
// order they were taken.
func screenshotInfos(entries []models.ProtocolEntry) []screenshotInfo {
	shots := make([]screenshotInfo, 0, len(entries))
	for _, e := range entries {
		if isScreenshotFile(e.Data) {
			shots = append(shots, screenshotInfo{File: e.Data, Message: e.Message, Runtime: e.Runtime})
		}
	}
	return shots
}

// pickScreenshot chooses the entry to serve: the newest one that refers to
// file, or without a file the newest screenshot. Only files an entry of the
// protocol refers to are served, so the tool can't read anything else from
// the data folder.
func pickScreenshot(entries []models.ProtocolEntry, in ProtocolScreenshotArgs) (*models.ProtocolEntry, error) {
	for i := len(entries) - 1; i >= 0; i-- {
		e := &entries[i]
		if (in.File != "" && e.Data == in.File) || (in.File == "" && isScreenshotFile(e.Data)) {
			return e, nil
		}
	}
	if in.File != "" {
		return nil, apperr.NotFound("not found: screenshot %s in protocol %d", in.File, in.ProtocolID)
	}
	return nil, apperr.NotFound("not found: protocol %d has no screenshots", in.ProtocolID)
}

// loadScreenshot reads and decodes the file of the entry from the data folder.
func (t *tools) loadScreenshot(e *models.ProtocolEntry) (image.Image, error) {
	name := e.Data
	if !isScreenshotFile(name) || name != filepath.Base(name) {
		return nil, fmt.Errorf("%s is not a screenshot", name)
	}
	data, err := os.ReadFile(filepath.Join(t.DataDir, name))
	if os.IsNotExist(err) {
		return nil, apperr.NotFound("not found: the file of screenshot %s is gone", name)
	}
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode screenshot %s: %w", name, err)
	}
	return img, nil
}

func (t *tools) getProtocolScreenshot(ctx context.Context, req *mcp.CallToolRequest, in ProtocolScreenshotArgs) (*mcp.CallToolResult, any, error) {
	entries, err := t.Protocols.ProtocolData(ctx, t.project.ID, in.ProtocolID)
	if err != nil {
		return nil, nil, err
	}
	entry, err := pickScreenshot(entries, in)
	if err != nil {
		return nil, nil, err
	}
	img, err := t.loadScreenshot(entry)
	if err != nil {
		return nil, nil, err
	}
	scaled, jpg, err := scaleToJPEG(img, in.MaxWidth)
	if err != nil {
		return nil, nil, err
	}
	return jpegResult(jpg, shot{entry, in.ProtocolID}.note(img, scaled))
}

// shot is the screenshot an entry of a protocol refers to.
type shot struct {
	entry      *models.ProtocolEntry
	protocolID uint
}

// note says which screenshot an image is and how it was scaled.
func (s shot) note(original, scaled image.Image) string {
	e := s.entry
	note := fmt.Sprintf("screenshot %s of protocol %d at %.1f s", e.Data, s.protocolID, e.Runtime)
	if e.Message != "" {
		note += ": " + e.Message
	}
	o, sc := original.Bounds(), scaled.Bounds()
	return note + fmt.Sprintf(" (%dx%d px, image %dx%d)", o.Dx(), o.Dy(), sc.Dx(), sc.Dy())
}
