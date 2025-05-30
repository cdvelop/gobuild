# gobuild

Thread-safe Go/WASM build handler with sync/async compilation support.

## Installation

```bash
go get github.com/cdvelop/gobuild
```

## Quick Start

```go
config := &gobuild.Config{
    Command:      "go",           // or "tinygo"
    MainFilePath: "main.go",
    OutName:      "app",
    Extension:    ".exe",         // ".wasm" for WASM, "" for Unix
    OutFolder:    "dist",
    Log:          os.Stdout,
    Timeout:      5 * time.Second,
}

compiler := gobuild.New(config)
err := compiler.CompileProgram() // Synchronous
```

## Async Compilation

```go
config.Callback = func(err error) {
    if err != nil {
        log.Printf("Failed: %v", err)
    } else {
        log.Printf("Success!")
    }
}
err := compiler.CompileProgram() // Returns immediately
```

## Thread-Safe Control

```go
// Cancel ongoing compilation
compiler.Cancel()

// Check compilation status
if compiler.IsCompiling() {
    fmt.Println("Compilation in progress...")
}
```

## Configuration

```go
type Config struct {
    Command             string          // "go" or "tinygo"
    MainFilePath        string          // Path to main.go
    OutName             string          // Output name (without extension)
    Extension           string          // ".exe", ".wasm", ""
    OutFolder           string          // Output directory
    Log                 io.Writer       // Output writer (optional)
    CompilingArguments  func() []string // Build arguments (optional)
    Callback            func(error)     // Async callback (optional)
    Timeout             time.Duration   // Default: 5s
}
```

## Methods

- `CompileProgram() error` - Compile (sync/async based on callback)
- `Cancel() error` - Cancel current compilation
- `IsCompiling() bool` - Check if compilation is active

## Features

- **Thread-safe**: Automatic cancellation of previous compilations
- **Unique temp files**: Prevents conflicts during concurrent builds
- **Context-aware**: Proper cancellation and timeout handling
