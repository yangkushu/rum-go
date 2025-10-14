package messagequeue

import "io"

type IMessageQueue interface {
	Publish(topic Topic, message interface{}) error
	Subscribe(topic Topic, groupId string, handler IMessageSubscriber) error
	io.Closer
}
