package messagequeue

type IKeyMessage interface {
	GetMessageData() ([]byte, error)
	GetKey() []byte // 发送时用来匹配分区的key
}

func NewKeyMessage(key []byte, message []byte) IKeyMessage {
	return &KeyMessage{
		key:     key,
		message: message,
	}
}

type KeyMessage struct {
	key     []byte
	message []byte
}

func (m *KeyMessage) GetMessageData() ([]byte, error) {
	return m.message, nil
}

func (m *KeyMessage) GetKey() []byte {
	return m.key
}
