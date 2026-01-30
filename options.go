package loghq

// Option configures a Logger.
type Option func(*Logger)

// WithLevel sets the minimum log level.
func WithLevel(l Level) Option {
	return func(lg *Logger) {
		lg.level.Store(int32(l))
	}
}

// WithHandler sets the log handler.
func WithHandler(h Handler) Option {
	return func(lg *Logger) {
		lg.handler = h
	}
}

// WithCaller enables/disables automatic caller info capture.
func WithCaller(on bool) Option {
	return func(lg *Logger) {
		lg.addCaller = on
	}
}

// WithStackLevel sets the minimum level for stack trace capture.
func WithStackLevel(l Level) Option {
	return func(lg *Logger) {
		lg.stackLevel = l
	}
}

// WithCallerSkip adds additional frames to skip when capturing caller info.
func WithCallerSkip(skip int) Option {
	return func(lg *Logger) {
		lg.callerSkip = skip
	}
}
