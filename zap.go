package gecko

import (
	"go.uber.org/zap"
)

var _BootstrapLogger, _ = zap.NewDevelopment()

func Zap() *zap.SugaredLogger {
	return _BootstrapLogger.Sugar()
}

