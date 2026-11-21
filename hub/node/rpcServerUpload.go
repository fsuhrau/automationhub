package node

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// downloadBufferSize is the size of one read of the app download.
const downloadBufferSize = 4 * 1024 * 1024

// UploadProgress tracks one app the node downloads from the master.
type UploadProgress struct {
	Request  *UploadAppRequest
	received atomic.Int64
	done     atomic.Bool
	errMu    sync.Mutex
	err      error
}

func (p *UploadProgress) setError(err error) {
	p.errMu.Lock()
	p.err = err
	p.errMu.Unlock()
	logrus.Errorf("app upload %d failed: %v", p.Request.AppID, err)
}

func (p *UploadProgress) Error() error {
	p.errMu.Lock()
	defer p.errMu.Unlock()
	return p.err
}

func (s *RPCNode) UploadApp(req *UploadAppRequest, resp *BoolResponse) error {
	logrus.Info("RPC: UploadApp")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	progress := &UploadProgress{Request: req}
	s.uploadRequests[req.AppID] = progress
	go s.downloadFile(progress)

	resp.Value = true
	return nil
}

// downloadFile downloads and stores the app; the outcome lands in progress.
func (s *RPCNode) downloadFile(progress *UploadProgress) {
	if err := s.fetchAndStore(progress); err != nil {
		progress.setError(err)
		return
	}
	progress.done.Store(true)
}

func (s *RPCNode) fetchAndStore(progress *UploadProgress) error {
	data, err := s.fetchBundle(progress)
	if err != nil {
		return err
	}
	if size := progress.Request.Size; size > 0 && int64(len(data)) != size {
		return fmt.Errorf("size mismatch: got %d bytes, expected %d", len(data), size)
	}
	logrus.Infof("Download complete, storing file")
	err = s.abm.StoreData(data, &AppBundleMetaData{
		Filename: progress.Request.Name,
		FileHash: progress.Request.Hash,
		FileSize: progress.Request.Size,
	})
	if err != nil {
		return fmt.Errorf("store app bundle: %w", err)
	}
	return nil
}

// fetchBundle downloads the app from the master, counting the bytes received.
func (s *RPCNode) fetchBundle(progress *UploadProgress) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, progress.Request.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if s.config.Auth.Token != nil {
		req.Header.Set("X-Auth-Token", s.authToken())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download: %s", resp.Status)
	}
	return readCounting(resp.Body, progress)
}

// readCounting reads r to the end and adds every chunk to progress.received.
func readCounting(r io.Reader, progress *UploadProgress) ([]byte, error) {
	data := make([]byte, 0, progress.Request.Size)
	buffer := make([]byte, downloadBufferSize)
	for {
		n, err := r.Read(buffer)
		data = append(data, buffer[:n]...)
		progress.received.Add(int64(n))
		if errors.Is(err, io.EOF) {
			return data, nil
		}
		if err != nil {
			return nil, fmt.Errorf("read file data: %w", err)
		}
	}
}

func (s *RPCNode) UploadAppProgress(req *UploadAppProgressRequest, resp *UploadAppProgressResponse) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	progress, exists := s.uploadRequests[req.AppID]
	if !exists {
		return fmt.Errorf("no upload in progress for app %d", req.AppID)
	}
	if err := progress.Error(); err != nil {
		delete(s.uploadRequests, req.AppID)
		return err
	}

	resp.AppID = req.AppID
	resp.DataReceived = progress.received.Load()
	if progress.done.Load() {
		// report completion only once the bundle is stored
		resp.DataReceived = progress.Request.Size
		delete(s.uploadRequests, req.AppID)
	} else if resp.DataReceived >= progress.Request.Size {
		// downloaded but not stored yet
		resp.DataReceived = progress.Request.Size - 1
	}
	return nil
}
