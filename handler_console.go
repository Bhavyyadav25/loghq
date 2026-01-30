package loghq

import "os"

// ConsoleHandler writes colored, icon-prefixed log output.
// Thin configuration wrapper over BaseHandler.
type ConsoleHandler struct {
	*BaseHandler
}

// NewConsoleHandler creates a handler for beautiful terminal output.
func NewConsoleHandler(opts ...ConsoleOption) *ConsoleHandler {
	cfg := &consoleConfig{
		writer: Stderr,
		level:  TraceLevel,
		enc:    &ConsoleEncoder{TimeLayout: defaultTimeLayout},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return &ConsoleHandler{
		BaseHandler: NewBaseHandler(cfg.enc, cfg.writer, cfg.level),
	}
}

// consoleConfig holds construction-time configuration.
type consoleConfig struct {
	enc    *ConsoleEncoder
	writer WriteSyncer
	level  Level
}

// ConsoleOption configures a ConsoleHandler.
type ConsoleOption func(*consoleConfig)

// WithConsoleWriter sets the output writer.
func WithConsoleWriter(w WriteSyncer) ConsoleOption {
	return func(c *consoleConfig) { c.writer = w }
}

// WithConsoleNoColor disables ANSI colors.
func WithConsoleNoColor() ConsoleOption {
	return func(c *consoleConfig) { c.enc.NoColor = true }
}

// WithConsoleTimeLayout sets the time format.
func WithConsoleTimeLayout(layout string) ConsoleOption {
	return func(c *consoleConfig) { c.enc.TimeLayout = layout }
}

// WithConsoleLevel sets the minimum level.
func WithConsoleLevel(l Level) ConsoleOption {
	return func(c *consoleConfig) { c.level = l }
}

// WithConsoleStdout writes to stdout instead of stderr.
func WithConsoleStdout() ConsoleOption {
	return func(c *consoleConfig) { c.writer = &fileWriteSyncer{os.Stdout} }
}
