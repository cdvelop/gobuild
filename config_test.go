package gobuild

import (
	"bytes"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	var logOutput bytes.Buffer
	callbackCalled := false

	config := &Config{
		Command:               "go",
		MainFileRelativePath:  "web/main.go",
		OutName:               "myapp",
		Extension:             ".exe",
		CompilingArguments:    func() []string { return []string{"-X", "main.version=v1.0.0"} },
		OutFolderRelativePath: "dist",
		Logger:                &logOutput,
		Callback: func(err error) {
			callbackCalled = true
		},
		Timeout: 30 * time.Second,
	}

	if config.Command != "go" {
		t.Errorf("Expected Command to be 'go', got '%s'", config.Command)
	}

	if config.MainFileRelativePath != "web/main.go" {
		t.Errorf("Expected MainFileRelativePath to be 'web/main.go', got '%s'", config.MainFileRelativePath)
	}

	if config.OutName != "myapp" {
		t.Errorf("Expected OutName to be 'myapp', got '%s'", config.OutName)
	}

	if config.Extension != ".exe" {
		t.Errorf("Expected Extension to be '.exe', got '%s'", config.Extension)
	}

	if config.OutFolderRelativePath != "dist" {
		t.Errorf("Expected OutFolderRelativePath to be 'dist', got '%s'", config.OutFolderRelativePath)
	}

	if config.Logger != &logOutput {
		t.Error("Writer not properly assigned")
	}

	if config.Callback == nil {
		t.Error("Callback should not be nil")
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30 seconds, got %v", config.Timeout)
	}

	// Test CompilingArguments function
	if config.CompilingArguments == nil {
		t.Error("CompilingArguments should not be nil")
	} else {
		args := config.CompilingArguments()
		if len(args) != 2 || args[0] != "-X" || args[1] != "main.version=v1.0.0" {
			t.Errorf("CompilingArguments returned unexpected args: %v", args)
		}
	}

	// Test callback
	config.Callback(nil)
	if !callbackCalled {
		t.Error("Callback was not called")
	}
}

func TestCompileCallback(t *testing.T) {
	var receivedError error
	callback := CompileCallback(func(err error) {
		receivedError = err
	})

	// Test with nil error
	callback(nil)
	if receivedError != nil {
		t.Errorf("Expected nil error, got %v", receivedError)
	}

	// Test with actual error
	testError := &ConfigError{Message: "test error"}
	callback(testError)
	if receivedError != testError {
		t.Errorf("Expected %v, got %v", testError, receivedError)
	}
}

// ConfigError is a test error type
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
