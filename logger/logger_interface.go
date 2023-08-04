package logger

// AppLoggerInterface AppLogger Интерфейс логгера
type AppLoggerInterface interface {
	// Debug Сообщение отладочного уровня
	Debug(msg string, keysAndValues ...interface{})
	// Info Сообщение информационного уровня
	Info(msg string, keysAndValues ...interface{})
	// Error Сообщение об ошибке
	Error(err error, msg string, keysAndValues ...interface{})
}
