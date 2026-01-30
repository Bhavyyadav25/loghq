package loghq

import (
	"io"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// --- Helpers ---

// discardWriteSyncer wraps io.Discard as a WriteSyncer for loghq benchmarks.
type discardWriteSyncer struct{}

func (discardWriteSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (discardWriteSyncer) Sync() error                 { return nil }

func newBenchLogger() *Logger {
	h := NewJSONHandler(discardWriteSyncer{})
	return New(
		WithHandler(h),
		WithLevel(InfoLevel),
		WithCaller(false),
		WithStackLevel(FatalLevel+1),
	)
}

func newBenchZap() *zap.Logger {
	enc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, zapcore.AddSync(io.Discard), zap.InfoLevel)
	return zap.New(core)
}

// --- loghq benchmarks ---

func BenchmarkLoghqDisabled(b *testing.B) {
	l := newBenchLogger()
	l.SetLevel(ErrorLevel)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("this is disabled")
	}
}

func BenchmarkLoghqInfoNoFields(b *testing.B) {
	l := newBenchLogger()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkLoghqInfo5FieldsKV(b *testing.B) {
	l := newBenchLogger()
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
	l := newBenchLogger()
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
	l := newBenchLogger()
	child := l.WithFields(Fields{"service": "api", "version": "1.0"})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		child.Info("request", "status", 200)
	}
}

func BenchmarkLoghqParallel(b *testing.B) {
	l := newBenchLogger()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", "id", 42)
		}
	})
}

// --- zap benchmarks ---

func BenchmarkZapDisabled(b *testing.B) {
	l := newBenchZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Debug("this is disabled") // Info logger, Debug is disabled
	}
}

func BenchmarkZapInfoNoFields(b *testing.B) {
	l := newBenchZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkZapInfo5Fields(b *testing.B) {
	l := newBenchZap()
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
	l := newBenchZap()
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
	l := newBenchZap().With(zap.String("service", "api"), zap.String("version", "1.0"))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("request", zap.Int("status", 200))
	}
}

func BenchmarkZapParallel(b *testing.B) {
	l := newBenchZap()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", zap.Int("id", 42))
		}
	})
}
