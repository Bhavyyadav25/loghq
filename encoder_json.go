package loghq

import "time"

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

	// Fields
	if rec.NumFields() > 0 {
		fe := jsonFieldEnc{buf: buf}
		rec.EachField(func(f *Field) {
			buf.AppendByte(',')
			f.Encode(&fe)
		})
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

// jsonFieldEnc is a stack-local FieldEncoder for JSON output.
type jsonFieldEnc struct {
	buf *Buffer
}

func (e *jsonFieldEnc) writeKey(key string) {
	e.buf.AppendByte('"')
	e.buf.AppendString(key)
	e.buf.AppendString(`":`)
}

func (e *jsonFieldEnc) EncodeString(key, val string) {
	e.writeKey(key)
	appendJSONString(e.buf, val)
}

func (e *jsonFieldEnc) EncodeInt64(key string, val int64) {
	e.writeKey(key)
	e.buf.AppendInt(val)
}

func (e *jsonFieldEnc) EncodeFloat64(key string, val float64) {
	e.writeKey(key)
	e.buf.AppendFloat(val)
}

func (e *jsonFieldEnc) EncodeBool(key string, val bool) {
	e.writeKey(key)
	e.buf.AppendBool(val)
}

func (e *jsonFieldEnc) EncodeDuration(key string, val time.Duration) {
	e.writeKey(key)
	appendJSONString(e.buf, val.String())
}

func (e *jsonFieldEnc) EncodeTime(key string, val time.Time) {
	e.writeKey(key)
	e.buf.AppendByte('"')
	e.buf.AppendTime(val, time.RFC3339Nano)
	e.buf.AppendByte('"')
}

func (e *jsonFieldEnc) EncodeError(key string, msg string) {
	e.writeKey(key)
	appendJSONString(e.buf, msg)
}

func (e *jsonFieldEnc) EncodeAny(key string, val interface{}) {
	e.writeKey(key)
	appendJSONString(e.buf, formatAny(val))
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
