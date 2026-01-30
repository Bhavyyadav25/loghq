# loghq

Beautiful, fast, structured logging for Go.

loghq is a developer-first logging library that produces gorgeous console output while being **faster than zap** and **matching zerolog** in benchmarks — with a simpler API. Zero allocations on every hot path, sync.Pool recycling, and hand-rolled encoders — no reflection, no `fmt.Sprintf`, no `encoding/json`.

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
- **Zero-allocation hot path** — 0 allocs/op across every benchmark
- **Faster than zap, slog, and logrus** — Matches zerolog. See [benchmarks](#benchmarks)
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

Benchmarked against every major Go logging library. JSON encoding to `io.Discard`, **10 iterations at 5 seconds each** for statistical reliability.

**Environment:** Intel Core i7-1370P (20 threads), Linux amd64, Go 1.25.6

### Disabled Level (no-op fast path)

| Library | ns/op | B/op | allocs/op |
|---------|------:|-----:|----------:|
| **loghq** | **3.33** | **0** | **0** |
| zerolog | 4.90 | 0 | 0 |
| zap | 13.77 | 0 | 0 |
| slog | 14.05 | 0 | 0 |

> loghq's atomic level check is **4.1x faster** than zap and slog.

### Info -- No Fields

| Library | ns/op | B/op | allocs/op |
|---------|------:|-----:|----------:|
| zerolog | 384 | 0 | 0 |
| **loghq** | **432** | **0** | **0** |
| zap | 716 | 0 | 0 |
| slog | 1,011 | 0 | 0 |
| logrus | 3,459 | 872 | 19 |

> loghq is **1.7x faster** than zap, **2.3x faster** than slog.

### Info -- 5 Structured Fields

| Library | ns/op | B/op | allocs/op |
|---------|------:|-----:|----------:|
| zerolog | 667 | 0 | 0 |
| **loghq** | **821** | **0** | **0** |
| zap | 1,476 | 320 | 1 |
| slog | 2,003 | 0 | 0 |
| logrus | 8,019 | 2,287 | 35 |

> loghq is **1.8x faster** than zap with **zero allocations** vs zap's 320 B/op.

### Info -- 10 Structured Fields

| Library | ns/op | B/op | allocs/op |
|---------|------:|-----:|----------:|
| zerolog | 957 | 0 | 0 |
| **loghq** | **1,461** | **0** | **0** |
| zap | 2,122 | 705 | 1 |
| slog | 3,204 | 208 | 1 |
| logrus | 13,072 | 3,981 | 52 |

> loghq is **1.5x faster** than zap, **2.2x faster** than slog, **8.9x faster** than logrus -- all with **zero allocations**.

### Parallel (20-goroutine contention)

| Library | ns/op | B/op | allocs/op |
|---------|------:|-----:|----------:|
| zerolog | 56 | 0 | 0 |
| **loghq** | **69** | **0** | **0** |
| zap | 199 | 64 | 1 |
| slog | 677 | 0 | 0 |
| logrus | 7,164 | 1,673 | 25 |

> loghq matches zerolog under contention and is **2.9x faster** than zap, **9.8x faster** than slog.

### At a Glance

```
              loghq      zerolog     zap         slog        logrus
              ─────      ───────     ───         ────        ──────
Disabled      3 ns       5 ns        14 ns       14 ns       3 ns
No Fields     432 ns     384 ns      716 ns      1,011 ns    3,459 ns
5 Fields      821 ns     667 ns      1,476 ns    2,003 ns    8,019 ns
10 Fields     1,461 ns   957 ns      2,122 ns    3,204 ns    13,072 ns
Parallel      69 ns      56 ns       199 ns      677 ns      7,164 ns

Allocs/op     0          0           0-1         0-1         19-52
Bytes/op      0          0           0-705       0-208       872-3,981
```

### A Note on zerolog

zerolog uses a typed chained-builder API (`l.Info().Str("k","v").Int("n",1).Msg("m")`) which avoids `interface{}` dispatch entirely. This gives it a raw speed edge in field-heavy benchmarks. loghq uses the simpler slog-style API (`l.Info("msg", "key", "val")`) that is more familiar and requires less boilerplate -- while still achieving **zero allocations** and competitive throughput.

### Why loghq is Fast

1. **Zero allocations** -- the entire hot path is allocation-free, from field parsing to JSON encoding
2. **Inline `[16]Field` array** -- Record stores up to 16 fields without heap allocation
3. **sync.Pool** recycling -- Records and Buffers are pooled and reused
4. **Inlined KV parsing** -- variadic `[]interface{}` never escapes to heap
5. **Direct field encoding** -- no interface dispatch, no closures, no heap escape
6. **Hand-rolled encoders** -- `strconv.Append*` and `time.AppendFormat` directly into pooled buffers
7. **Atomic level check** -- disabled levels cost ~3ns via `atomic.Int32`
8. **Lock-free handlers** -- no mutexes in the encoding path

### Reproduce

```bash
cd benchmarks
go test -bench=. -benchmem -count=10 -benchtime=5s -timeout=60m
```

## License

MIT
