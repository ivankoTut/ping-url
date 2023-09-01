package logger

import (
	baseLog "log"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func NewLogger(env string, logFile string) *slog.Logger {
	var log *slog.Logger

	opts := &slog.HandlerOptions{}

	switch env {
	case envLocal:
		opts.Level = slog.LevelDebug
		h := slog.NewTextHandler(os.Stdout, opts)
		log = slog.New(h)
	case envDev:
		opts.Level = slog.LevelInfo
		opts.AddSource = true
		h := slog.NewJSONHandler(os.Stdout, opts)
		log = slog.New(h)
	case envProd:
		opts.Level = slog.LevelWarn
		opts.AddSource = true
		file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			baseLog.Fatalf("error opening file: %v", err)
		}

		h := slog.NewJSONHandler(file, opts)
		log = slog.New(h)
	}

	return log
}
