package gobuild

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"path"
	"strings"
)

// compileSync performs the actual compilation synchronously with context timeout
func (h *GoBuild) compileSync() error {
	var this = errors.New("CompileProgram")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), h.Timeout)
	defer cancel()

	buildArgs := h.buildArguments()
	h.Cmd = exec.CommandContext(ctx, h.Command, buildArgs...)

	stderr, err := h.Cmd.StderrPipe()
	if err != nil {
		return errors.Join(this, err)
	}

	stdout, err := h.Cmd.StdoutPipe()
	if err != nil {
		return errors.Join(this, err)
	}

	err = h.Cmd.Start()
	if err != nil {
		return errors.Join(this, err)
	}

	if h.Log != nil {
		go io.Copy(h.Log, stderr)
		go io.Copy(h.Log, stdout)
	}

	if err := h.Cmd.Wait(); err != nil {
		// Clean up temporary file if compilation failed
		h.cleanupTempFile()
		return errors.Join(this, err)
	}

	return h.renameOutputFile()
}

// buildArguments constructs the command line arguments for go build
func (h *GoBuild) buildArguments() []string {
	buildArgs := []string{"build"}
	ldFlags := []string{}

	if h.CompilingArguments != nil {
		args := h.CompilingArguments()
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

	buildArgs = append(buildArgs, "-o", path.Join(h.OutFolder, h.outTempFileName), h.MainFilePath)
	return buildArgs
}
