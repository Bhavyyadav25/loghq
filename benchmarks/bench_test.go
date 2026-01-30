package benchmarks

import (
	"io"
	"testing"

	"github.com/Bhavyyadav25/loghq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// --- Helpers ---

type discardWriteSyncer struct{}

func (discardWriteSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (discardWriteSyncer) Sync() error                 { return nil }

func newLoghq() *loghq.Logger {
	return loghq.New(
		loghq.WithHandler(loghq.NewJSONHandler(discardWriteSyncer{})),
		loghq.WithLevel(loghq.InfoLevel),
		loghq.WithCaller(false),
		loghq.WithStackLevel(loghq.FatalLevel+1),
	)
}

func newZap() *zap.Logger {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	return zap.New(core)
}

// --- loghq ---

func BenchmarkLoghqDisabled(b *testing.B) {
	l := newLoghq()
	l.SetLevel(loghq.ErrorLevel)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("this is disabled")
	}
}

func BenchmarkLoghqInfoNoFields(b *testing.B) {
	l := newLoghq()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkLoghqInfo5FieldsKV(b *testing.B) {
	l := newLoghq()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("request",
			"method", "GET",
			"path", "/api/users",
			"status", 200,
			"bytes", 1024,
			"elapsed", "12ms",
		)
	}
}

func BenchmarkLoghqInfo10FieldsKV(b *testing.B) {
	l := newLoghq()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("request",
			"method", "GET",
			"path", "/api/users",
			"status", 200,
			"bytes", 1024,
			"elapsed", "12ms",
			"user", "ali",
			"ip", "192.168.1.1",
			"ua", "Mozilla/5.0",
			"ref", "https://example.com",
			"rid", "abc-123",
		)
	}
}

func BenchmarkLoghqWithFields(b *testing.B) {
	l := newLoghq()
	child := l.WithFields(loghq.Fields{"service": "api", "version": "1.0"})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		child.Info("request", "status", 200)
	}
}

func BenchmarkLoghqParallel(b *testing.B) {
	l := newLoghq()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", "id", 42)
		}
	})
}

// --- zap ---

func BenchmarkZapDisabled(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Debug("this is disabled")
	}
}

func BenchmarkZapInfoNoFields(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkZapInfo5Fields(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("request",
			zap.String("method", "GET"),
			zap.String("path", "/api/users"),
			zap.Int("status", 200),
			zap.Int("bytes", 1024),
			zap.String("elapsed", "12ms"),
		)
	}
}

func BenchmarkZapInfo10Fields(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("request",
			zap.String("method", "GET"),
			zap.String("path", "/api/users"),
			zap.Int("status", 200),
			zap.Int("bytes", 1024),
			zap.String("elapsed", "12ms"),
			zap.String("user", "ali"),
			zap.String("ip", "192.168.1.1"),
			zap.String("ua", "Mozilla/5.0"),
			zap.String("ref", "https://example.com"),
			zap.String("rid", "abc-123"),
		)
	}
}

func BenchmarkZapWithFields(b *testing.B) {
	l := newZap().With(zap.String("service", "api"), zap.String("version", "1.0"))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("request", zap.Int("status", 200))
	}
}

func BenchmarkZapParallel(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", zap.Int("id", 42))
		}
	})
}
