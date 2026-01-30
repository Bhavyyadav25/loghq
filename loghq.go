// Package loghq provides beautiful, fast, structured logging for Go.
//
// loghq is a developer-first logging library that produces gorgeous console
// output while maintaining zero-allocation performance on the hot path.
//
// Usage:
//
//	loghq.Info("server started", "port", 8080)
//	loghq.WithFields(loghq.Fields{"service": "api"}).Info("ready")
//	loghq.Error("request failed", "status", 500)
package loghq

import (
	"context"
	"sync/atomic"
)

// defaultLogger is the package-level logger, protected by atomic.Pointer
// for thread-safe reads and writes.
var defaultLogger atomic.Pointer[Logger]

func init() {
	l := New(WithHandler(NewConsoleHandler()))
	defaultLogger.Store(l)
}

// SetDefault replaces the default logger. Safe for concurrent use.
func SetDefault(l *Logger) {
	defaultLogger.Store(l)
}

// Default returns the current default logger.
func Default() *Logger {
	return defaultLogger.Load()
}

// --- Package-level convenience functions ---

func Trace(msg string, kvs ...interface{})   { defaultLogger.Load().log(TraceLevel, msg, kvs) }
func Debug(msg string, kvs ...interface{})   { defaultLogger.Load().log(DebugLevel, msg, kvs) }
func Info(msg string, kvs ...interface{})    { defaultLogger.Load().log(InfoLevel, msg, kvs) }
func Success(msg string, kvs ...interface{}) { defaultLogger.Load().log(SuccessLevel, msg, kvs) }
func Warn(msg string, kvs ...interface{})    { defaultLogger.Load().log(WarnLevel, msg, kvs) }
func Error(msg string, kvs ...interface{})   { defaultLogger.Load().log(ErrorLevel, msg, kvs) }
func Fatal(msg string, kvs ...interface{})   { defaultLogger.Load().log(FatalLevel, msg, kvs) }

func WithFields(f Fields) *Logger            { return defaultLogger.Load().WithFields(f) }
func WithContext(ctx context.Context) *Logger { return defaultLogger.Load().WithContext(ctx) }
func With(fields ...Field) *Logger           { return defaultLogger.Load().With(fields...) }

func Flush() error { return defaultLogger.Load().Flush() }
func Close() error { return defaultLogger.Load().Close() }
