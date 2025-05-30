package gobuild

import (
	"os/exec"
	"time"
)

// GoBuild represents a Go compiler instance
type GoBuild struct {
	*Config
	Cmd             *exec.Cmd
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
		Config:          c,
		Cmd:             &exec.Cmd{},
		outFileName:     c.OutName + c.Extension,
		outTempFileName: c.OutName + "_temp" + c.Extension,
	}
}

// CompileProgram compiles the Go program
// If a callback is configured, it runs asynchronously and returns immediately
// Otherwise, it runs synchronously and returns the compilation result
func (h *GoBuild) CompileProgram() error {
	// If callback is defined, run asynchronously
	if h.Callback != nil {
		go func() {
			err := h.compileSync()
			h.Callback(err)
		}()
		return nil
	}

	// Run synchronously
	return h.compileSync()
}
