package loghq

import "testing"

// discardWriteSyncer wraps io.Discard as a WriteSyncer for benchmarks.
type discardWriteSyncer struct{}

func (discardWriteSyncer) Write(p []byte) (int, error) { return len(p), nil }
func (discardWriteSyncer) Sync() error                 { return nil }

func newBenchLogger() *Logger {
	return New(
		WithHandler(NewJSONHandler(discardWriteSyncer{})),
		WithLevel(InfoLevel),
		WithCaller(false),
		WithStackLevel(FatalLevel+1),
	)
}

func BenchmarkDisabled(b *testing.B) {
	l := newBenchLogger()
	l.SetLevel(ErrorLevel)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("this is disabled")
	}
}

func BenchmarkInfoNoFields(b *testing.B) {
	l := newBenchLogger()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

func BenchmarkInfo5FieldsKV(b *testing.B) {
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

func BenchmarkInfo10FieldsKV(b *testing.B) {
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

func BenchmarkWithFields(b *testing.B) {
	l := newBenchLogger()
	child := l.WithFields(Fields{"service": "api", "version": "1.0"})
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		child.Info("request", "status", 200)
	}
}

func BenchmarkParallel(b *testing.B) {
	l := newBenchLogger()
	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel", "id", 42)
		}
	})
}
