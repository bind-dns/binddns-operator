package zlog

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *zap.SugaredLogger
)

func syslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05"))
}

func DefaultLog(logName string) {
	core := zapCore(logName, zapcore.InfoLevel, 500, 30, 15, true)
	logger = zap.New(core, zap.AddCaller(), zap.Development(), zap.AddCallerSkip(1)).Sugar()
}

func InitLog(logName string, maxSize, maxBackups, maxAge int, compress bool) {
	core := zapCore(logName, zapcore.InfoLevel, maxSize, maxBackups, maxAge, compress)
	logger = zap.New(core, zap.AddCaller(), zap.Development(), zap.AddCallerSkip(1)).Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return logger
}

func zapCore(logFile string, logLevel zapcore.Level, maxSize int, maxBackups int, maxAge int, compress bool) zapcore.Core {
	// set log level
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(logLevel)

	// set encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     syslogTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	return zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   compress,
		})),
		atomicLevel,
	)
}

func Debug(args ...interface{}) () {
	logger.Debug(args)
}

func Debugf(template string, args ...interface{}) () {
	logger.Debugf(template, args...)
}

func Info(args ...interface{}) {
	logger.Info(args)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	logger.Error(args)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func DPanic(args ...interface{}) {
	logger.DPanic(args)
}

func DPanicf(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args)
}

func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args)
}

func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}
