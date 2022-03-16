package log

import (
	"io"

	"go.uber.org/zap/zapcore"
)

const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
)

type Logger interface {
	baseLogger
	sugaredLogger
	toWriter
}

type PrefixLogger interface {
	Logger
	Prefix() string
	WithPrefix(prefix string) PrefixLogger
}

type loggerWithZapCore interface {
	Logger
	zapCore() zapcore.Core
}

// DebugLogger returns logs as string in tests.
type DebugLogger interface {
	Logger
	ConnectTo(writer io.Writer)
	Truncate()
	AllMessages() string
	DebugMessages() string
	InfoMessages() string
	WarnMessages() string
	WarnAndErrorMessages() string
	ErrorMessages() string
}

type baseLogger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Sync() error
}

type sugaredLogger interface {
	With(args ...interface{}) Logger // creates a child logger and adds structured context to it.
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

type toWriter interface {
	DebugWriter() *LevelWriter
	InfoWriter() *LevelWriter
	WarnWriter() *LevelWriter
	ErrorWriter() *LevelWriter
}
