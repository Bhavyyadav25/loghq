package loghq

import "sync/atomic"

// Handler processes log records. Minimal interface per ISP —
// only the two methods every handler must have.
type Handler interface {
	Enabled(Level) bool
	Handle(rec *Record) error
}

// Flusher is optionally implemented by handlers that buffer output.
type Flusher interface {
	Flush() error
}

// Closer is optionally implemented by handlers that hold resources.
type Closer interface {
	Close() error
}

// BaseHandler composes an Encoder, WriteSyncer, and level filter.
// Concrete handlers embed this to eliminate boilerplate.
type BaseHandler struct {
	enc    Encoder
	writer WriteSyncer
	level  atomic.Int32
}

// NewBaseHandler creates a handler with the given encoder, writer, and level.
func NewBaseHandler(enc Encoder, w WriteSyncer, lvl Level) *BaseHandler {
	h := &BaseHandler{enc: enc, writer: w}
	h.level.Store(int32(lvl))
	return h
}

// Enabled returns true if the level passes the filter.
func (h *BaseHandler) Enabled(lvl Level) bool {
	return lvl >= Level(h.level.Load())
}

// Handle encodes the record and writes it. Buffer is pooled for zero-alloc.
func (h *BaseHandler) Handle(rec *Record) error {
	buf := getBuffer()
	h.enc.Encode(buf, rec)
	_, err := h.writer.Write(buf.Bytes())
	putBuffer(buf)
	return err
}

// Flush syncs the underlying writer.
func (h *BaseHandler) Flush() error {
	return h.writer.Sync()
}

// Close syncs the underlying writer.
func (h *BaseHandler) Close() error {
	return h.writer.Sync()
}

// SetLevel changes the handler's level atomically.
func (h *BaseHandler) SetLevel(lvl Level) {
	h.level.Store(int32(lvl))
}
