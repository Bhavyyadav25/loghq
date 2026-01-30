package loghq

// discardHandler is a no-op handler. Enabled returns true so that
// the logger's own level check is the sole gatekeeper.
type discardHandler struct{}

func (discardHandler) Enabled(Level) bool       { return true }
func (discardHandler) Handle(*Record) error     { return nil }
