package graviolalog

import (
	"github.com/jademcosta/graviola/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.SugaredLogger
}

func NewLogger(conf config.LogConfig) *Logger {
	logLevel, err := zapcore.ParseLevel(conf.Level)
	if err != nil {
		panic("error parsing log level: " + err.Error())
	}
	zap.NewProductionConfig()

	zapconfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(logLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    encoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := zapconfig.Build()
	if err != nil {
		panic("Error initializing logger: " + err.Error())
	}

	return &Logger{logger: logger.Sugar()}
}

func (l *Logger) Log(keyvals ...interface{}) error {
	l.logger.Infow("", keyvals...)
	return nil
}

func (l *Logger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debugw(msg, keyvals...)
}

func (l *Logger) Info(msg string, keyvals ...interface{}) {
	l.logger.Infow(msg, keyvals...)
}

func (l *Logger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warnw(msg, keyvals...)
}

func (l *Logger) Error(msg string, keyvals ...interface{}) {
	l.logger.Errorw(msg, keyvals...)
}

func (l *Logger) Panic(msg string, keyvals ...interface{}) {
	l.logger.Panicw(msg, keyvals...)
}

func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
