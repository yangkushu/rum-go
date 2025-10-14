// kafka.go

package messagequeue

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	kafkasasl "github.com/segmentio/kafka-go/sasl/plain"
	"github.com/yangkushu/rum-go/iface"
	"github.com/yangkushu/rum-go/log"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Kafka struct {
	brokers          []string
	user             string
	password         string
	caFile           string
	saslMechanisms   string
	securityProtocol string
	//group            string
	writer     *kafka.Writer
	readers    []*kafka.Reader
	dialer     *kafka.Dialer
	config     *KafkaConfig
	readerLock sync.Mutex
	log        iface.ILogger
}

// NewKafka creates a new Kafka client.
func NewKafka(config *KafkaConfig) (IMessageQueue, error) {

	var brokers []string
	brokersStr := config.Brokers
	if "" == brokersStr {
		return nil, errors.New("kafka config 'brokers' is empty")
	} else {
		brokers = strings.Split(brokersStr, ",")
	}

	var user string
	username := config.Username
	if "" == username {
		return nil, errors.New("kafka config 'username' is empty")
	} else {
		user = username
	}

	if "" == config.Password {
		return nil, errors.New("kafka config 'password' is empty")
	}

	saslMechanisms := config.Mechanisms
	if "" == saslMechanisms {
		return nil, errors.New("kafka config 'mechanisms' is empty")
	}

	securityProtocol := config.Protocol
	if "" == securityProtocol {
		return nil, errors.New("kafka config 'protocol' is empty")
	}

	var tlsConfig *tls.Config

	// 用证书登陆
	caFile := config.CaFile
	if "" != caFile {
		// Load CA certificate
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read kafka CA certificate:%w", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// TLS Config
		tlsConfig = &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: true,
		}
	}

	// SASL Config
	mechanism := kafkasasl.Mechanism{
		Username: user,
		Password: config.Password,
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		TLS:           tlsConfig,
		SASLMechanism: mechanism,
	}

	k := &Kafka{
		brokers:          brokers,
		user:             user,
		password:         config.Password,
		caFile:           caFile,
		saslMechanisms:   saslMechanisms,
		securityProtocol: securityProtocol,
		//group:            "operation-server",
		//writer: writer,
		dialer: dialer,
		config: config,
		log:    config.Logger,
	}

	if err := k.checkConnection(); err != nil {
		return k, fmt.Errorf("failed to connect to Kafka:%w", err)
	}

	// Initialize Kafka Writer (Producer)
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Transport:              &kafka.Transport{TLS: tlsConfig, SASL: mechanism},
		BatchSize:              1,    // For immediate delivery
		AllowAutoTopicCreation: true, // Create topic if it doesn't exist
		Logger:                 NewKafkaLogger(config.Logger),
		ErrorLogger:            NewKafkaErrorLogger(config.Logger),
	}

	k.writer = writer

	return k, nil
}

func (k *Kafka) Close() error {
	if k.writer != nil {
		k.writer.Close()
	}
	if len(k.readers) > 0 {
		for _, reader := range k.readers {
			reader.Close()
		}
	}

	return nil
}

// Publish sends a message to the specified topic.
// message 会根据类型做不同的处理
func (k *Kafka) Publish(topic Topic, message interface{}) error {
	if k.writer == nil {
		return errors.New("kafka writer (producer) is not initialized")
	}

	var value []byte
	var key []byte

	// Transform message to appropriate format
	switch msg := message.(type) {
	case string:
		value = []byte(msg)
	case []byte:
		value = msg
	case IKeyMessage:
		var err error
		value, err = msg.GetMessageData()
		if err != nil {
			return fmt.Errorf("failed to get message data:%w", err)
		}
		key = msg.GetKey()
	default:
		encoded, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message to JSON:%w", err)
		}
		// Recursively call Publish with the JSON string
		return k.Publish(topic, string(encoded))
	}

	if value == nil {
		k.log.Error("message is nil", log.String("topic", string(topic)))
		return errors.New("message is nil")
	}

	// Prepare the message for Kafka
	kafkaMsg := kafka.Message{
		Topic: string(topic),
		Value: value,
	}

	if key != nil && len(key) > 0 {
		kafkaMsg.Key = key
	}

	// Send the message to Kafka
	err := k.writer.WriteMessages(context.Background(), kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka:%w", err)
	}

	return nil
}

// Subscribe subscribes to the specified topic and listens for messages.
func (k *Kafka) Subscribe(topic Topic, groupId string, handler IMessageSubscriber) error {
	// Initialize Kafka Reader (Consumer)
	readerConfig := kafka.ReaderConfig{
		Brokers:                k.brokers,
		Topic:                  string(topic),
		GroupID:                groupId,
		MinBytes:               10e3, // 10KB
		MaxBytes:               10e6, // 10MB
		Dialer:                 k.dialer,
		WatchPartitionChanges:  true, // Watch for partition changes
		PartitionWatchInterval: time.Minute,
		Logger:                 NewKafkaLogger(k.config.Logger),
		ErrorLogger:            NewKafkaErrorLogger(k.config.Logger),
	}

	reader := kafka.NewReader(readerConfig)

	k.readerLock.Lock()
	defer k.readerLock.Unlock()
	k.readers = append(k.readers, reader)

	go func() {
		for {
			// Read message from Kafka
			msg, err := reader.FetchMessage(context.Background())
			if err != nil {
				if k.config.IsDebug {
					k.log.Debug("Error while reading Kafka message", log.String("error", err.Error()))
				}
				if err == io.EOF {
					k.log.Error("EOF while reading Kafka message", log.String("topic", string(topic)))
					return
				}
				handler.HandleError(nil, fmt.Errorf("error while reading Kafka message:%w", err))
				continue
			}

			// Send message to callback channel
			message := &KafKaMessage{originalMsg: msg}
			if k.config.IsDebug {
				k.log.Debug("Received message from Kafka", log.String("topic", msg.Topic), log.Any("message", message), log.Any("msg", msg), log.String("data", string(message.GetMessageData())))
			}
			if !handler.HandleMessage(message) {
				// 返回false不提交消息
				continue
			}

			// Commit the message after processing
			if err := reader.CommitMessages(context.Background(), msg); err != nil {
				handler.HandleError(message, fmt.Errorf("failed to commit message:%w", err))
				if k.config.IsDebug {
					k.log.Debug("Failed to commit message", log.String("error", err.Error()))
				}
				continue
			}
		}
	}()

	return nil
}

func (k *Kafka) checkConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for _, broker := range k.brokers {
		conn, err := k.dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			return err
		}
		// 简单获取一些 metadata 来验证连接
		_, err = conn.Brokers()
		conn.Close()
		return err
	}
	return nil
}
