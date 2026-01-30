package loghq

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

// testWriter captures output for assertions.
type testWriter struct {
	buf bytes.Buffer
}

func (w *testWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *testWriter) Sync() error                 { return nil }
func (w *testWriter) String() string               { return w.buf.String() }
func (w *testWriter) Reset()                       { w.buf.Reset() }

func newTestLogger(w WriteSyncer, handler Handler) *Logger {
	return New(
		WithHandler(handler),
		WithLevel(TraceLevel),
		WithCaller(false),
		WithStackLevel(FatalLevel+1), // disable stack traces
	)
}

// --- Level tests ---

func TestLevelString(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{TraceLevel, "TRACE"},
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{SuccessLevel, "OK"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{FatalLevel, "FATAL"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("Level(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestLevelEnabled(t *testing.T) {
	if !ErrorLevel.Enabled(InfoLevel) {
		t.Error("ErrorLevel should be enabled at InfoLevel threshold")
	}
	if DebugLevel.Enabled(InfoLevel) {
		t.Error("DebugLevel should not be enabled at InfoLevel threshold")
	}
}

func TestParseLevel(t *testing.T) {
	if ParseLevel("error") != ErrorLevel {
		t.Error("ParseLevel(error) failed")
	}
	if ParseLevel("OK") != SuccessLevel {
		t.Error("ParseLevel(OK) failed")
	}
}

// --- Field tests ---

func TestTypedFields(t *testing.T) {
	f := String("name", "ali")
	if f.Type != FieldString || f.Key != "name" || f.Str != "ali" {
		t.Errorf("String field: %+v", f)
	}

	f = Int("count", 42)
	if f.Type != FieldInt64 || f.Ival != 42 {
		t.Errorf("Int field: %+v", f)
	}

	f = Bool("ok", true)
	if f.Type != FieldBool || f.Ival != 1 {
		t.Errorf("Bool field: %+v", f)
	}
}

func TestParseKVPairs(t *testing.T) {
	fields := parseKVPairs([]interface{}{"name", "ali", "age", 25, "active", true})
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
	if fields[0].Key != "name" || fields[0].Str != "ali" {
		t.Errorf("field 0: %+v", fields[0])
	}
	if fields[1].Key != "age" || fields[1].Ival != 25 {
		t.Errorf("field 1: %+v", fields[1])
	}
	if fields[2].Key != "active" || fields[2].Ival != 1 {
		t.Errorf("field 2: %+v", fields[2])
	}
}

func TestParseKVPairsOddArgs(t *testing.T) {
	fields := parseKVPairs([]interface{}{"key"})
	if len(fields) != 1 || fields[0].Str != "MISSING" {
		t.Errorf("odd args: %+v", fields)
	}
}

// --- Buffer tests ---

func TestBuffer(t *testing.T) {
	buf := getBuffer()
	defer putBuffer(buf)

	buf.AppendString("hello ")
	buf.AppendInt(42)
	buf.AppendByte(' ')
	buf.AppendBool(true)

	got := string(buf.Bytes())
	if got != "hello 42 true" {
		t.Errorf("buffer: %q", got)
	}
}

// --- Record tests ---

func TestRecordInlineFields(t *testing.T) {
	rec := acquireRecord()
	defer releaseRecord(rec)

	for i := 0; i < 16; i++ {
		rec.AddField(Int("f", i))
	}
	if rec.NumFields() != 16 {
		t.Errorf("expected 16, got %d", rec.NumFields())
	}
	if len(rec.extra) != 0 {
		t.Error("should not have overflow fields")
	}

	// Add one more to trigger overflow
	rec.AddField(Int("overflow", 99))
	if rec.NumFields() != 17 {
		t.Errorf("expected 17, got %d", rec.NumFields())
	}
	if len(rec.extra) != 1 {
		t.Errorf("expected 1 overflow, got %d", len(rec.extra))
	}
}

// --- Console encoder tests ---

func TestConsoleEncoder(t *testing.T) {
	w := &testWriter{}
	h := NewConsoleHandler(
		WithConsoleWriter(w),
		WithConsoleNoColor(),
		WithConsoleLevel(TraceLevel),
	)

	logger := newTestLogger(w, h)
	logger.Info("hello world", "user", "ali", "count", 5)

	out := w.String()
	if !strings.Contains(out, "INFO") {
		t.Error("missing INFO in output")
	}
	if !strings.Contains(out, "hello world") {
		t.Error("missing message")
	}
	if !strings.Contains(out, "user=ali") {
		t.Error("missing field user=ali")
	}
	if !strings.Contains(out, "count=5") {
		t.Error("missing field count=5")
	}
	if !strings.Contains(out, "●") {
		t.Error("missing icon")
	}
}

func TestConsoleSuccess(t *testing.T) {
	w := &testWriter{}
	h := NewConsoleHandler(WithConsoleWriter(w), WithConsoleNoColor())
	logger := newTestLogger(w, h)
	logger.Success("done", "tables", 12)

	out := w.String()
	if !strings.Contains(out, "OK") {
		t.Error("missing OK level")
	}
	if !strings.Contains(out, "✓") {
		t.Error("missing success icon")
	}
}

// --- JSON encoder tests ---

func TestJSONEncoder(t *testing.T) {
	w := &testWriter{}
	h := NewJSONHandler(w)
	logger := newTestLogger(w, h)
	logger.Info("hello", "key", "val", "num", 42)

	out := w.String()
	if !strings.Contains(out, `"msg":"hello"`) {
		t.Errorf("missing message in JSON: %s", out)
	}
	if !strings.Contains(out, `"key":"val"`) {
		t.Errorf("missing field in JSON: %s", out)
	}
	if !strings.Contains(out, `"num":42`) {
		t.Errorf("missing num field in JSON: %s", out)
	}
	if !strings.Contains(out, `"level":"INFO"`) {
		t.Errorf("missing level in JSON: %s", out)
	}
}

// --- Logfmt encoder tests ---

func TestLogfmtEncoder(t *testing.T) {
	w := &testWriter{}
	h := NewLogfmtHandler(w)
	logger := newTestLogger(w, h)
	logger.Info("hello world", "key", "val")

	out := w.String()
	if !strings.Contains(out, "level=info") {
		t.Errorf("missing level in logfmt: %s", out)
	}
	if !strings.Contains(out, `msg="hello world"`) {
		t.Errorf("missing msg in logfmt: %s", out)
	}
	if !strings.Contains(out, "key=val") {
		t.Errorf("missing field in logfmt: %s", out)
	}
}

// --- Logger tests ---

func TestLevelFiltering(t *testing.T) {
	w := &testWriter{}
	h := NewConsoleHandler(WithConsoleWriter(w), WithConsoleNoColor())
	logger := New(
		WithHandler(h),
		WithLevel(WarnLevel),
		WithCaller(false),
	)

	logger.Info("should not appear")
	if w.String() != "" {
		t.Error("Info should be filtered at WarnLevel")
	}

	logger.Warn("should appear")
	if !strings.Contains(w.String(), "should appear") {
		t.Error("Warn should pass at WarnLevel")
	}
}

func TestWithFields(t *testing.T) {
	w := &testWriter{}
	h := NewConsoleHandler(WithConsoleWriter(w), WithConsoleNoColor())
	logger := newTestLogger(w, h)

	child := logger.WithFields(Fields{"service": "api"})
	child.Info("request")

	out := w.String()
	if !strings.Contains(out, "service=api") {
		t.Errorf("missing pre-bound field: %s", out)
	}
}

func TestWithContext(t *testing.T) {
	w := &testWriter{}
	h := NewJSONHandler(w)
	logger := newTestLogger(w, h)

	ctx := ContextWithFields(context.Background(), String("request_id", "abc-123"))
	logger.WithContext(ctx).Info("processing")

	out := w.String()
	if !strings.Contains(out, `"request_id":"abc-123"`) {
		t.Errorf("missing context field: %s", out)
	}
}

func TestMultiHandler(t *testing.T) {
	w1 := &testWriter{}
	w2 := &testWriter{}
	h := NewMultiHandler(
		NewConsoleHandler(WithConsoleWriter(w1), WithConsoleNoColor()),
		NewJSONHandler(w2),
	)
	logger := newTestLogger(w1, h)
	logger.Info("multi")

	if !strings.Contains(w1.String(), "multi") {
		t.Error("console handler missing output")
	}
	if !strings.Contains(w2.String(), "multi") {
		t.Error("json handler missing output")
	}
}

func TestCallerCapture(t *testing.T) {
	w := &testWriter{}
	h := NewConsoleHandler(WithConsoleWriter(w), WithConsoleNoColor())
	logger := New(
		WithHandler(h),
		WithLevel(TraceLevel),
		WithCaller(true),
		WithStackLevel(FatalLevel+1),
	)

	logger.Info("with caller")
	out := w.String()
	if !strings.Contains(out, "caller=") {
		t.Errorf("missing caller info: %s", out)
	}
	if !strings.Contains(out, "loghq_test.go:") {
		t.Errorf("wrong caller file: %s", out)
	}
}

func TestDurationField(t *testing.T) {
	w := &testWriter{}
	h := NewJSONHandler(w)
	logger := newTestLogger(w, h)

	logger.Info("timing", "elapsed", time.Second*3)

	out := w.String()
	if !strings.Contains(out, `"elapsed":"3s"`) {
		t.Errorf("missing duration field: %s", out)
	}
}

func TestSetLevel(t *testing.T) {
	w := &testWriter{}
	h := NewConsoleHandler(WithConsoleWriter(w), WithConsoleNoColor())
	logger := newTestLogger(w, h)

	logger.SetLevel(ErrorLevel)
	logger.Info("hidden")
	if w.String() != "" {
		t.Error("should be filtered")
	}

	logger.SetLevel(TraceLevel)
	logger.Info("visible")
	if !strings.Contains(w.String(), "visible") {
		t.Error("should be visible after level change")
	}
}
