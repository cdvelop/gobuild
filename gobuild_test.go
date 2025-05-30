package gobuild

import (
	"bytes"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	config := &Config{
		Command:      "go",
		MainFilePath: "main.go",
		OutName:      "test",
		Extension:    ".exe",
		OutFolder:    "build",
	}
	gb := New(config)

	if gb == nil {
		t.Fatal("New() returned nil")
	}
	if gb.config != config {
		t.Error("Config not properly assigned")
	}

	if gb.outFileName != "test.exe" {
		t.Errorf("Expected outFileName to be 'test.exe', got '%s'", gb.outFileName)
	}

	if gb.outTempFileName != "test_temp.exe" {
		t.Errorf("Expected outTempFileName to be 'test_temp.exe', got '%s'", gb.outTempFileName)
	}
	// Test default timeout
	if gb.config.Timeout != 5*time.Second {
		t.Errorf("Expected default timeout to be 5 seconds, got %v", gb.config.Timeout)
	}
}

func TestNewWithCustomTimeout(t *testing.T) {
	config := &Config{
		Command:      "go",
		MainFilePath: "main.go",
		OutName:      "test",
		Extension:    ".exe",
		OutFolder:    "build",
		Timeout:      10 * time.Second,
	}
	gb := New(config)
	if gb.config.Timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10 seconds, got %v", gb.config.Timeout)
	}
}

func TestCompileProgramSync(t *testing.T) {
	var logOutput bytes.Buffer
	config := &Config{
		Command:      "echo", // Use echo command for testing
		MainFilePath: "test",
		OutName:      "test",
		Extension:    "",
		OutFolder:    ".",
		Log:          &logOutput,
		Timeout:      1 * time.Second,
	}

	gb := New(config)

	// This should return immediately since no callback is set
	err := gb.CompileProgram()
	if err == nil {
		t.Log("Sync compilation completed without error")
	} else {
		// Error is expected since we're using echo instead of go build
		t.Logf("Expected error for echo command: %v", err)
	}
}

func TestCompileProgramAsync(t *testing.T) {
	var logOutput bytes.Buffer
	callbackCalled := make(chan error, 1)

	config := &Config{
		Command:      "echo",
		MainFilePath: "test",
		OutName:      "test",
		Extension:    "",
		OutFolder:    ".",
		Log:          &logOutput,
		Timeout:      1 * time.Second,
		Callback: func(err error) {
			callbackCalled <- err
		},
	}

	gb := New(config)

	// This should return immediately and run async
	err := gb.CompileProgram()
	if err != nil {
		t.Errorf("Async compilation should return nil immediately, got: %v", err)
	}

	// Wait for callback
	select {
	case callbackErr := <-callbackCalled:
		t.Logf("Callback called with error: %v", callbackErr)
	case <-time.After(2 * time.Second):
		t.Error("Callback was not called within timeout")
	}
}
