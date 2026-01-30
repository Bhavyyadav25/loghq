package loghq

import (
	"math"
	"time"
)

// JSONEncoder writes records as JSON without using encoding/json.
// Thread-safe: no mutable state stored between Encode calls.
type JSONEncoder struct {
	TimeKey    string
	LevelKey   string
	MessageKey string
	CallerKey  string
	StackKey   string
	TimeLayout string
}

func (e *JSONEncoder) key(custom, fallback string) string {
	if custom != "" {
		return custom
	}
	return fallback
}

func (e *JSONEncoder) timeLayout() string {
	if e.TimeLayout != "" {
		return e.TimeLayout
	}
	return time.RFC3339Nano
}

// Encode writes a full JSON record. Thread-safe.
func (e *JSONEncoder) Encode(buf *Buffer, rec *Record) {
	buf.AppendByte('{')

	// Time
	buf.AppendByte('"')
	buf.AppendString(e.key(e.TimeKey, "time"))
	buf.AppendString(`":"`)
	buf.AppendTime(rec.Time, e.timeLayout())
	buf.AppendByte('"')

	// Level
	buf.AppendString(`,"`)
	buf.AppendString(e.key(e.LevelKey, "level"))
	buf.AppendString(`":"`)
	buf.AppendString(rec.Level.String())
	buf.AppendByte('"')

	// Message
	buf.AppendString(`,"`)
	buf.AppendString(e.key(e.MessageKey, "msg"))
	buf.AppendString(`":`)
	appendJSONString(buf, rec.Message)

	// Caller
	if rec.Caller.Defined() {
		buf.AppendString(`,"`)
		buf.AppendString(e.key(e.CallerKey, "caller"))
		buf.AppendString(`":"`)
		buf.AppendString(rec.Caller.String())
		buf.AppendByte('"')
	}

	// Fields — direct encoding avoids interface escape to heap
	for i, nf := 0, rec.NumFields(); i < nf; i++ {
		buf.AppendByte(',')
		f := rec.FieldAt(i)
		e.encodeField(buf, f)
	}

	// Stack
	if rec.Stack != "" {
		buf.AppendString(`,"`)
		buf.AppendString(e.key(e.StackKey, "stack"))
		buf.AppendString(`":`)
		appendJSONString(buf, rec.Stack)
	}

	buf.AppendString("}\n")
}

// encodeField encodes a single field directly without going through the
// FieldEncoder interface, avoiding heap escape of the receiver.
func (e *JSONEncoder) encodeField(buf *Buffer, f *Field) {
	// key
	buf.AppendByte('"')
	buf.AppendString(f.Key)
	buf.AppendString(`":`)
	// value
	switch f.Type {
	case FieldString:
		appendJSONString(buf, f.Str)
	case FieldInt64:
		buf.AppendInt(f.Ival)
	case FieldFloat64:
		buf.AppendFloat(math.Float64frombits(uint64(f.Ival)))
	case FieldBool:
		buf.AppendBool(f.Ival == 1)
	case FieldDuration:
		appendJSONString(buf, time.Duration(f.Ival).String())
	case FieldTime:
		if t, ok := f.Iface.(time.Time); ok {
			buf.AppendByte('"')
			buf.AppendTime(t, time.RFC3339Nano)
			buf.AppendByte('"')
		}
	case FieldError:
		appendJSONString(buf, f.Str)
	case FieldAny:
		appendJSONString(buf, formatAny(f.Iface))
	}
}

// --- JSON helpers ---

func appendJSONString(buf *Buffer, s string) {
	buf.AppendByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf.AppendString(`\"`)
		case '\\':
			buf.AppendString(`\\`)
		case '\n':
			buf.AppendString(`\n`)
		case '\r':
			buf.AppendString(`\r`)
		case '\t':
			buf.AppendString(`\t`)
		default:
			if c < 0x20 {
				buf.AppendString(`\u00`)
				buf.AppendByte(hexChar(c >> 4))
				buf.AppendByte(hexChar(c & 0x0f))
			} else {
				buf.AppendByte(c)
			}
		}
	}
	buf.AppendByte('"')
}

func hexChar(c byte) byte {
	if c < 10 {
		return '0' + c
	}
	return 'a' + c - 10
}
