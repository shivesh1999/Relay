package logger

import (
	"log/slog"
	"os"

	"github.com/relay/backend/internal/config"
)

type Logger struct {
	*slog.Logger
}

func New(cfg *config.Config) *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel(cfg.App.Debug),
	})

	return &Logger{slog.New(handler)}
}

func logLevel(debug bool) slog.Level {
	if debug {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
