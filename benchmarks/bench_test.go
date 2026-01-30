package benchmarks

import (
	"io"
	"log/slog"
	"testing"

	"github.com/Bhavyyadav25/loghq"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ============================================================
// Helpers
// ============================================================

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

func newZerolog() zerolog.Logger {
	return zerolog.New(io.Discard).With().Timestamp().Logger()
}

func newLogrus() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	l.SetFormatter(&logrus.JSONFormatter{})
	l.SetLevel(logrus.InfoLevel)
	l.SetReportCaller(false)
	return l
}

func newSlog() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// ============================================================
// Disabled level (should be near-zero cost)
// ============================================================

func BenchmarkDisabled_Loghq(b *testing.B) {
	l := newLoghq()
	l.SetLevel(loghq.ErrorLevel)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("this is disabled")
	}
}

func BenchmarkDisabled_Zap(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Debug("this is disabled")
	}
}

func BenchmarkDisabled_Zerolog(b *testing.B) {
	l := newZerolog().Level(zerolog.WarnLevel)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info().Msg("this is disabled")
	}
}

func BenchmarkDisabled_Slog(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("this is disabled")
	}
}

func BenchmarkDisabled_Logrus(b *testing.B) {
	l := newLogrus()
	l.SetLevel(logrus.WarnLevel)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("this is disabled")
	}
}

// ============================================================
// Info with no fields
// ============================================================

func BenchmarkInfoNoFields_Loghq(b *testing.B) {
	l := newLoghq()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkInfoNoFields_Zap(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkInfoNoFields_Zerolog(b *testing.B) {
	l := newZerolog()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info().Msg("hello world")
	}
}

func BenchmarkInfoNoFields_Slog(b *testing.B) {
	l := newSlog()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkInfoNoFields_Logrus(b *testing.B) {
	l := newLogrus()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

// ============================================================
// Info with 5 fields
// ============================================================

func BenchmarkInfo5Fields_Loghq(b *testing.B) {
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

func BenchmarkInfo5Fields_Zap(b *testing.B) {
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

func BenchmarkInfo5Fields_Zerolog(b *testing.B) {
	l := newZerolog()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info().
			Str("method", "GET").
			Str("path", "/api/users").
			Int("status", 200).
			Int("bytes", 1024).
			Str("elapsed", "12ms").
			Msg("request")
	}
}

func BenchmarkInfo5Fields_Slog(b *testing.B) {
	l := newSlog()
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

func BenchmarkInfo5Fields_Logrus(b *testing.B) {
	l := newLogrus()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{
			"method":  "GET",
			"path":    "/api/users",
			"status":  200,
			"bytes":   1024,
			"elapsed": "12ms",
		}).Info("request")
	}
}

// ============================================================
// Info with 10 fields
// ============================================================

func BenchmarkInfo10Fields_Loghq(b *testing.B) {
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

func BenchmarkInfo10Fields_Zap(b *testing.B) {
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

func BenchmarkInfo10Fields_Zerolog(b *testing.B) {
	l := newZerolog()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info().
			Str("method", "GET").
			Str("path", "/api/users").
			Int("status", 200).
			Int("bytes", 1024).
			Str("elapsed", "12ms").
			Str("user", "ali").
			Str("ip", "192.168.1.1").
			Str("ua", "Mozilla/5.0").
			Str("ref", "https://example.com").
			Str("rid", "abc-123").
			Msg("request")
	}
}

func BenchmarkInfo10Fields_Slog(b *testing.B) {
	l := newSlog()
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

func BenchmarkInfo10Fields_Logrus(b *testing.B) {
	l := newLogrus()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.WithFields(logrus.Fields{
			"method":  "GET",
			"path":    "/api/users",
			"status":  200,
			"bytes":   1024,
			"elapsed": "12ms",
			"user":    "ali",
			"ip":      "192.168.1.1",
			"ua":      "Mozilla/5.0",
			"ref":     "https://example.com",
			"rid":     "abc-123",
		}).Info("request")
	}
}

// ============================================================
// Parallel (goroutine contention test)
// ============================================================

func BenchmarkParallel_Loghq(b *testing.B) {
	l := newLoghq()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", "id", 42)
		}
	})
}

func BenchmarkParallel_Zap(b *testing.B) {
	l := newZap()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", zap.Int("id", 42))
		}
	})
}

func BenchmarkParallel_Zerolog(b *testing.B) {
	l := newZerolog()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info().Int("id", 42).Msg("parallel")
		}
	})
}

func BenchmarkParallel_Slog(b *testing.B) {
	l := newSlog()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", "id", 42)
		}
	})
}

func BenchmarkParallel_Logrus(b *testing.B) {
	l := newLogrus()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.WithField("id", 42).Info("parallel")
		}
	})
}
