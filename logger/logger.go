package logger

import "go.uber.org/zap"

var logger *zap.Logger

func init() {
	log, err := zap.NewProduction()
	defer log.Sync()
	logger = log
	if err != nil {
		panic(err)
	}
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	logger.Panic(msg, fields...)
}
