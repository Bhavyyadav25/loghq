package loghq

// LogfmtHandler writes logfmt-encoded log records.
// Thin configuration wrapper over BaseHandler.
type LogfmtHandler struct {
	*BaseHandler
}

// NewLogfmtHandler creates a handler that writes logfmt logs to the given writer.
func NewLogfmtHandler(w WriteSyncer, opts ...LogfmtOption) *LogfmtHandler {
	cfg := &logfmtConfig{
		writer: w,
		level:  TraceLevel,
		enc:    &LogfmtEncoder{},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &LogfmtHandler{
		BaseHandler: NewBaseHandler(cfg.enc, cfg.writer, cfg.level),
	}
}

type logfmtConfig struct {
	enc    *LogfmtEncoder
	writer WriteSyncer
	level  Level
}

// LogfmtOption configures a LogfmtHandler.
type LogfmtOption func(*logfmtConfig)

// WithLogfmtLevel sets the minimum level.
func WithLogfmtLevel(l Level) LogfmtOption {
	return func(c *logfmtConfig) { c.level = l }
}
