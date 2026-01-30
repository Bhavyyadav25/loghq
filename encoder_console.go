package loghq

import (
	"math"
	"time"
)

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

	// Fields — direct encoding avoids interface escape to heap
	if nf := rec.NumFields(); nf > 0 {
		buf.AppendString("  ")
		for i := 0; i < nf; i++ {
			if i > 0 {
				buf.AppendByte(' ')
			}
			e.encodeField(buf, rec.FieldAt(i))
		}
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

func (e *ConsoleEncoder) appendFieldKey(buf *Buffer, key string) {
	if e.NoColor {
		buf.AppendString(key)
		buf.AppendByte('=')
	} else {
		buf.AppendString(colorDim)
		buf.AppendString(key)
		buf.AppendString("=" + colorReset)
	}
}

// encodeField encodes a single field directly without going through the
// FieldEncoder interface, avoiding heap escape.
func (e *ConsoleEncoder) encodeField(buf *Buffer, f *Field) {
	e.appendFieldKey(buf, f.Key)
	switch f.Type {
	case FieldString:
		buf.AppendString(f.Str)
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
		buf.AppendString(f.Str)
	case FieldAny:
		buf.AppendString(formatAny(f.Iface))
	}
}
