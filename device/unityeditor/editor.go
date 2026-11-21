package unityeditor

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
)

var (
	ErrorGetProjectVersion = errors.New("could not determine the project / unity version")

	editorVersionPattern = regexp.MustCompile(`m_EditorVersion:\s*([0-9]+\.[0-9]+\.[0-9]+[a-z0-9]*)`)
)

// GetUnityEditorPath is the editor binary of the version below editorPath.
// Only macOS is supported, other systems give an empty path.
func GetUnityEditorPath(editorPath string, projectVersion string) string {
	return editorBinary(runtime.GOOS, editorPath, projectVersion)
}

func editorBinary(goos, editorPath, projectVersion string) string {
	if goos == "darwin" {
		return filepath.Join(editorPath, projectVersion, "Unity.app/Contents/MacOS/Unity")
	}
	return ""
}

// GetUsedUnityVersion reads the editor version of the Unity project.
func GetUsedUnityVersion(projectDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(projectDir, "ProjectSettings", "ProjectVersion.txt"))
	if err != nil {
		return "", ErrorGetProjectVersion
	}
	match := editorVersionPattern.FindStringSubmatch(string(data))
	if len(match) < 2 {
		return "", ErrorGetProjectVersion
	}
	return match[1], nil
}
