package graviolalog

import (
	"fmt"
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

type GraviolaLogger struct {
	logg *slog.Logger
}

func AdaptToGoKitLogger(logg *slog.Logger) *GraviolaLogger {
	return &GraviolaLogger{
		logg: logg,
	}
}

func (gravLogger *GraviolaLogger) Log(args ...interface{}) error {
	if len(args) == 0 {
		return nil
	}

	if len(args) == 1 {
		gravLogger.logg.Info(fmt.Sprintf("%v", args[0]))
		return nil
	}

	gravLogger.logg.Info(fmt.Sprintf("%v", args[0]), args[1:]...)
	return nil
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
