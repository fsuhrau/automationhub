package session

import (
	"fmt"
	"os"
	"path/filepath"
)

const createFlags = os.O_RDWR | os.O_CREATE | os.O_TRUNC

// Storage keeps the files of one session below Path/<unix time>/<session id>.
type Storage struct {
	Path      string
	SessionID string

	fs          FS
	clock       Clock
	deviceLog   File // nil if the log could not be opened
	storagePath string
}

// NewStorage creates the session directory and opens its remote device log.
// A directory or log that can't be created is not fatal: the log lines are
// dropped and storing a file reports the error.
func NewStorage(deps Deps, path, sessionID string) *Storage {
	fs, clock := deps.FS, deps.Clock
	s := &Storage{Path: path, SessionID: sessionID, fs: fs, clock: clock}
	s.storagePath = filepath.Join(path, fmt.Sprintf("%d", clock.Now().Unix()), sessionID)
	_ = fs.MkdirAll(s.storagePath, os.ModePerm)
	logPath := filepath.Join(s.storagePath, "remote_device.log")
	if f, err := fs.OpenFile(logPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm); err == nil {
		s.deviceLog = f
	}
	return s
}

// GetPath returns the directory of the session.
func (s *Storage) GetPath() string {
	return s.storagePath
}

// StoreSceneGraph writes the scene graph xml and returns its path.
func (s *Storage) StoreSceneGraph(data []byte) (string, error) {
	return s.store(fmt.Sprintf("%d_sceengraph.xml", s.clock.Now().Unix()), data)
}

// StoreImage writes a screenshot and returns its path.
func (s *Storage) StoreImage(content []byte) (string, error) {
	return s.store(fmt.Sprintf("%d_screenshot.png", s.clock.Now().Unix()), content)
}

func (s *Storage) store(name string, data []byte) (string, error) {
	filePath := filepath.Join(s.GetPath(), name)
	if err := s.fs.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return "", fmt.Errorf("create %s: %w", filepath.Dir(filePath), err)
	}
	file, err := s.fs.OpenFile(filePath, createFlags, 0o666)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", filePath, err)
	}
	_, err = file.Write(data)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return "", fmt.Errorf("write %s: %w", filePath, err)
	}
	return filePath, nil
}

// RemoteDeviceLog appends a line, prefixed with the time in nanoseconds, to the device log.
func (s *Storage) RemoteDeviceLog(line string) error {
	if s.deviceLog == nil {
		return nil
	}
	if _, err := fmt.Fprintf(s.deviceLog, "%d %s\n", s.clock.Now().UnixNano(), line); err != nil {
		return fmt.Errorf("write device log: %w", err)
	}
	return nil
}

// Close closes the device log.
func (s *Storage) Close() error {
	if s.deviceLog == nil {
		return nil
	}
	return s.deviceLog.Close()
}
