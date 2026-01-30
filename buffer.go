package loghq

import (
	"strconv"
	"sync"
	"time"
)

const defaultBufSize = 1024

// Buffer is a byte buffer with allocation-free append helpers.
// Recycled via sync.Pool for zero-alloc logging.
type Buffer struct {
	B []byte
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{B: make([]byte, 0, defaultBufSize)}
	},
}

func getBuffer() *Buffer {
	return bufferPool.Get().(*Buffer)
}

func putBuffer(b *Buffer) {
	if cap(b.B) > 16*1024 {
		return
	}
	b.B = b.B[:0]
	bufferPool.Put(b)
}

func (b *Buffer) AppendByte(c byte) {
	b.B = append(b.B, c)
}

func (b *Buffer) AppendBytes(p []byte) {
	b.B = append(b.B, p...)
}

func (b *Buffer) AppendString(s string) {
	b.B = append(b.B, s...)
}

func (b *Buffer) AppendInt(i int64) {
	b.B = strconv.AppendInt(b.B, i, 10)
}

func (b *Buffer) AppendUint(u uint64) {
	b.B = strconv.AppendUint(b.B, u, 10)
}

func (b *Buffer) AppendFloat(f float64) {
	b.B = strconv.AppendFloat(b.B, f, 'f', -1, 64)
}

func (b *Buffer) AppendBool(v bool) {
	b.B = strconv.AppendBool(b.B, v)
}

func (b *Buffer) AppendTime(t time.Time, layout string) {
	b.B = t.AppendFormat(b.B, layout)
}

func (b *Buffer) Len() int {
	return len(b.B)
}

func (b *Buffer) Bytes() []byte {
	return b.B
}

func (b *Buffer) Reset() {
	b.B = b.B[:0]
}
