package utils

import (
	"os"
	"path/filepath"
)

const testDataPath = "testdata"

// GetExpected returns the expected file content
func GetExpected(name string) (string, error) {
	path := filepath.Join(testDataPath, name)
	bytes, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
