package gobuild

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Integration tests that simulate real-world compilation scenarios

func TestIntegrationSuccessfulCompilation(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple, valid Go program
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World! Version 1")
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "testapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	compiler := New(config)

	// First compilation
	t.Log("Starting first compilation...")
	err = compiler.CompileProgram()
	if err != nil {
		t.Fatalf("First compilation failed: %v", err)
	}

	// Check that output file exists
	outputFile := filepath.Join(outputDir, "testapp"+getExecutableExtension())
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist after compilation: %s", outputFile)
	}

	// Get hash of first compilation
	firstHash, err := getFileHash(outputFile)
	if err != nil {
		t.Fatalf("Failed to get hash of first compilation: %v", err)
	}

	// Modify the source file
	modifiedContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World! Version 2 - Modified")
}
`
	if err := os.WriteFile(mainGoPath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify main.go: %v", err)
	}

	// Second compilation
	t.Log("Starting second compilation with modified source...")
	err = compiler.CompileProgram()
	if err != nil {
		t.Fatalf("Second compilation failed: %v", err)
	}

	// Check that output file still exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist after second compilation: %s", outputFile)
	}

	// Get hash of second compilation
	secondHash, err := getFileHash(outputFile)
	if err != nil {
		t.Fatalf("Failed to get hash of second compilation: %v", err)
	}

	// Verify the output file has changed
	if firstHash == secondHash {
		t.Errorf("Output file has not changed after modifying source code. First hash: %s, Second hash: %s", firstHash, secondHash)
	} else {
		t.Logf("Success: Output file changed after modification. First hash: %s, Second hash: %s", firstHash, secondHash)
	}

	// Verify temp file doesn't exist
	tempFile := filepath.Join(outputDir, "testapp_temp"+getExecutableExtension())
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Errorf("Temporary file still exists after successful compilation: %s", tempFile)
	}
}

func TestIntegrationCompilationWithErrors(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_integration_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid Go program first
	validContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "testapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	compiler := New(config)

	// First compilation (should succeed)
	t.Log("Starting compilation with valid code...")
	err = compiler.CompileProgram()
	if err != nil {
		t.Fatalf("First compilation with valid code failed: %v", err)
	}

	outputFile := filepath.Join(outputDir, "testapp"+getExecutableExtension())

	// Verify output file exists
	originalInfo, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Output file does not exist after successful compilation: %v", err)
	}

	// Get hash of successful compilation
	originalHash, err := getFileHash(outputFile)
	if err != nil {
		t.Fatalf("Failed to get hash of successful compilation: %v", err)
	}

	// Now introduce compilation errors
	errorContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!"  // Missing closing parenthesis
	undeclaredVariable = 42       // Undeclared variable
	someFunction()                // Undefined function
}
`
	if err := os.WriteFile(mainGoPath, []byte(errorContent), 0644); err != nil {
		t.Fatalf("Failed to write erroneous main.go: %v", err)
	}

	// Compilation with errors (should fail)
	t.Log("Starting compilation with erroneous code...")
	err = compiler.CompileProgram()
	if err == nil {
		t.Fatalf("Expected compilation to fail with erroneous code, but it succeeded")
	}
	t.Logf("Compilation failed as expected: %v", err)

	// Verify original output file is unchanged
	newInfo, err := os.Stat(outputFile)
	if err != nil {
		t.Errorf("Original output file was removed after failed compilation: %v", err)
	} else {
		// Check that file wasn't modified
		if !newInfo.ModTime().Equal(originalInfo.ModTime()) {
			t.Errorf("Original output file was modified after failed compilation")
		}

		// Verify hash is the same
		newHash, err := getFileHash(outputFile)
		if err != nil {
			t.Errorf("Failed to get hash after failed compilation: %v", err)
		} else if originalHash != newHash {
			t.Errorf("Original output file content changed after failed compilation")
		} else {
			t.Log("Success: Original output file remains unchanged after failed compilation")
		}
	}

	// Verify temp file doesn't exist (should be cleaned up)
	tempFile := filepath.Join(outputDir, "testapp_temp"+getExecutableExtension())
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Errorf("Temporary file exists after failed compilation: %s", tempFile)
	} else {
		t.Log("Success: Temporary file was properly cleaned up after failed compilation")
	}
}

func TestIntegrationAsyncCompilation(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_integration_async_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a valid Go program
	mainGoContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello from async compilation!")
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Channel to receive compilation result
	done := make(chan error, 1)

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "asyncapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
		Callback: func(err error) {
			done <- err
		},
	}

	compiler := New(config)

	// Start async compilation
	t.Log("Starting async compilation...")
	err = compiler.CompileProgram()
	if err != nil {
		t.Fatalf("Failed to start async compilation: %v", err)
	}

	// Wait for completion with timeout
	select {
	case compileErr := <-done:
		if compileErr != nil {
			t.Fatalf("Async compilation failed: %v", compileErr)
		}
		t.Log("Async compilation completed successfully")
	case <-time.After(45 * time.Second):
		t.Fatal("Async compilation timed out")
	}

	// Verify output file exists
	outputFile := filepath.Join(outputDir, "asyncapp"+getExecutableExtension())
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file does not exist after async compilation: %s", outputFile)
	}

	// Verify temp file doesn't exist
	tempFile := filepath.Join(outputDir, "asyncapp_temp"+getExecutableExtension())
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Errorf("Temporary file still exists after async compilation: %s", tempFile)
	}
}

// Helper functions

func getExecutableExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func TestIntegrationTempFileCleanupOnError(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_cleanup_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Go file with compilation errors
	errorContent := `package main

import "fmt"

func main() {
	fmt.Println("Missing quote)  // Syntax error
	undeclaredVar = 42           // Undeclared variable
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(errorContent), 0644); err != nil {
		t.Fatalf("Failed to create erroneous main.go: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "errorapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	compiler := New(config)

	// Attempt compilation (should fail)
	t.Log("Starting compilation with erroneous code...")
	err = compiler.CompileProgram()
	if err == nil {
		t.Fatalf("Expected compilation to fail, but it succeeded")
	}
	t.Logf("Compilation failed as expected: %v", err)

	// Verify that no output file was created
	outputFile := filepath.Join(outputDir, "errorapp"+getExecutableExtension())
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Errorf("Output file exists after failed compilation: %s", outputFile)
	}

	// Verify that no temporary file exists
	tempFile := filepath.Join(outputDir, "errorapp_temp"+getExecutableExtension())
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Errorf("Temporary file exists after failed compilation: %s", tempFile)
	} else {
		t.Log("Success: No temporary file left after failed compilation")
	}
}

func TestIntegrationTimeoutHandling(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_timeout_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a Go program that should compile quickly
	quickContent := `package main

import "fmt"

func main() {
	fmt.Println("Quick compilation test")
}
`
	mainGoPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(quickContent), 0644); err != nil {
		t.Fatalf("Failed to create main.go: %v", err)
	}

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "quickapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      100 * time.Millisecond, // Very short timeout to test timeout handling
	}

	compiler := New(config)

	// Attempt compilation (might timeout)
	t.Log("Starting compilation with very short timeout...")
	err = compiler.CompileProgram()

	// We don't assert on the result here because:
	// - If compilation is very fast, it might succeed
	// - If compilation takes longer than 100ms, it will timeout
	// Both are valid scenarios for this test

	if err != nil {
		t.Logf("Compilation failed (possibly due to timeout): %v", err)

		// If it failed, verify cleanup was done
		tempFile := filepath.Join(outputDir, "quickapp_temp"+getExecutableExtension())
		if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
			t.Errorf("Temporary file exists after timeout/failure: %s", tempFile)
		} else {
			t.Log("Success: Temporary file was cleaned up after timeout/failure")
		}
	} else {
		t.Log("Compilation succeeded despite short timeout (very fast system)")

		// Verify output file exists
		outputFile := filepath.Join(outputDir, "quickapp"+getExecutableExtension())
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file missing after successful compilation: %s", outputFile)
		}
	}
}

func TestIntegrationMultipleSuccessiveCompilations(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_successive_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	mainGoPath := filepath.Join(tempDir, "main.go")

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "successiveapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	compiler := New(config)
	outputFile := filepath.Join(outputDir, "successiveapp"+getExecutableExtension())
	tempFile := filepath.Join(outputDir, "successiveapp_temp"+getExecutableExtension())

	versions := []string{
		`package main
import "fmt"
func main() { fmt.Println("Version 1") }`,

		`package main
import "fmt"
func main() { fmt.Println("Version 2 with changes") }`,

		`package main
// This will cause compilation error
import "fmt"
func main() { 
	fmt.Println("Version 3")
	undeclaredVariable = 42  // Error here
}`,

		`package main
import "fmt"
func main() { fmt.Println("Version 4 - Fixed") }`,
	}

	var lastSuccessfulHash string

	for i, content := range versions {
		t.Logf("Testing compilation iteration %d", i+1)

		// Write the current version
		if err := os.WriteFile(mainGoPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write version %d: %v", i+1, err)
		}

		// Compile
		err = compiler.CompileProgram()

		if i == 2 { // Version 3 should fail
			if err == nil {
				t.Errorf("Expected version 3 to fail compilation, but it succeeded")
			} else {
				t.Logf("Version 3 failed as expected: %v", err)

				// Verify temp file was cleaned up
				if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
					t.Errorf("Temp file exists after failed compilation in iteration %d", i+1)
				}

				// Verify previous successful binary is still there and unchanged
				if lastSuccessfulHash != "" {
					if _, err := os.Stat(outputFile); os.IsNotExist(err) {
						t.Errorf("Previous successful binary was removed after failed compilation")
					} else {
						currentHash, err := getFileHash(outputFile)
						if err != nil {
							t.Errorf("Failed to get hash after failed compilation: %v", err)
						} else if currentHash != lastSuccessfulHash {
							t.Errorf("Previous successful binary was modified after failed compilation")
						} else {
							t.Log("Success: Previous binary preserved after failed compilation")
						}
					}
				}
			}
		} else { // Versions 1, 2, and 4 should succeed
			if err != nil {
				t.Fatalf("Expected version %d to succeed, but it failed: %v", i+1, err)
			}

			// Verify output file exists
			if _, err := os.Stat(outputFile); os.IsNotExist(err) {
				t.Errorf("Output file missing after successful compilation %d", i+1)
			} else {
				// Store hash for comparison
				hash, err := getFileHash(outputFile)
				if err != nil {
					t.Errorf("Failed to get hash for version %d: %v", i+1, err)
				} else {
					if lastSuccessfulHash != "" && hash == lastSuccessfulHash {
						t.Errorf("Binary didn't change between different source versions")
					}
					lastSuccessfulHash = hash
					t.Logf("Version %d compiled successfully, hash: %s", i+1, hash[:8])
				}
			}

			// Verify temp file was cleaned up
			if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
				t.Errorf("Temp file exists after successful compilation %d", i+1)
			}
		}
	}
}

// Test real scenario: existing valid Go file is modified and recompiled
func TestIntegrationRealScenarioFileModificationAndRecompilation(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_real_scenario_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	mainGoPath := filepath.Join(tempDir, "main.go")

	// Step 1: Create initial valid Go program
	initialContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World! This is version 1.0")
	fmt.Println("Initial compilation")
}
`
	if err := os.WriteFile(mainGoPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create initial main.go: %v", err)
	}

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "myapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	gb := New(config)
	// Step 2: First compilation
	t.Log("Performing initial compilation...")
	err = gb.CompileProgram()
	if err != nil {
		t.Fatalf("Initial compilation failed: %v", err)
	}

	// Verify first output file exists
	firstOutputPath := filepath.Join(outputDir, "myapp"+getExecutableExtension())
	if _, err := os.Stat(firstOutputPath); os.IsNotExist(err) {
		t.Fatalf("First output file should exist at: %s", firstOutputPath)
	}

	// Get hash of first output file
	firstHash, err := getFileHash(firstOutputPath)
	if err != nil {
		t.Fatalf("Failed to get hash of first output file: %v", err)
	}
	t.Logf("First compilation output hash: %s", firstHash)

	// Step 3: Modify the Go source file with different content
	modifiedContent := `package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello, World! This is version 2.0")
	fmt.Println("Modified compilation with timestamp:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("This version has more functionality")
}
`
	t.Log("Modifying source file...")
	if err := os.WriteFile(mainGoPath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify main.go: %v", err)
	}

	// Small delay to ensure different compilation time
	time.Sleep(100 * time.Millisecond)
	// Step 4: Second compilation with modified content
	t.Log("Performing second compilation with modified content...")
	err = gb.CompileProgram()
	if err != nil {
		t.Fatalf("Second compilation failed: %v", err)
	}

	// Step 5: Verify second output file exists and is different
	if _, err := os.Stat(firstOutputPath); os.IsNotExist(err) {
		t.Fatalf("Second output file should exist at: %s", firstOutputPath)
	}

	secondHash, err := getFileHash(firstOutputPath)
	if err != nil {
		t.Fatalf("Failed to get hash of second output file: %v", err)
	}
	t.Logf("Second compilation output hash: %s", secondHash)

	// Step 6: Verify that the output files are different
	if firstHash == secondHash {
		t.Error("Output files should be different after source modification, but hashes are identical")
		t.Errorf("First hash: %s", firstHash)
		t.Errorf("Second hash: %s", secondHash)
	} else {
		t.Log("✓ Output files are correctly different after source modification")
	}

	// Step 7: Verify no temporary files remain
	tempPattern := filepath.Join(outputDir, "myapp_temp*")
	matches, err := filepath.Glob(tempPattern)
	if err != nil {
		t.Fatalf("Failed to check for temp files: %v", err)
	}
	if len(matches) > 0 {
		t.Errorf("Temporary files should not exist after successful compilation, found: %v", matches)
	} else {
		t.Log("✓ No temporary files remain after compilation")
	}
}

// Test real scenario: Go file with errors, compilation fails, original file unchanged, temp file cleaned up
func TestIntegrationRealScenarioErrorHandlingAndCleanup(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_error_scenario_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outputDir := filepath.Join(tempDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	mainGoPath := filepath.Join(tempDir, "main.go")

	// Step 1: Create a valid Go file first (to establish baseline)
	validContent := `package main

import "fmt"

func main() {
	fmt.Println("This is a valid Go program")
}
`
	if err := os.WriteFile(mainGoPath, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to create valid main.go: %v", err)
	}

	// Get hash of the valid source file
	originalSourceHash, err := getFileHash(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to get hash of original source file: %v", err)
	}
	t.Logf("Original source file hash: %s", originalSourceHash)

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "errorapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	gb := New(config)
	// Step 2: First successful compilation to create baseline
	t.Log("Performing initial successful compilation...")
	err = gb.CompileProgram()
	if err != nil {
		t.Fatalf("Initial compilation should succeed: %v", err)
	}

	outputPath := filepath.Join(outputDir, "errorapp"+getExecutableExtension())
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Initial output file should exist at: %s", outputPath)
	}

	// Get modification time of initial output
	initialOutputInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to get initial output file info: %v", err)
	}
	initialModTime := initialOutputInfo.ModTime()
	t.Logf("Initial output file modification time: %v", initialModTime)

	// Step 3: Modify source file to introduce compilation errors
	errorContent := `package main

import "fmt"

func main() {
	// Syntax error: missing quote
	fmt.Println("This will cause a compilation error
	
	// Undefined variable error
	fmt.Println(undefinedVariable)
	
	// Missing closing brace
`
	t.Log("Introducing compilation errors in source file...")
	if err := os.WriteFile(mainGoPath, []byte(errorContent), 0644); err != nil {
		t.Fatalf("Failed to write error content to main.go: %v", err)
	}

	// Small delay to ensure different modification time
	time.Sleep(100 * time.Millisecond)
	// Step 4: Attempt compilation with erroneous content
	t.Log("Attempting compilation with erroneous content (should fail)...")
	err = gb.CompileProgram()
	if err == nil {
		t.Error("Compilation should have failed with syntax errors, but succeeded")
	} else {
		t.Logf("✓ Compilation correctly failed with error: %v", err)
	}

	// Step 5: Verify original output file is unchanged
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Original output file should still exist after failed compilation")
	} else {
		currentOutputInfo, err := os.Stat(outputPath)
		if err != nil {
			t.Fatalf("Failed to get current output file info: %v", err)
		}
		currentModTime := currentOutputInfo.ModTime()

		if !currentModTime.Equal(initialModTime) {
			t.Errorf("Output file should not have been modified after failed compilation")
			t.Errorf("Initial mod time: %v", initialModTime)
			t.Errorf("Current mod time: %v", currentModTime)
		} else {
			t.Log("✓ Original output file correctly unchanged after failed compilation")
		}
	}

	// Step 6: Verify source file has the error content (was modified as expected)
	currentSourceHash, err := getFileHash(mainGoPath)
	if err != nil {
		t.Fatalf("Failed to get hash of current source file: %v", err)
	}

	if currentSourceHash == originalSourceHash {
		t.Error("Source file should have been modified (this test modifies it to introduce errors)")
	} else {
		t.Log("✓ Source file correctly shows it was modified")
	}

	// Step 7: Verify no temporary files exist in output directory
	tempPattern := filepath.Join(outputDir, "errorapp_temp*")
	matches, err := filepath.Glob(tempPattern)
	if err != nil {
		t.Fatalf("Failed to check for temp files: %v", err)
	}
	if len(matches) > 0 {
		t.Errorf("Temporary files should have been cleaned up after failed compilation, found: %v", matches)
	} else {
		t.Log("✓ Temporary files correctly cleaned up after failed compilation")
	}

	// Step 8: Additional check - verify temp files don't exist with any extension
	allTempPattern := filepath.Join(outputDir, "*_temp*")
	allMatches, err := filepath.Glob(allTempPattern)
	if err != nil {
		t.Fatalf("Failed to check for any temp files: %v", err)
	}
	if len(allMatches) > 0 {
		t.Errorf("No temporary files should exist in output directory after failed compilation, found: %v", allMatches)
	} else {
		t.Log("✓ Output directory correctly contains no temporary files")
	}
}
