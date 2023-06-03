package logger

import (
	"github.com/go-logr/logr"
)

type AppLogger struct {
	logger logr.Logger
}

func (al *AppLogger) Init(l logr.Logger) {
	al.logger = l
}

// Debug Сообщение отладочного уровня
func (al *AppLogger) Debug(msg string, keysAndValues ...interface{}) {
	al.logger.V(1).Info(msg, keysAndValues)
}

// Info Сообщение информационного уровня
func (al *AppLogger) Info(msg string, keysAndValues ...interface{}) {
	al.logger.Info(msg, keysAndValues)
}

// Error Сообщение об ошибке
func (al *AppLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	al.logger.Error(err, msg, keysAndValues)
}
