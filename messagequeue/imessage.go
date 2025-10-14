package messagequeue

type IMessage interface {
	GetMessageData() []byte
	GetTopic() string
}
