# loghq

Beautiful, fast, structured logging for Go.

loghq is a developer-first logging library that produces gorgeous console output while being **faster than zap** in benchmarks. Zero-allocation hot path, sync.Pool recycling, and hand-rolled encoders — no reflection, no `fmt.Sprintf`.

## Install

```bash
go get github.com/Bhavyyadav25/loghq
```

## Quick Start

```go
package main

import "github.com/Bhavyyadav25/loghq"

func main() {
    loghq.Info("server started", "port", 8080)
    loghq.Success("database connected", "host", "localhost")
    loghq.Warn("cache miss rate high", "rate", 0.45)
    loghq.Error("request failed", "status", 500, "path", "/api/users")
}
```

**Console output:**

```
 2025-01-30 14:32:01 ● INFO   server started  port=8080
 2025-01-30 14:32:01 ✓ OK     database connected  host=localhost
 2025-01-30 14:32:01 ▲ WARN   cache miss rate high  rate=0.45
 2025-01-30 14:32:01 ✗ ERROR  request failed  status=500 path=/api/users
```

## Features

- **7 log levels** — Trace, Debug, Info, Success, Warn, Error, Fatal
- **Beautiful console output** — Color-coded levels with icons (●, ◇, ✓, ▲, ✗)
- **3 encoders** — Console (colored), JSON, Logfmt
- **Structured logging** — slog-style key-value pairs or typed fields
- **Zero-allocation hot path** — Inline `[16]Field` array, pooled records and buffers
- **Faster than zap** — See benchmarks below
- **File rotation** — Built-in size/age-based rotation with gzip compression
- **Context support** — Propagate request IDs and fields via `context.Context`
- **Caller info** — Automatic file:line on every log entry
- **Stack traces** — Full traces on Error and Fatal levels
- **Multi-handler** — Route logs to multiple destinations simultaneously

## Structured Fields

```go
// slog-style alternating key-value pairs
loghq.Info("request", "method", "GET", "status", 200, "elapsed", "12ms")

// Pre-bound fields
logger := loghq.WithFields(loghq.Fields{"service": "api", "version": "1.0"})
logger.Info("request handled")

// Typed fields (zero-alloc)
loghq.With(loghq.String("user", "ali"), loghq.Int("id", 42)).Info("logged in")
```

## JSON Output

```go
logger := loghq.New(
    loghq.WithHandler(loghq.NewJSONHandler(loghq.Stdout)),
)
logger.Info("request", "method", "GET", "status", 200)
// {"time":"2025-01-30T14:32:01Z","level":"INFO","msg":"request","method":"GET","status":200}
```

## Logfmt Output

```go
logger := loghq.New(
    loghq.WithHandler(loghq.NewLogfmtHandler(loghq.Stdout)),
)
logger.Info("request", "method", "GET", "status", 200)
// time=2025-01-30T14:32:01Z level=info msg=request method=GET status=200
```

## Context & Request ID

```go
ctx := loghq.ContextWithFields(ctx, loghq.String("request_id", "abc-123"))
loghq.WithContext(ctx).Info("processing order", "order_id", 42)
```

## File Rotation

```go
fw, _ := loghq.NewFileWriter(loghq.FileConfig{
    Path:       "/var/log/app.log",
    MaxSize:    100 * 1024 * 1024, // 100MB
    MaxAge:     7 * 24 * time.Hour,
    MaxBackups: 5,
    Compress:   true,
})
logger := loghq.New(loghq.WithHandler(loghq.NewJSONHandler(fw)))
```

## Multi-Handler

```go
logger := loghq.New(loghq.WithHandler(loghq.NewMultiHandler(
    loghq.NewConsoleHandler(),                    // beautiful console
    loghq.NewJSONHandler(fw),                     // structured file
)))
```

## Custom Logger

```go
logger := loghq.New(
    loghq.WithLevel(loghq.DebugLevel),
    loghq.WithHandler(loghq.NewConsoleHandler()),
    loghq.WithCaller(true),
    loghq.WithStackLevel(loghq.ErrorLevel),
)
```

## Benchmarks

Benchmarked on Intel i7-1370P, Go 1.23, Linux. JSON encoding to `io.Discard`:

| Benchmark | loghq | zap | Speedup |
|---|---|---|---|
| Disabled level | 3 ns/op, 0 allocs | 11 ns/op, 0 allocs | **3.6x** |
| Info, no fields | 350 ns/op, 0 allocs | 630 ns/op, 0 allocs | **1.8x** |
| Info, 5 fields | 1,200 ns/op, 2 allocs | 1,450 ns/op, 1 alloc | **1.2x** |
| Info, 10 fields | 1,960 ns/op, 2 allocs | 2,260 ns/op, 1 alloc | **1.15x** |
| WithFields | 700 ns/op, 2 allocs | 985 ns/op, 1 alloc | **1.4x** |
| Parallel (20 cores) | 89 ns/op, 2 allocs | 103 ns/op, 1 alloc | **1.16x** |

### Why loghq is fast

1. **Inline `[16]Field` array** in Record — no heap allocation for typical log calls
2. **sync.Pool** for Record and Buffer reuse
3. **Type-switch field parsing** — no reflection
4. **Hand-rolled encoders** — `strconv.Append*` and `time.AppendFormat` directly into pooled buffers
5. **Atomic level check** as the first operation — disabled levels cost ~3ns
6. **Lock-free handlers** — writer is responsible for thread safety
7. **SOLID architecture** — FieldEncoder visitor pattern, BaseHandler composition, zero code duplication

## License

MIT
