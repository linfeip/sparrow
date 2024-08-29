package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log = zap.New(
	zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		os.Stdout,
		zap.DebugLevel,
	),
	zap.WithCaller(true),
).Sugar()

func GetLogger() *zap.SugaredLogger {
	return log
}

func SetLogger(l *zap.SugaredLogger) {
	log = l
}

func Log(lvl zapcore.Level, args ...interface{}) {
	log.Log(lvl, args...)
}

func Logf(lvl zapcore.Level, tpl string, args ...interface{}) {
	log.Logf(lvl, tpl, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(tp string, args ...interface{}) {
	log.Infof(tp, args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Warnf(tp string, args ...interface{}) {
	log.Warnf(tp, args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(tp string, args ...interface{}) {
	log.Errorf(tp, args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Fatalf(tp string, args ...interface{}) {
	log.Fatalf(tp, args...)
}

func Panic(args ...interface{}) {
	log.Panic(args...)
}

func Panicf(tp string, args ...interface{}) {
	log.Panicf(tp, args...)
}
