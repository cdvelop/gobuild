package gobuild

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

// synchronizedWriter wraps an io.Writer to make it safe for concurrent use
type synchronizedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (sw *synchronizedWriter) Write(p []byte) (n int, err error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

// compileSync performs the actual compilation synchronously with context timeout
func (h *GoBuild) compileSync(ctx context.Context, comp *compilation) error {
	var this = errors.New("compileSync")
	buildArgs := h.buildArguments(comp.tempFile)
	comp.cmd = exec.CommandContext(ctx, h.config.Command, buildArgs...)

	// Set environment variables if provided
	if len(h.config.Env) > 0 {
		comp.cmd.Env = append(os.Environ(), h.config.Env...)
	}

	stderr, err := comp.cmd.StderrPipe()
	if err != nil {
		return errors.Join(this, err)
	}

	stdout, err := comp.cmd.StdoutPipe()
	if err != nil {
		return errors.Join(this, err)
	}
	err = comp.cmd.Start()
	if err != nil {
		return errors.Join(this, err)
	}
	if h.config.Writer != nil {
		// Create a synchronized writer to handle concurrent stdout/stderr writes
		syncWriter := &synchronizedWriter{w: h.config.Writer}

		go io.Copy(syncWriter, stderr)
		go io.Copy(syncWriter, stdout)
	}

	if err := comp.cmd.Wait(); err != nil {
		// Clean up temporary file if compilation failed
		h.cleanupTempFile(comp.tempFile)
		return errors.Join(this, err)
	}

	return h.renameOutputFile(comp.tempFile)
}

// buildArguments constructs the command line arguments for go build
func (h *GoBuild) buildArguments(tempFileName string) []string {
	buildArgs := []string{"build"}
	ldFlags := []string{}

	if h.config.CompilingArguments != nil {
		args := h.config.CompilingArguments()
		for i := 0; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "-X") {
				if arg == "-X" && i+1 < len(args) {
					// -X followed by separate argument
					ldFlags = append(ldFlags, arg, args[i+1])
					i++ // Skip next argument as it's part of -X
				} else if strings.Contains(arg, "=") {
					// -X key=value in single argument
					ldFlags = append(ldFlags, arg)
				} else {
					// Just -X without value, add to ldFlags
					ldFlags = append(ldFlags, arg)
				}
			} else {
				buildArgs = append(buildArgs, arg)
			}
		}
	}

	// Add ldflags if any were found
	if len(ldFlags) > 0 {
		buildArgs = append(buildArgs, "-ldflags="+strings.Join(ldFlags, " "))
	}

	buildArgs = append(buildArgs, "-o", path.Join(h.config.OutFolder, tempFileName), h.config.MainFilePath)
	return buildArgs
}
