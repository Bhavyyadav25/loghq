package loghq

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileConfig configures the FileWriter.
type FileConfig struct {
	// Path is the log file path.
	Path string

	// MaxSize is the maximum size in bytes before rotation. Default: 100MB.
	MaxSize int64

	// MaxAge is how long to keep old log files. Default: 7 days. 0 means no limit.
	MaxAge time.Duration

	// MaxBackups is the maximum number of old log files to keep. Default: 5. 0 means no limit.
	MaxBackups int

	// Compress enables gzip compression of rotated files.
	Compress bool
}

func (c *FileConfig) maxSize() int64 {
	if c.MaxSize > 0 {
		return c.MaxSize
	}
	return 100 * 1024 * 1024 // 100MB
}

func (c *FileConfig) maxAge() time.Duration {
	if c.MaxAge > 0 {
		return c.MaxAge
	}
	return 7 * 24 * time.Hour
}

func (c *FileConfig) maxBackups() int {
	if c.MaxBackups > 0 {
		return c.MaxBackups
	}
	return 5
}

// FileWriter implements WriteSyncer with size-based rotation.
type FileWriter struct {
	cfg  FileConfig
	mu   sync.Mutex
	file *os.File
	size int64
}

// NewFileWriter opens a log file with rotation support.
func NewFileWriter(cfg FileConfig) (*FileWriter, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("loghq: file path is required")
	}

	// Ensure directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("loghq: cannot create directory %s: %w", dir, err)
	}

	fw := &FileWriter{cfg: cfg}
	if err := fw.openFile(); err != nil {
		return nil, err
	}
	return fw, nil
}

func (fw *FileWriter) openFile() error {
	f, err := os.OpenFile(fw.cfg.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("loghq: cannot open file %s: %w", fw.cfg.Path, err)
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}

	fw.file = f
	fw.size = info.Size()
	return nil
}

func (fw *FileWriter) Write(p []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.size+int64(len(p)) > fw.cfg.maxSize() {
		if err := fw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := fw.file.Write(p)
	fw.size += int64(n)
	return n, err
}

func (fw *FileWriter) Sync() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.file != nil {
		return fw.file.Sync()
	}
	return nil
}

// Close closes the file.
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if fw.file != nil {
		return fw.file.Close()
	}
	return nil
}

func (fw *FileWriter) rotate() error {
	// Close current file
	if fw.file != nil {
		fw.file.Close()
	}

	// Rename to timestamped backup
	ts := time.Now().Format("2006-01-02T15-04-05")
	ext := filepath.Ext(fw.cfg.Path)
	base := strings.TrimSuffix(fw.cfg.Path, ext)
	backupPath := fmt.Sprintf("%s-%s%s", base, ts, ext)

	if err := os.Rename(fw.cfg.Path, backupPath); err != nil {
		return err
	}

	// Compress in background if enabled
	if fw.cfg.Compress {
		go compressFile(backupPath)
	}

	// Clean old backups
	go fw.cleanup()

	// Open new file
	return fw.openFile()
}

func (fw *FileWriter) cleanup() {
	ext := filepath.Ext(fw.cfg.Path)
	base := strings.TrimSuffix(fw.cfg.Path, ext)
	pattern := base + "-*" + ext + "*"

	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return
	}

	// Sort by modification time, oldest first
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var files []fileInfo
	now := time.Now()

	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path: m, modTime: info.ModTime()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove by age
	if fw.cfg.maxAge() > 0 {
		for _, f := range files {
			if now.Sub(f.modTime) > fw.cfg.maxAge() {
				os.Remove(f.path)
			}
		}
	}

	// Refresh after age cleanup
	matches, _ = filepath.Glob(pattern)
	if maxB := fw.cfg.maxBackups(); maxB > 0 && len(matches) > maxB {
		sort.Strings(matches)
		for _, m := range matches[:len(matches)-maxB] {
			os.Remove(m)
		}
	}
}

func compressFile(path string) {
	src, err := os.Open(path)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(path + ".gz")
	if err != nil {
		return
	}

	gz := gzip.NewWriter(dst)
	if _, err := io.Copy(gz, src); err != nil {
		gz.Close()
		dst.Close()
		os.Remove(path + ".gz")
		return
	}

	gz.Close()
	dst.Close()
	src.Close()
	os.Remove(path) // remove uncompressed original
}
