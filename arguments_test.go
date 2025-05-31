package gobuild

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDynamicCompilingArguments verifies that CompilingArguments function is called
// dynamically on each compilation and that -X arguments are processed correctly
func TestDynamicCompilingArguments(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_arguments_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go program with version and secret variables
	mainGoContent := `package main

import "fmt"

var version string
var secret string

func main() {
	fmt.Printf("Version: %s Secret: %s\n", version, secret)
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(mainGoPath, []byte(mainGoContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write main.go: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Test counter to simulate dynamic behavior
	callCount := 0

	// Dynamic CompilingArguments function that changes behavior on each call
	dynamicArgsFunc := func() []string {
		callCount++
		switch callCount {
		case 1:
			// First call: no arguments (empty values expected)
			return []string{}
		case 2:
			// Second call: separated format (also works) - sets version only
			return []string{"-X", "main.version=1"}
		case 3:
			// Third call: combined format (preferred) - sets both variables
			return []string{"-X main.version=3", "-X main.secret=3"}
		default:
			return []string{}
		}
	}

	config := &Config{
		Command:            "go",
		MainFilePath:       mainGoPath,
		OutName:            "testapp",
		Extension:          getExecutableExtension(),
		CompilingArguments: dynamicArgsFunc,
		OutFolder:          outputDir,
		Timeout:            30 * time.Second,
	}

	compiler := New(config)
	// Test cases with expected outputs
	testCases := []struct {
		name           string
		expectedOutput string
		description    string
	}{
		{
			name:           "first_compilation_no_args",
			expectedOutput: "Version:  Secret:",
			description:    "No arguments - empty values expected",
		},
		{
			name:           "second_compilation_separated_format",
			expectedOutput: "Version: 1 Secret:",
			description:    "Separated -X format (works correctly) - should set version only",
		},
		{
			name:           "third_compilation_combined_format",
			expectedOutput: "Version: 3 Secret: 3",
			description:    "Combined -X format (preferred) - should set both variables",
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Test %d: %s", i+1, tc.description)

			// Compile the program
			err := compiler.CompileProgram()
			if err != nil {
				t.Fatalf("Compilation failed: %v", err)
			}

			// Verify output file exists
			outputFile := filepath.Join(outputDir, "testapp"+getExecutableExtension())
			if _, err := os.Stat(outputFile); err != nil {
				t.Fatalf("Output file does not exist: %v", err)
			}

			// Execute the compiled program and capture output
			cmd := exec.Command(outputFile)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Failed to execute compiled program: %v", err)
			}

			// Verify the output matches expected result
			actualOutput := strings.TrimSpace(string(output))
			if actualOutput != tc.expectedOutput {
				t.Errorf("Expected output: '%s', got: '%s'", tc.expectedOutput, actualOutput)
			}

			t.Logf("✓ Output matches expected: '%s'", actualOutput)

			// Clean up the output file for next iteration
			os.Remove(outputFile)
		})
	}

	// Verify that CompilingArguments was called the expected number of times
	if callCount != 3 {
		t.Errorf("Expected CompilingArguments to be called 3 times, but was called %d times", callCount)
	}

	t.Log("✓ CompilingArguments function demonstrated dynamic behavior across compilations")
}

// TestCompilingArgumentsLdflagsProcessing tests that -X arguments are correctly
// processed and added to ldflags in buildArguments method
func TestCompilingArgumentsLdflagsProcessing(t *testing.T) {
	config := &Config{
		Command:      "go",
		MainFilePath: "test.go",
		OutName:      "test",
		Extension:    "",
		OutFolder:    "/tmp",
	}

	compiler := New(config)

	testCases := []struct {
		name           string
		args           []string
		expectedInArgs []string
		description    string
	}{
		{
			name:           "no_arguments",
			args:           []string{},
			expectedInArgs: []string{"build", "-o", "/tmp/temp_test", "test.go"},
			description:    "No CompilingArguments should result in basic build command",
		},
		{
			name:           "single_X_combined",
			args:           []string{"-X main.version=1.0.0"},
			expectedInArgs: []string{"build", "-ldflags=-X main.version=1.0.0", "-o", "/tmp/temp_test", "test.go"},
			description:    "Single -X argument in correct format",
		},
		{
			name:           "multiple_X_combined",
			args:           []string{"-X main.version=1.0.0", "-X main.secret=mysecret"},
			expectedInArgs: []string{"build", "-ldflags=-X main.version=1.0.0 -X main.secret=mysecret", "-o", "/tmp/temp_test", "test.go"},
			description:    "Multiple -X arguments in correct format",
		},
		{
			name:           "mixed_arguments",
			args:           []string{"-race", "-X main.version=1.0.0", "-v"},
			expectedInArgs: []string{"build", "-race", "-v", "-ldflags=-X main.version=1.0.0", "-o", "/tmp/temp_test", "test.go"},
			description:    "Mix of -X and other build arguments",
		},
		{
			name:           "separated_X_arguments",
			args:           []string{"-X", "main.version=1.0.0"},
			expectedInArgs: []string{"build", "-ldflags=-X main.version=1.0.0", "-o", "/tmp/temp_test", "test.go"},
			description:    "Separated -X arguments should be combined correctly",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up CompilingArguments function for this test case
			config.CompilingArguments = func() []string {
				return tc.args
			}

			// Get build arguments
			buildArgs := compiler.buildArguments("temp_test")

			// Verify the build arguments match expected
			if len(buildArgs) != len(tc.expectedInArgs) {
				t.Errorf("Expected %d arguments, got %d", len(tc.expectedInArgs), len(buildArgs))
				t.Errorf("Expected: %v", tc.expectedInArgs)
				t.Errorf("Got:      %v", buildArgs)
				return
			}

			for i, expected := range tc.expectedInArgs {
				if buildArgs[i] != expected {
					t.Errorf("Argument %d: expected '%s', got '%s'", i, expected, buildArgs[i])
				}
			}

			t.Logf("✓ %s: Arguments processed correctly", tc.description)
		})
	}
}
