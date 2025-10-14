package log

var logger ILogger

func SetLogger(l ILogger) {
	logger = l
}

func GetLogger() ILogger {
	return logger
}

func Info(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Error(msg, fields...)
}

func Debug(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Debug(msg, fields...)
}
