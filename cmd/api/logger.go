package main

import (
	"log/slog"
	"os"
	"strings"
)

func InitLogger() {
	// Determine log level from Environment Variables (default to INFO)
	level := slog.LevelInfo
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
