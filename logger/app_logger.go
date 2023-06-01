package logger

import (
	"github.com/go-logr/logr"
)

var logger logr.Logger

func Init(l logr.Logger) {
	logger = l
}

func l() logr.Logger {
	return logger
}

// Debug Сообщение отладочного уровня
func Debug(msg string, keysAndValues ...interface{}) {
	l().V(1).Info(msg, keysAndValues)
}

// Info Сообщение информационного уровня
func Info(msg string, keysAndValues ...interface{}) {
	l().Info(msg, keysAndValues)
}

// Error Сообщение об ошибке
func Error(err error, msg string, keysAndValues ...interface{}) {
	l().Error(err, msg, keysAndValues)
}
