package helpers

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetFixturePath returns the absolute path to a fixture file
func GetFixturePath(name string) string {
	_, currentFile, _, _ := runtime.Caller(0)
	// Go up from tests/helpers to repo root, then into fixtures
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(currentFile)))
	return filepath.Join(repoRoot, "fixtures", name)
}

// FixtureExists checks if a fixture file exists
func FixtureExists(name string) bool {
	path := GetFixturePath(name)
	_, err := os.Stat(path)
	return err == nil
}

// ReadFixture reads the contents of a fixture file
func ReadFixture(name string) (string, error) {
	path := GetFixturePath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Common fixture names
const (
	FixtureSimpleNginx  = "simple-nginx.yaml"
	FixtureMultiService = "multi-service.yaml"
)

// TestRepository is the repository URL configured in test Helm values
const TestRepository = "https://github.com/lissto-dev/e2e"

// TestNamespaces
const (
	LisstoSystemNamespace = "lissto-system"
	GlobalNamespace       = "lissto-global"
	UserNamespacePrefix   = "dev-"
)

// GetUserNamespace returns the expected user namespace for a given user
func GetUserNamespace(username string) string {
	return UserNamespacePrefix + username
}
