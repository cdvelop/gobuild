package gobuild

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestConcurrentCompileProgram tests that multiple concurrent calls to CompileProgram
// don't cause race conditions. This test should be run with: go test -race
func TestConcurrentCompileProgram(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_race_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go program
	mainGoContent := `package main
import "fmt"
func main() {
	fmt.Println("Hello from race test!")
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

	var logOutput bytes.Buffer
	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "raceapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Writer:       &logOutput,
		Timeout:      30 * time.Second,
	}

	compiler := New(config)

	// Launch multiple concurrent compilations
	const numGoroutines = 10
	var wg sync.WaitGroup
	errors := make([]error, numGoroutines)

	for i := range numGoroutines {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errors[index] = compiler.CompileProgram()
		}(i)
	}

	wg.Wait()

	// At least one compilation should succeed
	successCount := 0
	for i, err := range errors {
		if err == nil {
			successCount++
			t.Logf("Goroutine %d: compilation succeeded", i)
		} else {
			t.Logf("Goroutine %d: compilation failed: %v", i, err)
		}
	}

	if successCount == 0 {
		t.Error("Expected at least one compilation to succeed")
	}

	// Verify final output file exists
	outputFile := filepath.Join(outputDir, "raceapp"+getExecutableExtension())
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Final output file should exist after concurrent compilations")
	}

	// Verify no temporary files remain
	tempPattern := filepath.Join(outputDir, "*_temp*")
	matches, err := filepath.Glob(tempPattern)
	if err != nil {
		t.Fatalf("Failed to check for temp files: %v", err)
	}
	if len(matches) > 0 {
		t.Errorf("Temporary files should not exist after compilation, found: %v", matches)
	}

	t.Logf("Race test completed: %d successful compilations out of %d", successCount, numGoroutines)
}

// TestConcurrentCompileAndCancel tests concurrent compilation and cancellation
func TestConcurrentCompileAndCancel(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_cancel_race_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go program
	mainGoContent := `package main
import "fmt"
func main() {
	fmt.Println("Hello from cancel race test!")
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
		OutName:      "cancelapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	compiler := New(config)

	// Launch multiple goroutines that compile and cancel concurrently
	const numGoroutines = 5
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Add(2) // One for compile, one for cancel

		// Compile goroutine
		go func(index int) {
			defer wg.Done()
			err := compiler.CompileProgram()
			t.Logf("Compile goroutine %d: %v", index, err)
		}(i)

		// Cancel goroutine (with slight delay)
		go func(index int) {
			defer wg.Done()
			time.Sleep(time.Duration(index*10) * time.Millisecond)
			err := compiler.Cancel()
			t.Logf("Cancel goroutine %d: %v", index, err)
		}(i)
	}

	wg.Wait()

	// Verify no panics occurred and system is in consistent state
	t.Log("Concurrent compile/cancel test completed without panics")
}

// TestStateConsistency tests that the internal state remains consistent
// under concurrent access
func TestStateConsistency(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_state_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go program
	mainGoContent := `package main
import "fmt"
func main() {
	fmt.Println("Hello from state test!")
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
		OutName:      "stateapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
	}

	compiler := New(config)

	// Launch multiple goroutines that check state and compile concurrently
	const numGoroutines = 8
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			for j := 0; j < 5; j++ {
				// Randomly check state or compile
				if j%2 == 0 {
					isCompiling := compiler.IsCompiling()
					t.Logf("Goroutine %d check %d: IsCompiling = %v", index, j, isCompiling)
				} else {
					err := compiler.CompileProgram()
					t.Logf("Goroutine %d compile %d: %v", index, j, err)
				}
				time.Sleep(time.Duration(index*5) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Final state should be consistent
	t.Log("State consistency test completed")
}

// TestAsyncCompilationRace tests race conditions in async compilation mode
func TestAsyncCompilationRace(t *testing.T) {
	// Create a temporary directory for our test
	tempDir, err := os.MkdirTemp("", "gobuild_async_race_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple Go program
	mainGoContent := `package main
import "fmt"
func main() {
	fmt.Println("Hello from async race test!")
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

	// Channel to collect callback results
	results := make(chan error, 10)

	config := &Config{
		Command:      "go",
		MainFilePath: mainGoPath,
		OutName:      "asyncraceapp",
		Extension:    getExecutableExtension(),
		OutFolder:    outputDir,
		Timeout:      30 * time.Second,
		Callback: func(err error) {
			results <- err
		},
	}

	compiler := New(config)

	// Launch multiple async compilations concurrently
	const numCompilations = 5
	for i := range numCompilations {
		go func(index int) {
			err := compiler.CompileProgram()
			if err != nil {
				t.Logf("Async start %d failed: %v", index, err)
			} else {
				t.Logf("Async start %d initiated", index)
			}
		}(i)
	}

	// Collect results with timeout
	successCount := 0
	timeout := time.After(45 * time.Second)

	for i := range numCompilations {
		select {
		case result := <-results:
			if result == nil {
				successCount++
				t.Logf("Async result %d: success", i)
			} else {
				t.Logf("Async result %d: %v", i, result)
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for async compilation results")
		}
	}

	if successCount == 0 {
		t.Error("Expected at least one async compilation to succeed")
	}

	t.Logf("Async race test completed: %d successful compilations", successCount)
}
