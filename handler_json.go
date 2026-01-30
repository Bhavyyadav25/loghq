package loghq

// JSONHandler writes JSON-encoded log records.
// Thin configuration wrapper over BaseHandler.
type JSONHandler struct {
	*BaseHandler
}

// NewJSONHandler creates a handler that writes JSON logs to the given writer.
func NewJSONHandler(w WriteSyncer, opts ...JSONOption) *JSONHandler {
	cfg := &jsonConfig{
		writer: w,
		level:  TraceLevel,
		enc:    &JSONEncoder{},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &JSONHandler{
		BaseHandler: NewBaseHandler(cfg.enc, cfg.writer, cfg.level),
	}
}

type jsonConfig struct {
	enc    *JSONEncoder
	writer WriteSyncer
	level  Level
}

// JSONOption configures a JSONHandler.
type JSONOption func(*jsonConfig)

// WithJSONLevel sets the minimum level.
func WithJSONLevel(l Level) JSONOption {
	return func(c *jsonConfig) { c.level = l }
}

// WithJSONTimeLayout sets the time format.
func WithJSONTimeLayout(layout string) JSONOption {
	return func(c *jsonConfig) { c.enc.TimeLayout = layout }
}

// WithJSONKeys sets the JSON key names for standard fields.
func WithJSONKeys(timeKey, levelKey, msgKey string) JSONOption {
	return func(c *jsonConfig) {
		if timeKey != "" {
			c.enc.TimeKey = timeKey
		}
		if levelKey != "" {
			c.enc.LevelKey = levelKey
		}
		if msgKey != "" {
			c.enc.MessageKey = msgKey
		}
	}
}
