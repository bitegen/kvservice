package logger

import (
	"log/slog"
	"os"
	"time"
)

const (
	EnvLocal string = "local"
	EnvProd  string = "prod"
)

func NewLogger(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case EnvProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: replaceTimeRFC3339,
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: replaceTimeHuman,
		})
	}

	return slog.New(handler)
}

func replaceTimeRFC3339(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		if t, ok := a.Value.Any().(time.Time); ok {
			return slog.Time(slog.TimeKey, t.UTC())
		}
	}
	return a
}

func replaceTimeHuman(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		if t, ok := a.Value.Any().(time.Time); ok {
			return slog.String(slog.TimeKey, t.Format("15:04:05.000"))
		}
	}
	return a
}
