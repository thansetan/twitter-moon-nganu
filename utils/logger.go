package utils

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))
	return logger
}
