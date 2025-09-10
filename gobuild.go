package gobuild

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// compilation represents an active compilation process
type compilation struct {
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	done      chan error
	tempFile  string
	startTime time.Time
}

// GoBuild represents a Go compiler instance
type GoBuild struct {
	config *Config

	// Thread-safe state
	mu              sync.RWMutex
	active          *compilation
	outFileName     string // eg: main.exe, app
	outTempFileName string // eg: app_temp.exe

}

// New creates a new GoBuild instance with the given configuration
func New(c *Config) *GoBuild {
	// Set default timeout if not specified
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}

	return &GoBuild{
		config:          c,
		outFileName:     c.OutName + c.Extension,
		outTempFileName: c.OutName + "_temp" + c.Extension,
	}
}

// CompileProgram compiles the Go program
// If a callback is configured, it runs asynchronously and returns immediately
// Otherwise, it runs synchronously and returns the compilation result
// Thread-safe: cancels any previous compilation automatically
func (h *GoBuild) CompileProgram() error {
	h.mu.Lock()

	// Cancel any active compilation
	if h.active != nil {
		h.active.cancel()
		// Don't wait for it to finish, just move on
		h.active = nil
	}

	// Create new compilation context
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout)

	// Generate unique temp file name to avoid conflicts
	tempFileName := fmt.Sprintf("%s_temp_%d%s",
		h.config.OutName,
		time.Now().UnixNano(),
		h.config.Extension)

	comp := &compilation{
		cancel:    cancel,
		done:      make(chan error, 1),
		tempFile:  tempFileName,
		startTime: time.Now(),
	}

	h.active = comp
	h.mu.Unlock()

	// If callback is defined, run asynchronously
	if h.config.Callback != nil {
		go func() {
			err := h.compileSync(ctx, comp)
			h.config.Callback(err)

			// Clean up active compilation
			h.mu.Lock()
			if h.active == comp {
				h.active = nil
			}
			h.mu.Unlock()
		}()
		return nil
	}

	// Run synchronously
	err := h.compileSync(ctx, comp)

	// Clean up
	h.mu.Lock()
	if h.active == comp {
		h.active = nil
	}
	h.mu.Unlock()

	return err
}

// Cancel cancels any active compilation
func (h *GoBuild) Cancel() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.active != nil {
		h.active.cancel()
		h.active = nil
		return nil
	}

	return nil // No active compilation to cancel
}

// IsCompiling returns true if there's an active compilation
func (h *GoBuild) IsCompiling() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.active != nil
}

// BuildArguments returns the build arguments that would be used for compilation
// This is exposed for testing purposes
func (h *GoBuild) BuildArguments() []string {
	return h.buildArguments(h.outTempFileName)
}

// RenameOutputFile renames the default temporary output file to the final output file
// This is exposed for testing purposes
func (h *GoBuild) RenameOutputFile() error {
	return h.renameOutputFile(h.outTempFileName)
}

// RenameOutputFileFrom renames a specific temporary file to the final output file
// This is exposed for testing purposes
func (h *GoBuild) RenameOutputFileFrom(tempFileName string) error {
	return h.renameOutputFile(tempFileName)
}

// MainOutputFileNameWithExtension returns the output filename with extension (e.g., "main.wasm", "app.exe")
func (h *GoBuild) MainOutputFileNameWithExtension() string {
	return h.outFileName
}

// MainInputFileRelativePath eg: cmd/main.go
func (h *GoBuild) MainInputFileRelativePath() string {
	return h.config.MainInputFileRelativePath
}
