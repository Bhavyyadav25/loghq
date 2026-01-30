package loghq

import "strconv"

// Level defines log severity levels.
type Level int8

const (
	TraceLevel   Level = -2
	DebugLevel   Level = -1
	InfoLevel    Level = 0
	SuccessLevel Level = 1
	WarnLevel    Level = 2
	ErrorLevel   Level = 3
	FatalLevel   Level = 4
)

var levelNames = [7]string{
	"TRACE",
	"DEBUG",
	"INFO",
	"OK",
	"WARN",
	"ERROR",
	"FATAL",
}

// String returns the human-readable level name.
func (l Level) String() string {
	idx := int(l) + 2 // TraceLevel(-2) maps to index 0
	if idx >= 0 && idx < len(levelNames) {
		return levelNames[idx]
	}
	return "LEVEL(" + strconv.Itoa(int(l)) + ")"
}

// Enabled returns true if this level is at or above the given threshold.
func (l Level) Enabled(threshold Level) bool {
	return l >= threshold
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) Level {
	switch s {
	case "trace", "TRACE":
		return TraceLevel
	case "debug", "DEBUG":
		return DebugLevel
	case "info", "INFO":
		return InfoLevel
	case "success", "SUCCESS", "ok", "OK":
		return SuccessLevel
	case "warn", "WARN", "warning", "WARNING":
		return WarnLevel
	case "error", "ERROR":
		return ErrorLevel
	case "fatal", "FATAL":
		return FatalLevel
	default:
		return InfoLevel
	}
}
