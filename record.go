package loghq

import (
	"fmt"
	"sync"
	"time"
)

const inlineFieldCap = 16

// Record represents a single log event. The inline Fields array avoids heap
// allocation for log calls with 16 or fewer fields.
type Record struct {
	Time    time.Time
	Level   Level
	Message string
	Caller  CallerInfo
	Stack   string

	// Inline storage for up to 16 fields — zero allocation.
	fields  [inlineFieldCap]Field
	nFields int

	// Overflow for >16 fields (rare).
	extra []Field
}

var recordPool = sync.Pool{
	New: func() interface{} {
		return &Record{}
	},
}

func acquireRecord() *Record {
	r := recordPool.Get().(*Record)
	r.reset()
	return r
}

func releaseRecord(r *Record) {
	recordPool.Put(r)
}

func (r *Record) reset() {
	r.Time = time.Time{}
	r.Level = InfoLevel
	r.Message = ""
	r.Caller = CallerInfo{}
	r.Stack = ""
	r.nFields = 0
	r.extra = r.extra[:0:0] // Reset length AND capacity to prevent pool memory bloat
}

// AddField appends a field to the record.
func (r *Record) AddField(f Field) {
	if r.nFields < inlineFieldCap {
		r.fields[r.nFields] = f
		r.nFields++
	} else {
		r.extra = append(r.extra, f)
	}
}

// AddFields appends multiple fields.
func (r *Record) AddFields(fs []Field) {
	for i := range fs {
		r.AddField(fs[i])
	}
}

// NumFields returns the total number of fields.
func (r *Record) NumFields() int {
	return r.nFields + len(r.extra)
}

// FieldAt returns a pointer to the i-th field (0-indexed).
// Panics if i is out of range.
func (r *Record) FieldAt(i int) *Field {
	if i < r.nFields {
		return &r.fields[i]
	}
	return &r.extra[i-r.nFields]
}

// AddKVPairs parses slog-style key-value pairs directly into the inline field
// array. This avoids allocating an intermediate []Field slice.
func (r *Record) AddKVPairs(kvs []interface{}) {
	n := len(kvs)
	for i := 0; i < n; i += 2 {
		var key string
		switch k := kvs[i].(type) {
		case string:
			key = k
		default:
			key = fmt.Sprint(kvs[i])
		}
		if i+1 >= n {
			r.AddField(Field{Key: key, Type: FieldString, Str: "MISSING"})
			break
		}
		r.AddField(toField(key, kvs[i+1]))
	}
}

// EachField calls fn for every field in order.
func (r *Record) EachField(fn func(f *Field)) {
	for i := 0; i < r.nFields; i++ {
		fn(&r.fields[i])
	}
	for i := range r.extra {
		fn(&r.extra[i])
	}
}
