package loghq

// MultiHandler dispatches records to multiple handlers.
// Continues to all handlers even if one returns an error.
type MultiHandler struct {
	handlers []Handler
}

// NewMultiHandler creates a handler that fans out to multiple handlers.
func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(lvl Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(lvl) {
			return true
		}
	}
	return false
}

// Handle sends the record to all enabled handlers, returning the first error.
func (m *MultiHandler) Handle(rec *Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if h.Enabled(rec.Level) {
			if err := h.Handle(rec); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Flush flushes all handlers that implement Flusher.
func (m *MultiHandler) Flush() error {
	var firstErr error
	for _, h := range m.handlers {
		if f, ok := h.(Flusher); ok {
			if err := f.Flush(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Close closes all handlers that implement Closer.
func (m *MultiHandler) Close() error {
	var firstErr error
	for _, h := range m.handlers {
		if c, ok := h.(Closer); ok {
			if err := c.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
