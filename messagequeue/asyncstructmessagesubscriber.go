package messagequeue

import (
	"encoding/json"
	"github.com/yangkushu/rum-go/log"
	"reflect"
	"sync"
)

// AsyncStructMessageSubscriber 异步结构体订阅者。使用泛型，处理指定结构体或者字符串数据
type AsyncStructMessageSubscriber[M any] struct {
	messageChannel chan M
	stopChannel    chan bool
	onMessage      func(M)
	onceStart      sync.Once
	goroutineNum   int
}

// NewAsyncStructMessageSubscriber 创建新的AsyncStructMessageSubscriber的函数
func NewAsyncStructMessageSubscriber[M any](onMessage func(M), goroutineNum int) *AsyncStructMessageSubscriber[M] {
	a := &AsyncStructMessageSubscriber[M]{
		messageChannel: make(chan M),
		stopChannel:    make(chan bool),
		onMessage:      onMessage,
		goroutineNum:   goroutineNum,
	}
	a.start()
	return a
}

// HandleMessage 处理消息的函数
func (a *AsyncStructMessageSubscriber[M]) HandleMessage(message IMessage) bool {
	// 通过反射检查M的类型是否为字符串
	msgReflectType := reflect.TypeOf((*M)(nil)).Elem()
	if msgReflectType.Kind() == reflect.String {
		// 如果M是字符串，直接设置消息
		msg := reflect.New(msgReflectType).Elem()
		msg.SetString(string(message.GetMessageData()))
		a.messageChannel <- msg.Interface().(M)
	} else {
		// 如果M不是字符串，将消息反序列化到msg中
		var msg M
		err := json.Unmarshal(message.GetMessageData(), &msg)
		if err != nil {
			a.HandleError(message, err)
			return true
		}
		a.messageChannel <- msg
	}
	return true
}

// start 启动消息处理的函数
func (a *AsyncStructMessageSubscriber[M]) start() {
	a.onceStart.Do(func() {
		// 启动n个goroutine来处理消息
		for i := 0; i < a.goroutineNum; i++ {
			go func() {
				for {
					select {
					case msg := <-a.messageChannel:
						a.onMessage(msg)
					case <-a.stopChannel:
						return
					}
				}
			}()
		}
	})
}

// stop 是停止消息处理的函数
func (a *AsyncStructMessageSubscriber[M]) stop() {
	a.stopChannel <- true
	close(a.messageChannel)
	close(a.stopChannel)
}

// HandleError 是处理错误的函数
func (a *AsyncStructMessageSubscriber[M]) HandleError(message IMessage, err error) {
	log.Error("StructMessageSubscriber error", log.String("error", err.Error()))
}
