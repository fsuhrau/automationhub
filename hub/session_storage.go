package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type SessionStorage struct {
	Path string
	SessionID string
	deviceLog *os.File
	storagePath string
}

func NewSessionStorage(Path, SessionID string) *SessionStorage {
	storage := &SessionStorage{
		Path: Path,
		SessionID: SessionID,
	}
	storage.createPath()
	filePath := filepath.Join(storage.GetPath(), "remote_device.log")
	storage.deviceLog, _ = os.OpenFile(filePath, os.O_RDWR | os.O_APPEND | os.O_CREATE, os.ModePerm)
	return storage
}

func (s *SessionStorage)createPath() {
	s.storagePath = filepath.Join(s.Path, fmt.Sprintf("%d", time.Now().Unix()), s.SessionID)
	os.MkdirAll(s.storagePath, os.ModePerm)
}

func (s* SessionStorage)GetPath() string {
	return s.storagePath
}

func (s* SessionStorage)StoreSceneGraph(data []byte) (string, error) {
	filePath := filepath.Join(s.GetPath(), fmt.Sprintf("%d_sceengraph.xml", time.Now().Unix()))
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func (s* SessionStorage)StoreImage(content []byte) (string, error) {
	filePath := filepath.Join(s.GetPath(), fmt.Sprintf("%d_screenshot.png", time.Now().Unix()))
	os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return "", err
	}
	return filePath, nil
}

func (s* SessionStorage) RemoteDeviceLog(line string) {
	_, err := s.deviceLog.WriteString(fmt.Sprintf("%d %s\n", time.Now().UnixNano(),  line))
	if err != nil {
		fmt.Printf("error...")
	}
}

func (s* SessionStorage) Close() error {
	return s.deviceLog.Close()
}
