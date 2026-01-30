package main

import (
	"log"
	"time"

	"github.com/Bhavyyadav25/loghq"
)

func main() {
	// File writer with rotation
	fw, err := loghq.NewFileWriter(loghq.FileConfig{
		Path:       "/tmp/loghq-example/app.log",
		MaxSize:    10 * 1024 * 1024, // 10MB
		MaxAge:     24 * time.Hour,
		MaxBackups: 3,
		Compress:   true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	// JSON to file, console to stderr
	logger := loghq.New(
		loghq.WithHandler(loghq.NewMultiHandler(
			loghq.NewConsoleHandler(),
			loghq.NewJSONHandler(fw),
		)),
	)

	logger.Info("logging to file and console", "file", "/tmp/loghq-example/app.log")
	logger.Success("file rotation configured", "max_size", "10MB", "max_backups", 3)
}
