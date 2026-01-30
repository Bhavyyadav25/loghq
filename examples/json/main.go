package main

import (
	"github.com/Bhavyyadav25/loghq"
)

func main() {
	// JSON logger for production
	logger := loghq.New(
		loghq.WithHandler(loghq.NewJSONHandler(loghq.Stdout)),
		loghq.WithLevel(loghq.InfoLevel),
	)

	logger.Info("server started", "port", 8080)
	logger.Info("request handled",
		"method", "GET",
		"path", "/api/users",
		"status", 200,
		"bytes", 1024,
		"elapsed", "12ms",
	)
	logger.Error("database error", "query", "SELECT *", "err", "connection refused")

	// Multi-handler: console + JSON
	multi := loghq.NewMultiHandler(
		loghq.NewConsoleHandler(),
		loghq.NewJSONHandler(loghq.Stdout),
	)
	ml := loghq.New(loghq.WithHandler(multi))
	ml.Success("both outputs", "target", "console+json")
}
