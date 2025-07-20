package gobuild

import (
	"bytes"
	"testing"
)

func TestBuildArguments(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected []string
	}{
		{
			name: "basic build arguments",
			config: &Config{
				MainFilePath: "main.go",
				OutFolder:    "build",
				OutName:      "app",
				Extension:    ".exe",
			},
			expected: []string{"build", "-o", "build/app_temp.exe", "main.go"},
		},
		{
			name: "with ldflags",
			config: &Config{
				MainFilePath: "main.go",
				OutFolder:    "build",
				OutName:      "app",
				Extension:    ".exe",
				CompilingArguments: func() []string {
					return []string{"-X", "main.version=v1.0.0"}
				},
			},
			expected: []string{"build", "-ldflags=-X main.version=v1.0.0", "-o", "build/app_temp.exe", "main.go"},
		},
		{
			name: "with regular arguments and ldflags",
			config: &Config{
				MainFilePath: "main.go",
				OutFolder:    "build",
				OutName:      "app",
				Extension:    ".exe",
				CompilingArguments: func() []string {
					return []string{"-v", "-X", "main.version=v1.0.0", "-tags", "prod"}
				},
			},
			expected: []string{"build", "-v", "-tags", "prod", "-ldflags=-X main.version=v1.0.0", "-o", "build/app_temp.exe", "main.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gb := New(tt.config)
			args := gb.BuildArguments()

			if len(args) != len(tt.expected) {
				t.Errorf("Expected %d arguments, got %d. Expected: %v, Got: %v",
					len(tt.expected), len(args), tt.expected, args)
				return
			}

			for i, arg := range args {
				if arg != tt.expected[i] {
					t.Errorf("Argument %d: expected '%s', got '%s'", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestCompileSyncWithInvalidCommand(t *testing.T) {
	var logOutput bytes.Buffer
	config := &Config{
		Command:      "nonexistentcommand",
		MainFilePath: "main.go",
		OutName:      "test",
		Extension:    ".exe",
		OutFolder:    "build",
		Writer:       &logOutput,
	}
	gb := New(config)
	err := gb.CompileProgram()

	if err == nil {
		t.Error("Expected error for nonexistent command, got nil")
	}

	if err != nil {
		t.Logf("Expected error received: %v", err)
	}
}

func TestCompileSyncArgumentParsing(t *testing.T) {
	config := &Config{
		MainFilePath: "main.go",
		OutName:      "test",
		Extension:    "",
		OutFolder:    ".",
		CompilingArguments: func() []string {
			return []string{
				"-v",                    // regular argument
				"-race",                 // regular argument
				"-X", "main.version=v1", // ldflags
				"-tags", "integration", // regular argument
				"-X", "main.build=123", // more ldflags
			}
		},
	}

	gb := New(config)
	args := gb.BuildArguments()
	// Verify structure: build + regular args + ldflags + output + source
	expectedStructure := []string{
		"build",
		"-v",
		"-race",
		"-tags",
		"integration",
		"-ldflags=-X main.version=v1 -X main.build=123",
		"-o",
		"test_temp",
		"main.go",
	}

	if len(args) != len(expectedStructure) {
		t.Errorf("Expected %d arguments, got %d. Got: %v", len(expectedStructure), len(args), args)
		return
	}

	for i, expected := range expectedStructure {
		if args[i] != expected {
			t.Errorf("Argument %d: expected '%s', got '%s'", i, expected, args[i])
		}
	}
}
