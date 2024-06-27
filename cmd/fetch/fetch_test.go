package fetch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateTargetPath_ValidPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-validate-target-path")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	err = validateTargetPath(tempDir)
	if err != nil {
		t.Errorf("validateTargetPath() returned an error for a valid path: %v", err)
	}
}

func TestValidateTargetPath_NonExistentPath(t *testing.T) {
	nonExistentPath := "/tmp/non-existent-path"
	err := validateTargetPath(nonExistentPath)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("validateTargetPath() returned an unexpected error for a non-existent path: %v", err)
	}
}

func TestValidateTargetPath_NotADirectory(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test-validate-target-path")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	err = validateTargetPath(tempFile.Name())
	if err.Error() != fmt.Sprintf("target path '%s' is not a directory", tempFile.Name()) {
		t.Errorf("validateTargetPath() returned an unexpected error for a file: %v", err)
	}
}

func TestValidateTargetPath_NotWriteable(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-validate-target-path")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Make the directory read-only
	err = os.Chmod(tempDir, 0444)
	if err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	if err = validateTargetPath(tempDir); err == nil {
		t.Fatalf("validateTargetPath() did not return an error for a non-writeable directory")
	}

	if err.Error() != fmt.Sprintf("target path '%s' is not writeable", tempDir) {
		t.Errorf("validateTargetPath() returned an unexpected error for a non-writeable directory: %v", err)
	}
}

func TestValidateTargetPath_StatError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-validate-target-path")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a path that is invalid or inaccessible
	invalidPath := filepath.Join(tempDir, "invalid-path")
	err = validateTargetPath(invalidPath)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("validateTargetPath() returned an unexpected error for an invalid path: %v", err)
	}
}
