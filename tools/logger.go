package tools

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init()  {
	defaultLogger, _ := zap.NewProduction()
	logger = defaultLogger.Sugar()
}

func InitLogger(debug bool)  {
	var defaultLogger *zap.Logger
	if debug {
		defaultLogger, _ = zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel))
	} else {
		defaultLogger, _ = zap.NewProduction()
	}
	logger = defaultLogger.Sugar()
}

func Logger() *zap.SugaredLogger {
	return logger
}