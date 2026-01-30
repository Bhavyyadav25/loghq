package loghq

import (
	"strings"
	"time"
)

// LogfmtEncoder writes records in logfmt format (key=value pairs).
// Thread-safe: no mutable state stored between Encode calls.
type LogfmtEncoder struct {
	TimeLayout string
}

func (e *LogfmtEncoder) timeLayout() string {
	if e.TimeLayout != "" {
		return e.TimeLayout
	}
	return time.RFC3339
}

// Encode writes a full logfmt record. Thread-safe.
func (e *LogfmtEncoder) Encode(buf *Buffer, rec *Record) {
	// time=
	buf.AppendString("time=")
	buf.AppendTime(rec.Time, e.timeLayout())

	// level=
	buf.AppendString(" level=")
	buf.AppendString(strings.ToLower(rec.Level.String()))

	// msg=
	buf.AppendString(" msg=")
	appendLogfmtValue(buf, rec.Message)

	// caller=
	if rec.Caller.Defined() {
		buf.AppendString(" caller=")
		buf.AppendString(rec.Caller.String())
	}

	// Fields
	if rec.NumFields() > 0 {
		fe := logfmtFieldEnc{buf: buf}
		rec.EachField(func(f *Field) {
			buf.AppendByte(' ')
			f.Encode(&fe)
		})
	}

	buf.AppendByte('\n')
}

// logfmtFieldEnc is a stack-local FieldEncoder for logfmt output.
type logfmtFieldEnc struct {
	buf *Buffer
}

func (e *logfmtFieldEnc) EncodeString(key, val string) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	appendLogfmtValue(e.buf, val)
}

func (e *logfmtFieldEnc) EncodeInt64(key string, val int64) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	e.buf.AppendInt(val)
}

func (e *logfmtFieldEnc) EncodeFloat64(key string, val float64) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	e.buf.AppendFloat(val)
}

func (e *logfmtFieldEnc) EncodeBool(key string, val bool) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	e.buf.AppendBool(val)
}

func (e *logfmtFieldEnc) EncodeDuration(key string, val time.Duration) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	e.buf.AppendString(val.String())
}

func (e *logfmtFieldEnc) EncodeTime(key string, val time.Time) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	e.buf.AppendTime(val, time.RFC3339)
}

func (e *logfmtFieldEnc) EncodeError(key string, msg string) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	appendLogfmtValue(e.buf, msg)
}

func (e *logfmtFieldEnc) EncodeAny(key string, val interface{}) {
	e.buf.AppendString(key)
	e.buf.AppendByte('=')
	appendLogfmtValue(e.buf, formatAny(val))
}

// --- Logfmt helpers ---

func appendLogfmtValue(buf *Buffer, s string) {
	if s == "" {
		buf.AppendString(`""`)
		return
	}
	needsQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ' ' || c == '"' || c == '\\' || c == '=' || c < 0x20 {
			needsQuote = true
			break
		}
	}
	if needsQuote {
		buf.AppendByte('"')
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c == '"' || c == '\\' {
				buf.AppendByte('\\')
			}
			buf.AppendByte(c)
		}
		buf.AppendByte('"')
	} else {
		buf.AppendString(s)
	}
}
