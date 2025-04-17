package apps

import (
	"os"
	"path/filepath"
)

var (
	AppStoragePath       string
	AppBundleStoragePath string
	TestDataPath         string
)

func init() {
	AppStoragePath = filepath.Join(os.Getenv("HOME"), ".automationhub", "store")
	AppBundleStoragePath = filepath.Join(os.Getenv("HOME"), ".automationhub", "bundles")
	TestDataPath = filepath.Join(os.Getenv("HOME"), ".automationhub", "data")
	os.MkdirAll(AppStoragePath, os.ModePerm)
	os.MkdirAll(AppBundleStoragePath, os.ModePerm)
	os.MkdirAll(TestDataPath, os.ModePerm)
}
