package node

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/fsuhrau/automationhub/app"
	"github.com/sirupsen/logrus"
)

// UploadStallTimeout aborts an app upload to a node when no progress was made.
const UploadStallTimeout = 2 * time.Minute

// UploadApp has the node download the app from the master and waits until it
// is stored.
func (c *RPCClient) UploadApp(ctx context.Context, parameter *app.Parameter) error {
	request := c.uploadRequest(parameter)
	if err := c.forwardOK("UploadApp", request); err != nil {
		return err
	}
	return c.waitForUpload(ctx, request)
}

func (c *RPCClient) uploadRequest(parameter *app.Parameter) *UploadAppRequest {
	filename := filepath.Base(parameter.App.AppPath)
	fileURL, _ := url.JoinPath("http://"+c.masterURL, "upload", filename)
	return &UploadAppRequest{
		AppID:      int32(parameter.App.AppBinaryID),
		Identifier: parameter.Identifier,
		Name:       filename,
		Hash:       parameter.App.Hash,
		Size:       int64(parameter.App.Size),
		URL:        fileURL,
	}
}

// waitForUpload polls the node's progress until the app is complete, the
// context ends or the upload stalls for uploadStall.
func (c *RPCClient) waitForUpload(ctx context.Context, request *UploadAppRequest) error {
	ticker := time.NewTicker(c.uploadInterval)
	defer ticker.Stop()
	lastReceived := int64(-1)
	lastProgress := time.Now()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
		progress, err := call[UploadAppProgressResponse](c, "UploadAppProgress", &UploadAppProgressRequest{AppID: request.AppID})
		if err != nil {
			return fmt.Errorf("upload app to node: %w", err)
		}
		if progress.DataReceived >= request.Size {
			logrus.Info("RPCNode.UploadApp complete")
			return nil
		}
		if progress.DataReceived != lastReceived {
			lastReceived, lastProgress = progress.DataReceived, time.Now()
		} else if time.Since(lastProgress) > c.uploadStall {
			return fmt.Errorf("upload app to node: no progress for %v (%d/%d bytes)", c.uploadStall, lastReceived, request.Size)
		}
	}
}
