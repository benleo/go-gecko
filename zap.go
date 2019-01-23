package gecko

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ZapConfig = zap.Config{
	Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
	Development: true,
	Encoding:    "console",
	EncoderConfig: zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	},
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
}

func ZapConfig() zap.Config {
	return _ZapConfig
}

func ZapLogger() *zap.Logger {
	logger, _ := _ZapConfig.Build()
	return logger
}

func Zap() *zap.SugaredLogger {
	return ZapLogger().Sugar()
}

func ZapDebug(msg string) {
	Zapf(func(zap *zap.Logger) {
		zap.Debug(msg)
	})
}

func Zapf(f func(zap *zap.Logger)) {
	logger := ZapLogger()
	defer logger.Sync()
	f(logger)
}
