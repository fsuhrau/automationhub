package apps

import (
	"os"
	"path/filepath"
)

var (
	AppStoragePath string
	TestDataPath   string
)

func init() {
	AppStoragePath = filepath.Join(os.Getenv("HOME"), ".automationhub", "store")
	TestDataPath = filepath.Join(os.Getenv("HOME"), ".automationhub", "data")
	os.MkdirAll(AppStoragePath, os.ModePerm)
	os.MkdirAll(TestDataPath, os.ModePerm)
}
