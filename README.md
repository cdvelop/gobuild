# gobuild

Minimal Go/WASM build handler with sync/async compilation support.

## Installation

```bash
go get github.com/cdvelop/gobuild
```

## Usage

```go
package main

import (
    "os"
    "github.com/cdvelop/gobuild"
)

func main() {
    config := &gobuild.Config{
        Command:      "go",           // or "tinygo"
        MainFilePath: "main.go",
        OutName:      "app",
        Extension:    ".exe",         // ".wasm" for WASM, "" for Unix
        OutFolder:    "dist",
        Log:          os.Stdout,
        // Optional: Timeout, Callback for async, CompilingArguments
    }

    compiler := gobuild.New(config)
    
    if err := compiler.CompileProgram(); err != nil {
        panic(err)
    }
}
```

### Async compilation
```go
config.Callback = func(err error) {
    if err != nil {
        log.Printf("Failed: %v", err)
    }
}
```

### Custom build args
```go
config.CompilingArguments = func() []string {
    return []string{"-race", "-ldflags", "-s -w"}
}
```

## API Reference

### Config

```go
type Config struct {
    Command             string          // "go" or "tinygo"
    MainFilePath        string          // Path to main.go
    OutName             string          // Output name (without extension)
    Extension           string          // ".exe", ".wasm", ""
    OutFolder           string          // Output directory
    Log                 io.Writer       // Output writer
    CompilingArguments  func() []string // Build arguments
    Callback            func(error)     // Async callback (optional)
    Timeout             time.Duration   // Default: 5s
}
```

### Methods

- `gobuild.New(config *Config) *GoBuild` - Create compiler
- `compiler.CompileProgram() error` - Compile (sync if no callback, async if callback set)
- `compiler.UnobservedFiles() []string` - Get temp files list

## License

MIT
