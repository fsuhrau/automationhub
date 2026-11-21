package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/fsuhrau/automationhub/storage/models"
)

// UploadBundle uploads the app binary at path and returns the created binary.
func (c *Client) UploadBundle(ctx context.Context, path string) (*models.AppBinary, error) {
	body, contentType, err := multipartFile("test_target", path)
	if err != nil {
		return nil, err
	}
	var app models.AppBinary
	if err := c.send(ctx, rawCall{http.MethodPost, "/upload", contentType, body}, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *Client) UpdateApp(ctx context.Context, binary *models.AppBinary) error {
	var updated models.AppBinary
	return c.sendJSON(ctx, jsonCall{http.MethodPost, fmt.Sprintf("/bundle/%d", binary.ID), binary, &updated})
}

// multipartFile is a multipart form with the file at path in the field and
// its content type (which holds the boundary).
func multipartFile(field, path string) (io.Reader, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open bundle: %w", err)
	}
	defer func() { _ = f.Close() }()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	part, _ := w.CreateFormFile(field, f.Name()) // writes to a bytes.Buffer never fail
	if _, err := io.Copy(part, f); err != nil {
		return nil, "", fmt.Errorf("read bundle: %w", err)
	}
	_ = w.Close() // writes to a bytes.Buffer never fail
	return &b, w.FormDataContentType(), nil
}
