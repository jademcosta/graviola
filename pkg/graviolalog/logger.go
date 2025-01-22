package graviolalog

import (
	"log/slog"
	"os"
	"strings"

	"github.com/jademcosta/graviola/pkg/config"
)

func NewLogger(conf config.LogConfig) *slog.Logger {

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(conf.Level),
	})

	return slog.New(logHandler)
}

func parseLevel(lvl string) slog.Level {

	switch strings.ToUpper(lvl) {
	case "ERROR":
		return slog.LevelError
	case "WARN":
		return slog.LevelWarn
	case "DEBUG":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
