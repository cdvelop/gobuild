# gobuild

Minimal build handler for compiling Go or WebAssembly targets via CLI. Easy to integrate into other toolkits or workflows.

## Features

- **Synchronous and Asynchronous compilation**: Support for both blocking and non-blocking compilation modes
- **Context-based timeouts**: Built-in timeout management with cancellation support
- **Flexible logging**: Configurable output writer for integration with TUI and other tools
- **Cross-platform support**: Works with `go` and `tinygo` compilers
- **Clean file management**: Automatic handling of temporary files and output cleanup

## Requirements

- Go 1.20+ (for TinyGo compatibility)
- Go or TinyGo compiler installed

## Installation

```bash
go get github.com/cdvelop/gobuild
```

## Quick Start

### Basic Synchronous Compilation

```go
package main

import (
    "log"
    "os"
    "github.com/cdvelop/gobuild"
)

func main() {
    config := &gobuild.Config{
        Command:      "go",
        MainFilePath: "cmd/main.go",
        OutName:      "myapp",
        Extension:    ".exe", // Use "" for Unix/WASM
        OutFolder:    "dist",
        Log:          os.Stdout,
    }

    compiler := gobuild.New(config)
    
    if err := compiler.CompileProgram(); err != nil {
        log.Fatal("Compilation failed:", err)
    }
    
    log.Println("Compilation successful!")
}
```

### Asynchronous Compilation with Callback

```go
package main

import (
    "log"
    "os"
    "time"
    "github.com/cdvelop/gobuild"
)

func main() {
    config := &gobuild.Config{
        Command:      "go",
        MainFilePath: "cmd/main.go",
        OutName:      "myapp",
        Extension:    "",
        OutFolder:    "dist",
        Log:          os.Stdout,
        Timeout:      30 * time.Second, // Optional: defaults to 5 seconds
        Callback: func(err error) {
            if err != nil {
                log.Printf("Async compilation failed: %v", err)
            } else {
                log.Println("Async compilation successful!")
            }
        },
    }

    compiler := gobuild.New(config)
    
    // This returns immediately
    if err := compiler.CompileProgram(); err != nil {
        log.Fatal("Failed to start compilation:", err)
    }
    
    // Do other work while compilation runs in background
    log.Println("Compilation started in background...")
    
    // Wait for completion (in real applications, you'd handle this differently)
    time.Sleep(10 * time.Second)
}
```

### WebAssembly Compilation with TinyGo

```go
package main

import (
    "log"
    "os"
    "github.com/cdvelop/gobuild"
)

func main() {
    config := &gobuild.Config{
        Command:      "tinygo",
        MainFilePath: "web/main.wasm.go",
        OutName:      "app",
        Extension:    ".wasm",
        OutFolder:    "web/public/wasm",
        Log:          os.Stdout,
        CompilingArguments: func() []string {
            return []string{
                "-target", "wasm",
                "-no-debug",
                "-opt", "2",
            }
        },
    }

    compiler := gobuild.New(config)
    
    if err := compiler.CompileProgram(); err != nil {
        log.Fatal("WASM compilation failed:", err)
    }
    
    log.Println("WASM compilation successful!")
}
```

### Advanced Usage with Build Flags

```go
package main

import (
    "log"
    "os"
    "github.com/cdvelop/gobuild"
)

func main() {
    config := &gobuild.Config{
        Command:      "go",
        MainFilePath: "cmd/server/main.go",
        OutName:      "server",
        Extension:    "",
        OutFolder:    "bin",
        Log:          os.Stdout,
        CompilingArguments: func() []string {
            return []string{
                "-race",                           // Enable race detector
                "-X", "main.version=v1.0.0",      // Set version via ldflags
                "-X", "main.buildTime=" + buildTime, // Set build time
                "-tags", "production",             // Build tags
                "-trimpath",                       // Remove file system paths
            }
        },
    }

    compiler := gobuild.New(config)
    
    if err := compiler.CompileProgram(); err != nil {
        log.Fatal("Compilation failed:", err)
    }
    
    log.Println("Production build successful!")
}
```

## Configuration Reference

### Config Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `Command` | `string` | Compiler command to use | `"go"`, `"tinygo"` |
| `MainFilePath` | `string` | Path to main Go file | `"cmd/main.go"`, `"web/main.wasm.go"` |
| `OutName` | `string` | Output binary name (without extension) | `"app"`, `"server"`, `"main"` |
| `Extension` | `string` | Output file extension | `".exe"` (Windows), `""` (Unix), `".wasm"` |
| `CompilingArguments` | `func() []string` | Function returning compiler arguments | See examples below |
| `OutFolder` | `string` | Output directory | `"dist"`, `"bin"`, `"web/public"` |
| `Log` | `io.Writer` | Writer for compilation output | `os.Stdout`, `log.Writer()`, custom |
| `Callback` | `CompileCallback` | Optional async completion callback | `func(err error) { ... }` |
| `Timeout` | `time.Duration` | Max compilation time (default: 5s) | `30 * time.Second` |

### CompileCallback

```go
type CompileCallback func(error)
```

Called when asynchronous compilation completes. Receives `nil` on success or an error on failure.

### Compilation Behavior

- **Synchronous mode**: When `Callback` is `nil`, `CompileProgram()` blocks until completion
- **Asynchronous mode**: When `Callback` is set, `CompileProgram()` returns immediately and calls the callback when done
- **Timeout handling**: All compilations respect the configured timeout (default: 5 seconds)
- **Context cancellation**: Compilations can be cancelled via context timeout

## File Management

The compiler automatically manages temporary files during compilation:

```go
// Get list of files that should be ignored by file watchers
files := compiler.UnobservedFiles()
// Returns: ["app.exe", "app_temp.exe"] (or similar based on config)
```

## Error Handling

All errors are wrapped with context information:

```go
if err := compiler.CompileProgram(); err != nil {
    // err contains the original error plus context
    log.Printf("Compilation failed: %v", err)
}
```

## Integration Examples

### With File Watcher

```go
func setupWatcher(compiler *gobuild.GoBuild) {
    // Exclude build artifacts from watching
    excludeFiles := compiler.UnobservedFiles()
    
    // Setup your file watcher to ignore these files
    watcher.Exclude(excludeFiles...)
}
```

### With TUI Framework

```go
func buildWithProgress(app *tview.Application) {
    var logOutput strings.Builder
    
    config := &gobuild.Config{
        Command:      "go",
        MainFilePath: "main.go",
        OutName:      "app",
        Extension:    "",
        OutFolder:    "dist",
        Log:          &logOutput,
        Callback: func(err error) {
            app.QueueUpdateDraw(func() {
                if err != nil {
                    showError(err.Error())
                } else {
                    showSuccess("Build completed!")
                }
            })
        },
    }
    
    compiler := gobuild.New(config)
    compiler.CompileProgram()
}
```

## Testing

Run the test suite:

```bash
go test ./...
```

Tests cover:
- Configuration validation
- Synchronous and asynchronous compilation
- Argument parsing and ldflags handling
- File management operations
- Error conditions

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT

## Compatibility

- **Go**: 1.20+
- **TinyGo**: All supported versions
- **Platforms**: Windows, macOS, Linux
- **Architectures**: amd64, arm64, wasm

## Changelog

### v2.0.0 (Upcoming)
- Added asynchronous compilation support
- Added context-based timeout management
- Reorganized codebase into logical modules
- Renamed `Writer` to `Log` for clarity
- Renamed `NewGoCompiler` to `New`
- Improved argument parsing for ldflags
- Added comprehensive test coverage
- Updated documentation

### v1.0.0
- Initial release
- Basic synchronous compilation support
