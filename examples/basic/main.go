package main

import (
	"context"

	"github.com/Bhavyyadav25/loghq"
)

func main() {
	// The default logger writes beautiful colored output to stderr
	loghq.Info("application started", "version", "1.0.0")
	loghq.Debug("loading config", "path", "/etc/app/config.yaml")
	loghq.Success("database connected", "host", "localhost", "port", 5432)
	loghq.Warn("cache miss rate high", "rate", 0.45)
	loghq.Error("request failed", "status", 500, "path", "/api/users")

	// Structured fields
	loghq.WithFields(loghq.Fields{
		"service": "auth",
		"version": "2.1.0",
	}).Info("service initialized")

	// Typed fields
	loghq.With(
		loghq.String("method", "POST"),
		loghq.Int("status", 201),
	).Info("request completed")

	// Context with request ID
	ctx := loghq.ContextWithFields(
		context.Background(),
		loghq.String("request_id", "req-abc-123"),
	)
	loghq.WithContext(ctx).Info("processing order", "order_id", 42)

	// Instance logger with custom level
	logger := loghq.New(
		loghq.WithHandler(loghq.NewConsoleHandler()),
		loghq.WithLevel(loghq.WarnLevel),
	)
	logger.Info("this will not appear") // filtered out
	logger.Warn("disk space low", "free_gb", 2)

	loghq.Trace("detailed trace info")
}
