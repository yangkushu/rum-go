package messagequeue

import (
	kafka "github.com/segmentio/kafka-go"
)

type KafKaMessage struct {
	originalMsg kafka.Message
}

func NewKafkaMessage(message, topic, key string) IMessage {
	return &KafKaMessage{
		originalMsg: kafka.Message{
			Value: []byte(message),
			Topic: topic,
			Key:   []byte(key),
		},
	}
}

func (m *KafKaMessage) GetMessageData() []byte {
	return m.originalMsg.Value
}

func (m *KafKaMessage) GetTopic() string {
	return m.originalMsg.Topic
}

func (m *KafKaMessage) GetKey() []byte {
	return m.originalMsg.Key
}
