package loghq

import "time"

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"

	colorBoldRed = "\033[1;31m"
)

// Level rendering metadata.
var levelIcons = [7]string{"◦", "◇", "●", "✓", "▲", "✗", "✗"}
var levelColors = [7]string{colorGray, colorCyan, colorBlue, colorGreen, colorYellow, colorRed, colorBoldRed}
var levelPadded = [7]string{"TRACE ", "DEBUG ", "INFO  ", "OK    ", "WARN  ", "ERROR ", "FATAL "}

const defaultTimeLayout = "2006-01-02 15:04:05"

// ConsoleEncoder writes records as colored, icon-prefixed terminal output.
// Thread-safe: no mutable state stored between Encode calls.
type ConsoleEncoder struct {
	NoColor    bool
	TimeLayout string
}

func (e *ConsoleEncoder) timeLayout() string {
	if e.TimeLayout != "" {
		return e.TimeLayout
	}
	return defaultTimeLayout
}

func (e *ConsoleEncoder) levelIndex(lvl Level) int {
	idx := int(lvl) + 2
	if idx < 0 {
		return 0
	}
	if idx >= len(levelIcons) {
		return len(levelIcons) - 1
	}
	return idx
}

// Encode writes a full record to buf. Thread-safe.
func (e *ConsoleEncoder) Encode(buf *Buffer, rec *Record) {
	idx := e.levelIndex(rec.Level)

	// Timestamp
	if e.NoColor {
		buf.AppendString(" ")
	} else {
		buf.AppendString(colorDim + " ")
	}
	buf.AppendTime(rec.Time, e.timeLayout())
	if !e.NoColor {
		buf.AppendString(colorReset)
	}
	buf.AppendByte(' ')

	// Level icon + name
	if e.NoColor {
		buf.AppendString(levelIcons[idx])
		buf.AppendByte(' ')
		buf.AppendString(levelPadded[idx])
	} else {
		buf.AppendString(levelColors[idx])
		buf.AppendString(levelIcons[idx])
		buf.AppendByte(' ')
		buf.AppendString(levelPadded[idx])
		buf.AppendString(colorReset)
	}

	// Message
	buf.AppendString(rec.Message)

	// Fields
	if rec.NumFields() > 0 {
		buf.AppendString("  ")
		fe := consoleFieldEnc{buf: buf, noColor: e.NoColor, first: true}
		rec.EachField(func(f *Field) {
			if !fe.first {
				buf.AppendByte(' ')
			}
			fe.first = false
			fe.encodeField(f)
		})
	}

	// Caller
	if rec.Caller.Defined() {
		buf.AppendString("  ")
		if e.NoColor {
			buf.AppendString("caller=")
		} else {
			buf.AppendString(colorDim + "caller=" + colorReset)
		}
		buf.AppendString(rec.Caller.String())
	}

	buf.AppendByte('\n')

	// Stack trace
	if rec.Stack != "" {
		if e.NoColor {
			buf.AppendString(rec.Stack)
		} else {
			buf.AppendString(colorDim)
			buf.AppendString(rec.Stack)
			buf.AppendString(colorReset)
		}
	}
}

// consoleFieldEnc is a stack-local FieldEncoder for console output.
// Created per Encode call — no shared mutable state.
type consoleFieldEnc struct {
	buf     *Buffer
	noColor bool
	first   bool
}

// encodeField dispatches the field to the correct typed method.
// Uses a concrete receiver to avoid interface boxing (zero-alloc).
func (e *consoleFieldEnc) encodeField(f *Field) {
	f.Encode(e)
}

func (e *consoleFieldEnc) appendKey(key string) {
	if e.noColor {
		e.buf.AppendString(key)
		e.buf.AppendByte('=')
	} else {
		e.buf.AppendString(colorDim)
		e.buf.AppendString(key)
		e.buf.AppendString("=" + colorReset)
	}
}

func (e *consoleFieldEnc) EncodeString(key, val string) {
	e.appendKey(key)
	e.buf.AppendString(val)
}

func (e *consoleFieldEnc) EncodeInt64(key string, val int64) {
	e.appendKey(key)
	e.buf.AppendInt(val)
}

func (e *consoleFieldEnc) EncodeFloat64(key string, val float64) {
	e.appendKey(key)
	e.buf.AppendFloat(val)
}

func (e *consoleFieldEnc) EncodeBool(key string, val bool) {
	e.appendKey(key)
	e.buf.AppendBool(val)
}

func (e *consoleFieldEnc) EncodeDuration(key string, val time.Duration) {
	e.appendKey(key)
	e.buf.AppendString(val.String())
}

func (e *consoleFieldEnc) EncodeTime(key string, val time.Time) {
	e.appendKey(key)
	e.buf.AppendTime(val, time.RFC3339)
}

func (e *consoleFieldEnc) EncodeError(key string, msg string) {
	e.appendKey(key)
	e.buf.AppendString(msg)
}

func (e *consoleFieldEnc) EncodeAny(key string, val interface{}) {
	e.appendKey(key)
	e.buf.AppendString(formatAny(val))
}
