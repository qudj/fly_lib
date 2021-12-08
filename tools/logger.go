package tools

import (
	"context"
	"fmt"
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

func LogCtxInfo(ctx context.Context, template string, args ...interface{})  {
	traceId, _ := ctx.Value("trace_id").(string)
	template = fmt.Sprintf("trace_id: %s  %s", traceId, template)
	logger.Infof(template, args)
}

func LogCtxWarn(ctx context.Context, template string, args ...interface{})  {
	traceId, _ := ctx.Value("trace_id").(string)
	template = fmt.Sprintf("trace_id: %s  %s", traceId, template)
	logger.Warnf(template, args)
}

func LogCtxError(ctx context.Context, template string, args ...interface{})  {
	traceId, _ := ctx.Value("trace_id").(string)
	template = fmt.Sprintf("trace_id: %s  %s", traceId, template)
	logger.Errorf(template, args)
}