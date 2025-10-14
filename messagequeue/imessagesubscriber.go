package messagequeue

// IMessageSubscriber 接口现在接受一个类型参数，它指定了消息的类型。
type IMessageSubscriber interface {
	HandleMessage(message IMessage) (commit bool)
	HandleError(message IMessage, err error)
	//start()
	//Stop()
}
