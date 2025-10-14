package messagequeue

import (
	"fmt"
	"github.com/yangkushu/rum-go/iface"
)

type KafkaLogger struct {
	logger iface.ILogger
}

func NewKafkaLogger(logger iface.ILogger) *KafkaLogger {
	return &KafkaLogger{logger: logger}
}

func (l *KafkaLogger) Printf(format string, v ...interface{}) {
	if l.logger == nil {
		return
	}
	str := ""
	if len(v) == 0 {
		str = format
	} else {
		str = fmt.Sprintf(format, v...)
	}
	//str = ReplaceNewLine(str)
	l.logger.Info(str)
}

type KafkaErrorLogger struct {
	logger iface.ILogger
}

func NewKafkaErrorLogger(logger iface.ILogger) *KafkaErrorLogger {
	return &KafkaErrorLogger{logger: logger}
}

func (l *KafkaErrorLogger) Printf(format string, v ...interface{}) {
	if l.logger == nil {
		return
	}
	str := ""
	if len(v) == 0 {
		str = format
	} else {
		str = fmt.Sprintf(format, v...)
	}
	//str = ReplaceNewLine(str)
	l.logger.Error(str)
}

// 先去掉
//// 替换掉日志内容中的换行符
//func ReplaceNewLine(s string) string {
//	return strings.ReplaceAll(s, "\n", ">>")
//}
