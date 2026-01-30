package loghq

import "time"

// Encoder serializes a Record into a Buffer.
type Encoder interface {
	Encode(buf *Buffer, rec *Record)
}

// FieldEncoder receives typed field values during encoding.
// Each encoder format (JSON, console, logfmt) provides a factory that creates
// a buffer-bound FieldEncoder. This eliminates duplicated type-switch logic —
// Field.Encode dispatches once, encoders format once.
//
// Implementations must be safe for single-goroutine use within one Encode call.
// Thread safety is achieved by creating a new FieldEncoder per Encode invocation.
type FieldEncoder interface {
	EncodeString(key, val string)
	EncodeInt64(key string, val int64)
	EncodeFloat64(key string, val float64)
	EncodeBool(key string, val bool)
	EncodeDuration(key string, val time.Duration)
	EncodeTime(key string, val time.Time)
	EncodeError(key string, msg string)
	EncodeAny(key string, val interface{})
}
