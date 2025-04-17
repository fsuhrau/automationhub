package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrAppNotFound = fmt.Errorf("AppBundleMetaData not found")
)

type AppBundleMetaData struct {
	Filename string `json:"filename"`
	FileHash string `json:"file_hash"`
	FileSize int64  `json:"file_size"`
	FilePath string `json:"-"`
}

type AppBundleManager struct {
	dataDir  string
	metadata map[string]*AppBundleMetaData
	mutex    sync.Mutex
}

func NewAppBundleManager(dataDir string) (*AppBundleManager, error) {
	m := &AppBundleManager{
		dataDir:  dataDir,
		metadata: make(map[string]*AppBundleMetaData),
	}

	if err := m.loadMetadata(); err != nil {
		return nil, err
	}

	return m, nil
}

func (dm *AppBundleManager) loadMetadata() error {
	files, err := os.ReadDir(dm.dataDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".meta" {
			filePath := filepath.Join(dm.dataDir, file.Name())
			fileData, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read metaData file: %v", err)
			}

			var metaData AppBundleMetaData
			if err := json.Unmarshal(fileData, &metaData); err != nil {
				return fmt.Errorf("failed to unmarshal metaData: %v", err)
			}
			metaData.FilePath = filepath.Join(dm.dataDir, metaData.Filename)
			dm.metadata[metaData.FileHash] = &metaData
		}
	}

	return nil
}

func (dm *AppBundleManager) StoreData(appData []byte, metadata *AppBundleMetaData) error {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	// Save the binary data
	binaryFilePath := filepath.Join(dm.dataDir, metadata.Filename)
	if err := os.WriteFile(binaryFilePath, appData, os.ModePerm); err != nil {
		return fmt.Errorf("failed to save binary data: %v", err)
	}

	// Save the AppBundleMetaData as a JSON file
	metadataFilePath := filepath.Join(dm.dataDir, fmt.Sprintf("%s.meta", metadata.FileHash))
	metadata.FilePath = binaryFilePath
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal AppBundleMetaData: %v", err)
	}
	if err := os.WriteFile(metadataFilePath, metadataBytes, os.ModePerm); err != nil {
		return fmt.Errorf("failed to save AppBundleMetaData: %v", err)
	}

	dm.metadata[metadata.FileHash] = metadata
	return nil
}

func (dm *AppBundleManager) GetAppParameter(hash string) (*AppBundleMetaData, error) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	metadata, exists := dm.metadata[hash]
	if !exists {
		return nil, ErrAppNotFound
	}

	return metadata, nil
}
