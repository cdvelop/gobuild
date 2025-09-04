package gobuild

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

// compileSync performs the actual compilation synchronously with context timeout
func (h *GoBuild) compileSync(ctx context.Context, comp *compilation) error {
	var this = errors.New("compileSync")

	buildArgs := h.buildArguments(comp.tempFile)

	comp.cmd = exec.CommandContext(ctx, h.config.Command, buildArgs...)

	// Set working directory to output folder for relative paths
	comp.cmd.Dir = h.config.OutFolderRelativePath

	// Set environment variables if provided
	if len(h.config.Env) > 0 {
		comp.cmd.Env = append(os.Environ(), h.config.Env...)
	}

	// Use CombinedOutput for simpler and more reliable error capture
	output, err := comp.cmd.CombinedOutput()

	if err != nil {
		// Emit a single log entry containing the error and the raw build output (no processing)
		if h.config.Logger != nil {
			if len(output) > 0 {
				h.config.Logger(this, "build failed:", err, "\n"+string(output)+"\n")
			} else {
				h.config.Logger(this, "build failed:", err)
			}
		}
		// Clean up temporary file if compilation failed
		h.cleanupTempFile(comp.tempFile)

		// Return an error that contains both the original error and the raw build output
		return errors.Join(this, fmt.Errorf("%v: %s", err, strings.TrimSpace(string(output))))
	}

	// fmt.Fprintf(h.config.Logger, "Compilation successful, renaming %s\n", comp.tempFile)

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

	buildArgs = append(buildArgs, "-o", path.Join(h.config.OutFolderRelativePath, tempFileName), h.config.MainInputFileRelativePath)
	return buildArgs
}
