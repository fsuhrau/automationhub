package mcp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsuhrau/automationhub/storage/apps"
	"github.com/fsuhrau/automationhub/storage/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
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

// screenshot files are written to the test data folder and referenced from a
// protocol entry's data
var screenshotExts = []string{".png", ".jpg", ".jpeg"}

func isScreenshotFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	for _, e := range screenshotExts {
		if ext == e {
			return true
		}
	}
	return false
}

// protocolScreenshots lists the screenshots of a protocol in the order they
// were taken.
func (t *tools) protocolScreenshots(protocolID uint) ([]screenshotInfo, error) {
	var entries []models.ProtocolEntry
	if err := t.DB.Where("test_protocol_id = ? AND data <> ''", protocolID).Order("id").Find(&entries).Error; err != nil {
		return nil, err
	}
	shots := make([]screenshotInfo, 0, len(entries))
	for _, e := range entries {
		if isScreenshotFile(e.Data) {
			shots = append(shots, screenshotInfo{File: e.Data, Message: e.Message, Runtime: e.Runtime})
		}
	}
	return shots, nil
}

func (t *tools) getProtocolScreenshot(ctx context.Context, req *mcp.CallToolRequest, in ProtocolScreenshotArgs) (*mcp.CallToolResult, any, error) {
	if _, err := t.findProtocol(in.ProtocolID); err != nil {
		return nil, nil, err
	}

	// only files a log entry of this protocol refers to are served, so the
	// tool can't be used to read anything else from the data folder
	var entry models.ProtocolEntry
	query := t.DB.Where("test_protocol_id = ?", in.ProtocolID)
	if in.File != "" {
		query = query.Where("data = ?", in.File)
	} else {
		ors := t.DB
		for i, ext := range screenshotExts {
			if i == 0 {
				ors = ors.Where("LOWER(data) LIKE ?", "%"+ext)
			} else {
				ors = ors.Or("LOWER(data) LIKE ?", "%"+ext)
			}
		}
		query = query.Where(ors)
	}
	err := query.Order("id desc").First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if in.File != "" {
			return nil, nil, fmt.Errorf("%w: screenshot %s in protocol %d", errNotFound, in.File, in.ProtocolID)
		}
		return nil, nil, fmt.Errorf("%w: protocol %d has no screenshots", errNotFound, in.ProtocolID)
	}
	if err != nil {
		return nil, nil, err
	}
	name := entry.Data
	if !isScreenshotFile(name) || name != filepath.Base(name) {
		return nil, nil, fmt.Errorf("%s is not a screenshot", name)
	}

	data, err := os.ReadFile(filepath.Join(apps.TestDataPath, name))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, fmt.Errorf("%w: the file of screenshot %s is gone", errNotFound, name)
	}
	if err != nil {
		return nil, nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("decode screenshot %s: %w", name, err)
	}
	scaled, jpg, err := scaleToJPEG(img, in.MaxWidth)
	if err != nil {
		return nil, nil, err
	}

	note := fmt.Sprintf("screenshot %s of protocol %d at %.1f s", name, in.ProtocolID, entry.Runtime)
	if entry.Message != "" {
		note += ": " + entry.Message
	}
	b := img.Bounds()
	note += fmt.Sprintf(" (%dx%d px, image %dx%d)", b.Dx(), b.Dy(), scaled.Bounds().Dx(), scaled.Bounds().Dy())
	return &mcp.CallToolResult{Content: []mcp.Content{
		&mcp.ImageContent{Data: jpg, MIMEType: "image/jpeg"},
		&mcp.TextContent{Text: note},
	}}, nil, nil
}
