package messagequeue

import (
	"encoding/json"
	"github.com/yangkushu/rum-go/iface"
	"github.com/yangkushu/rum-go/log"
	"reflect"
)

// StructMessageSubscriber 同步结构体订阅者。使用泛型，处理指定结构体或者字符串数据
type StructMessageSubscriber[M any] struct {
	onMessage func(M) bool
	log       iface.ILogger
}

// NewStructMessageSubscriber 创建新的StructMessageSubscriber的函数
func NewStructMessageSubscriber[M any](onMessage func(M) bool, log iface.ILogger) *StructMessageSubscriber[M] {
	return &StructMessageSubscriber[M]{
		onMessage: onMessage,
		log:       log,
	}
}

// HandleMessage 处理消息的函数
func (a *StructMessageSubscriber[M]) HandleMessage(message IMessage) bool {
	// 通过反射检查M的类型是否为字符串
	msgReflectType := reflect.TypeOf((*M)(nil)).Elem()
	if msgReflectType.Kind() == reflect.String {
		// 如果M是字符串，直接设置消息
		msg := reflect.New(msgReflectType).Elem()
		msg.SetString(string(message.GetMessageData()))
		return a.onMessage(msg.Interface().(M))
	} else {
		// 如果M不是字符串，将消息反序列化到msg中
		var msg M
		err := json.Unmarshal(message.GetMessageData(), &msg)
		if err != nil {
			a.log.Error("StructMessageSubscriber HandleMessage error", log.String("error", err.Error()))
			return false
		}
		return a.onMessage(msg)
	}
}

// HandleError 处理错误的函数
func (a *StructMessageSubscriber[M]) HandleError(message IMessage, err error) {
	a.log.Error("StructMessageSubscriber error", log.String("error", err.Error()))
}
