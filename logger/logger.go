package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

// InitLogger initializes the logger with the given log level.
func InitLogger(level string) {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stdout, parseLevel(strings.ToLower(level)), true)))
}

func Debug(msg string, v ...interface{}) {
	log.Debug(msg, v...)
}

func Info(msg string, v ...interface{}) {
	log.Info(msg, v...)
}

func Warn(msg string, v ...interface{}) {
	log.Warn(msg, v...)
}

func Error(msg string, v ...interface{}) {
	log.Error(msg, v...)
}

type Instance struct{}

func (l *Instance) Debug(args ...interface{}) {
	Debug(fmt.Sprint(args...))
}

func (l *Instance) Debugf(msg string, args ...interface{}) {
	Debug(fmt.Sprintf(msg, args...))
}

func (l *Instance) Info(args ...interface{}) {
	Info(fmt.Sprint(args...))
}

func (l *Instance) Infof(msg string, args ...interface{}) {
	Info(fmt.Sprintf(msg, args...))
}

func (l *Instance) Warn(args ...interface{}) {
	Warn(fmt.Sprint(args...))
}

func (l *Instance) Warnf(msg string, args ...interface{}) {
	Warn(fmt.Sprintf(msg, args...))
}

func (l *Instance) Error(args ...interface{}) {
	Error(fmt.Sprint(args...))
}

func (l *Instance) Errorf(msg string, args ...interface{}) {
	Error(fmt.Sprintf(msg, args...))
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return log.LevelDebug
	case "info":
		return log.LevelInfo
	case "warn":
		return log.LevelWarn
	case "error":
		return log.LevelError
	default:
		return log.LevelInfo
	}
}
