package iface

type ILogger interface {
	Sync() error
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Debug(msg string, fields ...Field)
	GetLevel() string
}

type Field interface{}
