package logger

import (
	"context"
	"log/slog"
	"os"
)

var L *slog.Logger

func init() {
	// Default to text handler for local development, could be JSON for production
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	L = slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(L)
}

func Infof(format string, args ...any) {
	L.Info(format, args...)
}

func Warnf(format string, args ...any) {
	L.Warn(format, args...)
}

func Errorf(format string, args ...any) {
	L.Error(format, args...)
}

func With(args ...any) *slog.Logger {
	return L.With(args...)
}

func FromContext(ctx context.Context) *slog.Logger {
	// Pluggable if we want to add request IDs or other context-bound info later
	return L
}
