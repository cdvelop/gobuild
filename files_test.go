package gobuild

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestUnobservedFiles(t *testing.T) {
	config := &Config{
		OutName:   "myapp",
		Extension: ".exe",
		Logger:    io.Discard,
	}

	gb := New(config)
	files := gb.UnobservedFiles()

	expected := []string{"myapp.exe", "myapp_temp.exe"}

	if len(files) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(files))
		return
	}

	for i, file := range files {
		if file != expected[i] {
			t.Errorf("File %d: expected '%s', got '%s'", i, expected[i], file)
		}
	}
}

func TestUnobservedFilesWithoutExtension(t *testing.T) {
	config := &Config{
		OutName:   "myapp",
		Extension: "",
		Logger:    io.Discard,
	}

	gb := New(config)
	files := gb.UnobservedFiles()

	expected := []string{"myapp", "myapp_temp"}

	if len(files) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(files))
		return
	}

	for i, file := range files {
		if file != expected[i] {
			t.Errorf("File %d: expected '%s', got '%s'", i, expected[i], file)
		}
	}
}

func TestRenameOutputFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gobuild_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &Config{
		OutName:   "testapp",
		Extension: ".exe",
		OutFolder: tempDir,
		Logger:    io.Discard,
	}

	gb := New(config)

	// Create a temporary file to rename
	tempFile := filepath.Join(tempDir, gb.outTempFileName)
	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	file.Close()
	// Test the rename operation
	err = gb.RenameOutputFile()
	if err != nil {
		t.Errorf("renameOutputFile failed: %v", err)
	}

	// Check that the final file exists
	finalFile := filepath.Join(tempDir, gb.outFileName)
	if _, err := os.Stat(finalFile); os.IsNotExist(err) {
		t.Errorf("Final file does not exist: %s", finalFile)
	}

	// Check that the temp file no longer exists
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Errorf("Temp file still exists: %s", tempFile)
	}
}

func TestRenameOutputFileNonexistentSource(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gobuild_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &Config{
		OutName:   "testapp",
		Extension: ".exe",
		OutFolder: tempDir,
		Logger:    io.Discard,
	}

	gb := New(config)
	// Try to rename a file that doesn't exist
	err = gb.RenameOutputFile()
	if err == nil {
		t.Error("Expected error when renaming nonexistent file, got nil")
	}
}

func TestRenameOutputFileInvalidDestination(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gobuild_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &Config{
		OutName:   "testapp",
		Extension: ".exe",
		OutFolder: "/nonexistent/path",
		Logger:    io.Discard,
	}
	gb := New(config)

	// Create a source file in temp directory but try to move to nonexistent destination
	tempFileName := "testapp_temp.exe"
	sourcePath := filepath.Join(tempDir, tempFileName)
	file, err := os.Create(sourcePath)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}
	file.Close()

	err = gb.RenameOutputFileFrom(tempFileName)
	if err == nil {
		t.Error("Expected error when renaming to invalid destination, got nil")
	}
}
