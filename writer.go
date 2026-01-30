package loghq

import (
	"io"
	"os"
	"sync"
)

// WriteSyncer extends io.Writer with a Sync method.
type WriteSyncer interface {
	io.Writer
	Sync() error
}

// LockedWriter wraps an io.Writer with a mutex for thread-safe writes.
type LockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func NewLockedWriter(w io.Writer) *LockedWriter {
	return &LockedWriter{w: w}
}

func (lw *LockedWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	n, err := lw.w.Write(p)
	lw.mu.Unlock()
	return n, err
}

func (lw *LockedWriter) Sync() error {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	if s, ok := lw.w.(interface{ Sync() error }); ok {
		return s.Sync()
	}
	return nil
}

// writerSyncer wraps an io.Writer to implement WriteSyncer.
type writerSyncer struct {
	io.Writer
}

func (w writerSyncer) Sync() error {
	if s, ok := w.Writer.(interface{ Sync() error }); ok {
		return s.Sync()
	}
	return nil
}

// WrapWriter converts an io.Writer into a WriteSyncer.
func WrapWriter(w io.Writer) WriteSyncer {
	if ws, ok := w.(WriteSyncer); ok {
		return ws
	}
	return writerSyncer{w}
}

// Stdout and Stderr as WriteSyncer.
var (
	Stdout WriteSyncer = &fileWriteSyncer{os.Stdout}
	Stderr WriteSyncer = &fileWriteSyncer{os.Stderr}
)

type fileWriteSyncer struct {
	f *os.File
}

func (fw *fileWriteSyncer) Write(p []byte) (int, error) {
	return fw.f.Write(p)
}

func (fw *fileWriteSyncer) Sync() error {
	return fw.f.Sync()
}
