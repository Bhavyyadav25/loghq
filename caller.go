package loghq

import (
	"runtime"
	"strconv"
	"strings"
)

// CallerInfo holds source location data.
type CallerInfo struct {
	File     string
	Line     int
	Function string
	defined  bool
}

// String returns "file:line".
func (c CallerInfo) String() string {
	if !c.defined {
		return ""
	}
	return c.File + ":" + strconv.Itoa(c.Line)
}

// Defined returns true if caller info was captured.
func (c CallerInfo) Defined() bool {
	return c.defined
}

// captureCaller captures the caller's file and line at the given skip depth.
func captureCaller(skip int) CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return CallerInfo{}
	}

	// Shorten file path to just the last two segments (package/file.go)
	short := file
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		if idx2 := strings.LastIndex(file[:idx], "/"); idx2 >= 0 {
			short = file[idx2+1:]
		}
	}

	fn := runtime.FuncForPC(pc)
	funcName := ""
	if fn != nil {
		funcName = fn.Name()
		if idx := strings.LastIndex(funcName, "."); idx >= 0 {
			funcName = funcName[idx+1:]
		}
	}

	return CallerInfo{
		File:     short,
		Line:     line,
		Function: funcName,
		defined:  true,
	}
}

// captureStack returns a formatted stack trace.
func captureStack(skip int) string {
	const maxDepth = 32
	var pcs [maxDepth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	var b strings.Builder
	b.Grow(512)

	for {
		frame, more := frames.Next()
		b.WriteString(frame.Function)
		b.WriteByte('\n')
		b.WriteByte('\t')
		b.WriteString(frame.File)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(frame.Line))
		b.WriteByte('\n')
		if !more {
			break
		}
	}

	return b.String()
}
