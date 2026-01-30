package loghq

import (
	"math"
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

	// Fields — direct encoding avoids interface escape to heap
	for i, nf := 0, rec.NumFields(); i < nf; i++ {
		buf.AppendByte(' ')
		e.encodeField(buf, rec.FieldAt(i))
	}

	buf.AppendByte('\n')
}

// encodeField encodes a single field directly without going through the
// FieldEncoder interface, avoiding heap escape.
func (e *LogfmtEncoder) encodeField(buf *Buffer, f *Field) {
	buf.AppendString(f.Key)
	buf.AppendByte('=')
	switch f.Type {
	case FieldString:
		appendLogfmtValue(buf, f.Str)
	case FieldInt64:
		buf.AppendInt(f.Ival)
	case FieldFloat64:
		buf.AppendFloat(math.Float64frombits(uint64(f.Ival)))
	case FieldBool:
		buf.AppendBool(f.Ival == 1)
	case FieldDuration:
		buf.AppendString(time.Duration(f.Ival).String())
	case FieldTime:
		if t, ok := f.Iface.(time.Time); ok {
			buf.AppendTime(t, time.RFC3339)
		}
	case FieldError:
		appendLogfmtValue(buf, f.Str)
	case FieldAny:
		appendLogfmtValue(buf, formatAny(f.Iface))
	}
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
