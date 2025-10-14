package messagequeue

import (
	"fmt"
	"github.com/yangkushu/rum-go/log"
	"os"
	"testing"
	"time"
)

func TestKafkaPlainText(t *testing.T) {

	logger, err := log.NewLogger(log.NewDefaultConfig())
	if err != nil {
		panic(err)
	}
	log.SetLogger(logger)

	c := &KafkaConfig{
		Brokers:    os.Getenv("broker"),
		Protocol:   os.Getenv("protocol"),
		Mechanisms: os.Getenv("mechanisms"),
		Username:   os.Getenv("username"),
		Password:   os.Getenv("password"),
	}

	topic := os.Getenv("topic")
	group := "test-group"
	kafka, err := NewKafka(c)
	if err != nil {
		panic(err)
	}

	err = kafka.Subscribe(Topic(topic), group, &exampleHandler{})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		err = kafka.Publish(Topic(topic), fmt.Sprintf("hello world : %d", i))
		if err != nil {
			panic(err)
		}
	}

	time.Sleep(100 * time.Second)
}

type exampleHandler struct{}

func (e exampleHandler) HandleMessage(message IMessage) (commit bool) {
	fmt.Printf("message: %s\n", message.GetMessageData())
	return true
}

func (e exampleHandler) HandleError(message IMessage, err error) {
	//TODO implement me
	panic("implement me")
}
