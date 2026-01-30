package loghq

import (
	"fmt"
	"math"
	"time"
)

// FieldType identifies the type stored in a Field.
type FieldType uint8

const (
	FieldString FieldType = iota
	FieldInt64
	FieldFloat64
	FieldBool
	FieldError
	FieldDuration
	FieldTime
	FieldAny
)

// Field is a typed key-value pair. Using a tagged union avoids interface boxing
// for primitive types, enabling zero-allocation logging on the hot path.
type Field struct {
	Key   string
	Type  FieldType
	Ival  int64
	Str   string
	Iface interface{}
}

// Fields is a convenience alias for a map of key-value pairs.
type Fields map[string]interface{}

// --- Typed constructors (zero-alloc for primitives) ---

func String(key, val string) Field {
	return Field{Key: key, Type: FieldString, Str: val}
}

func Int(key string, val int) Field {
	return Field{Key: key, Type: FieldInt64, Ival: int64(val)}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Type: FieldInt64, Ival: val}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Type: FieldFloat64, Ival: int64(math.Float64bits(val))}
}

func Bool(key string, val bool) Field {
	var v int64
	if val {
		v = 1
	}
	return Field{Key: key, Type: FieldBool, Ival: v}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Type: FieldString, Str: "<nil>"}
	}
	return Field{Key: "error", Type: FieldError, Str: err.Error()}
}

func Duration(key string, d time.Duration) Field {
	return Field{Key: key, Type: FieldDuration, Ival: int64(d)}
}

func Time(key string, t time.Time) Field {
	return Field{Key: key, Type: FieldTime, Iface: t}
}

func Any(key string, val interface{}) Field {
	return Field{Key: key, Type: FieldAny, Iface: val}
}

// parseKVPairs converts slog-style alternating key-value pairs into typed Fields.
// Uses type switches instead of reflection for zero-alloc on common types.
func parseKVPairs(kvs []interface{}) []Field {
	n := len(kvs)
	if n == 0 {
		return nil
	}

	// Pre-allocate: at most n/2 fields
	fields := make([]Field, 0, n/2)

	for i := 0; i < n; i += 2 {
		// Get key
		var key string
		switch k := kvs[i].(type) {
		case string:
			key = k
		default:
			key = fmt.Sprint(kvs[i])
		}

		// If there's no value (odd number of args), use missing marker
		if i+1 >= n {
			fields = append(fields, Field{Key: key, Type: FieldString, Str: "MISSING"})
			break
		}

		val := kvs[i+1]
		fields = append(fields, toField(key, val))
	}

	return fields
}

// toField converts a single key-value pair to a typed Field.
func toField(key string, val interface{}) Field {
	switch v := val.(type) {
	case string:
		return Field{Key: key, Type: FieldString, Str: v}
	case int:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case int64:
		return Field{Key: key, Type: FieldInt64, Ival: v}
	case int32:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case int16:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case int8:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case uint:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case uint64:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case uint32:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case uint16:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case uint8:
		return Field{Key: key, Type: FieldInt64, Ival: int64(v)}
	case float64:
		return Field{Key: key, Type: FieldFloat64, Ival: int64(math.Float64bits(v))}
	case float32:
		return Field{Key: key, Type: FieldFloat64, Ival: int64(math.Float64bits(float64(v)))}
	case bool:
		var iv int64
		if v {
			iv = 1
		}
		return Field{Key: key, Type: FieldBool, Ival: iv}
	case error:
		if v == nil {
			return Field{Key: key, Type: FieldString, Str: "<nil>"}
		}
		return Field{Key: key, Type: FieldError, Str: v.Error()}
	case time.Duration:
		return Field{Key: key, Type: FieldDuration, Ival: int64(v)}
	case time.Time:
		return Field{Key: key, Type: FieldTime, Iface: v}
	case Field:
		// Allow passing typed Field directly
		v.Key = key
		return v
	default:
		return Field{Key: key, Type: FieldAny, Iface: v}
	}
}

// Encode dispatches this field's value to the appropriate FieldEncoder method.
// This is the single place where FieldType is switched for encoding,
// eliminating duplication across all encoder implementations.
func (f *Field) Encode(enc FieldEncoder) {
	switch f.Type {
	case FieldString:
		enc.EncodeString(f.Key, f.Str)
	case FieldInt64:
		enc.EncodeInt64(f.Key, f.Ival)
	case FieldFloat64:
		enc.EncodeFloat64(f.Key, math.Float64frombits(uint64(f.Ival)))
	case FieldBool:
		enc.EncodeBool(f.Key, f.Ival == 1)
	case FieldDuration:
		enc.EncodeDuration(f.Key, time.Duration(f.Ival))
	case FieldTime:
		if t, ok := f.Iface.(time.Time); ok {
			enc.EncodeTime(f.Key, t)
		}
	case FieldError:
		enc.EncodeError(f.Key, f.Str)
	case FieldAny:
		enc.EncodeAny(f.Key, f.Iface)
	}
}

// fieldsFromMap converts a Fields map into a slice of typed Fields.
func fieldsFromMap(m Fields) []Field {
	fields := make([]Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, toField(k, v))
	}
	return fields
}

// formatAny formats an arbitrary value as a string.
func formatAny(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	if s, ok := v.(interface{ String() string }); ok {
		return s.String()
	}
	return fmt.Sprint(v)
}
