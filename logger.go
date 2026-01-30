package loghq

import (
	"context"
	"os"
	"sync/atomic"
	"time"
)

// Logger is the core logging engine. It is safe for concurrent use.
type Logger struct {
	level      atomic.Int32
	handler    Handler
	addCaller  bool
	stackLevel Level
	callerSkip int
	fields     []Field
	ctx        context.Context
}

// New creates a new Logger with the given options.
// By default: InfoLevel, discard handler, caller capture enabled,
// stack traces on ErrorLevel and above.
func New(opts ...Option) *Logger {
	l := &Logger{
		handler:    discardHandler{},
		addCaller:  true,
		stackLevel: ErrorLevel,
	}
	l.level.Store(int32(InfoLevel))

	for _, opt := range opts {
		opt(l)
	}
	return l
}

// clone creates a shallow copy with independent fields slice.
func (l *Logger) clone() *Logger {
	c := &Logger{
		handler:    l.handler,
		addCaller:  l.addCaller,
		stackLevel: l.stackLevel,
		callerSkip: l.callerSkip,
		ctx:        l.ctx,
	}
	c.level.Store(l.level.Load())

	if len(l.fields) > 0 {
		c.fields = make([]Field, len(l.fields))
		copy(c.fields, l.fields)
	}
	return c
}

// WithFields returns a new Logger with the given fields pre-bound.
func (l *Logger) WithFields(f Fields) *Logger {
	c := l.clone()
	c.fields = append(c.fields, fieldsFromMap(f)...)
	return c
}

// With returns a new Logger with typed fields pre-bound.
func (l *Logger) With(fields ...Field) *Logger {
	c := l.clone()
	c.fields = append(c.fields, fields...)
	return c
}

// WithContext returns a new Logger that extracts fields from the context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	c := l.clone()
	c.ctx = ctx
	return c
}

// SetLevel changes the logger's level atomically.
func (l *Logger) SetLevel(lvl Level) {
	l.level.Store(int32(lvl))
}

// log is the core hot path. Everything funnels through here.
func (l *Logger) log(lvl Level, msg string, kvs []interface{}) {
	// Lock-free level check — costs ~1ns when disabled.
	if lvl < Level(l.level.Load()) {
		return
	}

	rec := acquireRecord()
	rec.Time = time.Now()
	rec.Level = lvl
	rec.Message = msg

	// Pre-bound fields
	rec.AddFields(l.fields)

	// Context fields
	if l.ctx != nil {
		rec.AddFields(fieldsFromContext(l.ctx))
	}

	// Parse slog-style key-value pairs
	if len(kvs) > 0 {
		rec.AddFields(parseKVPairs(kvs))
	}

	// Caller capture (skip 3 frames: log -> Trace/Info/etc -> user code)
	if l.addCaller {
		rec.Caller = captureCaller(3 + l.callerSkip)
	}

	// Stack trace for error+ levels
	if lvl >= l.stackLevel {
		rec.Stack = captureStack(3 + l.callerSkip)
	}

	// Handler errors are intentionally discarded on the hot path.
	// Use handler-level error callbacks for production error monitoring.
	_ = l.handler.Handle(rec)
	releaseRecord(rec)

	if lvl == FatalLevel {
		os.Exit(1)
	}
}

// --- Level methods ---

func (l *Logger) Trace(msg string, kvs ...interface{})   { l.log(TraceLevel, msg, kvs) }
func (l *Logger) Debug(msg string, kvs ...interface{})   { l.log(DebugLevel, msg, kvs) }
func (l *Logger) Info(msg string, kvs ...interface{})    { l.log(InfoLevel, msg, kvs) }
func (l *Logger) Success(msg string, kvs ...interface{}) { l.log(SuccessLevel, msg, kvs) }
func (l *Logger) Warn(msg string, kvs ...interface{})    { l.log(WarnLevel, msg, kvs) }
func (l *Logger) Error(msg string, kvs ...interface{})   { l.log(ErrorLevel, msg, kvs) }
func (l *Logger) Fatal(msg string, kvs ...interface{})   { l.log(FatalLevel, msg, kvs) }

// Flush flushes the handler if it implements Flusher.
func (l *Logger) Flush() error {
	if f, ok := l.handler.(Flusher); ok {
		return f.Flush()
	}
	return nil
}

// Close closes the handler if it implements Closer.
func (l *Logger) Close() error {
	if c, ok := l.handler.(Closer); ok {
		return c.Close()
	}
	return nil
}
