package gobuild

import (
	"io"
	"time"
)

// CompileCallback is called when compilation completes (success or failure)
type CompileCallback func(error)

// Config holds the configuration for Go compilation
type Config struct {
	Command            string          // eg: "go", "tinygo"
	MainFilePath       string          // eg: web/main.server.go, web/main.wasm.go
	OutName            string          // eg: app, user, main.server
	Extension          string          // eg: .exe, .wasm
	CompilingArguments func() []string // eg: []string{"-X 'main.version=v1.0.0'"}
	OutFolder          string          // eg: web, web/public/wasm
	Logger             io.Writer       // output for log messages to integrate with other tools (e.g., TUI)
	Callback           CompileCallback // optional callback for async compilation
	Timeout            time.Duration   // max compilation time, defaults to 5 seconds if not set
	Env                []string        // environment variables, eg: []string{"GOOS=js", "GOARCH=wasm"}
}
