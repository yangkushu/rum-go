package log

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/yangkushu/rum-go/iface"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"time"
)

type Logger struct {
	zapLogger *zap.Logger
	config    Config
}

var defaultEncoderTimeFormat = "2006-01-02 15:04:05.000000"

func iso8601TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(defaultEncoderTimeFormat))
}

/*
NewLogger
v 0.0.11 更新一下，使用了更底层的 API
*/
func NewLogger(config *Config) (iface.ILogger, error) {
	// 解析 config.Level 自定义级别
	var lvl zapcore.Level
	if config.Level != "" {
		var err error
		lvl, err = zapcore.ParseLevel(config.Level)
		if err != nil {
			return nil, err
		}
	} else {
		lvl = zap.InfoLevel
	}

	if len(config.TimeFormat) > 0 {
		defaultEncoderTimeFormat = config.TimeFormat
	}

	// 创建一个用于动态调整级别的原子级别对象
	atomicLevel := zap.NewAtomicLevelAt(lvl)

	//// 定义编码器配置
	//encoderCfg := zapcore.EncoderConfig{
	//	TimeKey:        "T",
	//	LevelKey:       "L",
	//	NameKey:        "N",
	//	CallerKey:      "C",
	//	MessageKey:     "M",
	//	StacktraceKey:  "S",
	//	LineEnding:     zapcore.DefaultLineEnding,
	//	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	//	EncodeTime:     iso8601TimeEncoder,
	//	EncodeDuration: zapcore.StringDurationEncoder,
	//	EncodeCaller:   zapcore.ShortCallerEncoder,
	//}

	encoderCfg := newEncodingConfig()

	if config.DisableCaller {
		encoderCfg.CallerKey = ""
	}

	// 根据配置定义编码器
	encoder := newEncoder(config.Encoding, encoderCfg)
	//var encoder zapcore.Encoder
	//if config.Encoding == "json" {
	//	encoder = zapcore.NewJSONEncoder(encoderCfg)
	//} else {
	//	encoder = zapcore.NewConsoleEncoder(encoderCfg)
	//}

	//errorOutputPaths := zapcore.AddSync(os.Stderr)
	//core := zapcore.NewCore(encoder, outputPaths, atomicLevel)

	// 设置日志输出
	var cores []zapcore.Core

	// 创建控制台的 WriteSyncer
	consoleSyncer := zapcore.AddSync(os.Stdout)
	cores = append(cores, zapcore.NewCore(encoder, consoleSyncer, atomicLevel))

	// 设置输出到文件
	if config.EnableWriteToFile {
		if len(config.LogFile) == 0 {
			return nil, fmt.Errorf("LogFile is empty")
		}

		fileEncoder := newEncoder(config.Encoding, encoderCfg)

		var logFile io.Writer

		if config.RollingFile != nil {
			// 配置日志切割
			logFile = &lumberjack.Logger{
				Filename:   config.LogFile,
				MaxSize:    config.RollingFile.MaxSize,
				MaxBackups: config.RollingFile.MaxBackups,
				MaxAge:     config.RollingFile.MaxAge,
				Compress:   config.RollingFile.Compress,
				LocalTime:  config.RollingFile.LocalTime,
			}
		} else {
			var err error
			logFile, err = os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
		}

		fileSyncer := zapcore.AddSync(logFile)
		cores = append(cores, zapcore.NewCore(fileEncoder, fileSyncer, atomicLevel))
	}

	// 设置 WriteSyncerChan
	var channelWriteSyncer *ChannelWriteSyncer
	if config.WriteSyncerChan != nil {
		writeSyncerEncoding := newEncoder(config.WriteSyncerEncoding, encoderCfg)
		channelWriteSyncer = newChannelWriteSyncer(config.WriteSyncerChan)
		level, err := newLevel(config.WriteSyncerLevel)
		if err != nil {
			return nil, errors.Wrap(err, "write syncer level parse error")
		}
		cores = append(cores, zapcore.NewCore(writeSyncerEncoding, channelWriteSyncer, level))
	}

	// 使用 zapcore.NewTee 来组合 WriteSyncer
	tee := zapcore.NewTee(cores...)
	// 使用组合的 Tee core 创建 logger
	logger := zap.New(tee)

	if config.Development {
		logger = logger.WithOptions(zap.Development())
	}
	if !config.DisableCaller {
		logger = logger.WithOptions(zap.AddCaller())
		logger = logger.WithOptions(zap.AddCallerSkip(2))
	}
	if !config.DisableStacktrace {
		logger = logger.WithOptions(zap.AddStacktrace(zapcore.WarnLevel)) // 根据需要调整级别
	}
	return &Logger{
		zapLogger: logger,
	}, nil
}

func NewDefaultConfig() *Config {
	// Set default values for LoggerConfig
	// Call NewLogger method to create a logger with default configuration
	return &Config{
		Level:             "info",
		Development:       false,
		DisableCaller:     false,
		CallerSkip:        2,
		DisableStacktrace: true,
		//SamplingInitial:    100,
		//SamplingThereafter: 100,
		Encoding: "",
	}
}

func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

func (l *Logger) Info(msg string, fields ...iface.Field) {
	zapFields := toZapFields(fields)
	l.zapLogger.Info(msg, zapFields...)
}

func (l *Logger) Warn(msg string, fields ...iface.Field) {
	zapFields := toZapFields(fields)
	l.zapLogger.Warn(msg, zapFields...)
}

func (l *Logger) Error(msg string, fields ...iface.Field) {
	zapFields := toZapFields(fields)
	l.zapLogger.Error(msg, zapFields...)
}

func (l *Logger) Debug(msg string, fields ...iface.Field) {
	zapFields := toZapFields(fields)
	l.zapLogger.Debug(msg, zapFields...)
}

func newEncoder(encoding string, encodingConfig zapcore.EncoderConfig) zapcore.Encoder {
	if encoding == "json" {
		return zapcore.NewJSONEncoder(encodingConfig)
	} else {
		return zapcore.NewConsoleEncoder(encodingConfig)
	}
}

func newEncodingConfig() zapcore.EncoderConfig {
	// 定义编码器配置
	return zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     iso8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func newLevel(level string) (zapcore.Level, error) {
	return zapcore.ParseLevel(level)
}

func (l *Logger) GetLevel() string {
	return l.zapLogger.Level().String()
}
