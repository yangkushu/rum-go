// kafka_integration_test.go - Kafka集成测试
package messagequeue

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

var kafkaConfig = &KafkaConfig{
	Brokers:    "127.0.0.1:9092", // 确保这是您的Kafka服务器地址
	Username:   "develop",
	Password:   "PtP6EfkrQGqM1G44!@",
	Mechanisms: "PLAIN",
	Protocol:   "SASL_PLAINTEXT",
}

const topic = "integration-test-topic"

// 测试Kafka客户端初始化
func TestKafkaClientInitialization(t *testing.T) {
	client, err := NewKafka(kafkaConfig)
	if err != nil {
		t.Fatalf("Failed to create Kafka client: %s", err)
	}
	defer client.Close()
}

// 测试发布消息
func TestPublishMessage(t *testing.T) {
	client, err := NewKafka(kafkaConfig)
	if err != nil {
		t.Fatalf("Failed to create Kafka client: %s", err)
	}
	defer client.Close()

	message := "Hello Kafka!"

	if err := client.Publish(topic, message); err != nil {
		t.Fatalf("Failed to publish message: %s", err)
	}
}

// 测试订阅消息
func TestSubscribeMessage(t *testing.T) {
	client, err := NewKafka(kafkaConfig)
	if err != nil {
		t.Fatalf("Failed to create Kafka client: %s", err)
	}
	defer client.Close()

	topic := Topic(topic)
	groupId := "test-group"

	// 设置一个简单的handler
	handler := &simpleHandler{}
	if err := client.Subscribe(topic, groupId, handler); err != nil {
		t.Fatalf("Failed to subscribe to topic: %s", err)
	}

	// 发布消息以供消费
	message := "Test message"
	if err := client.Publish(topic, message); err != nil {
		t.Fatalf("Failed to publish message: %s", err)
	}

	// 给消息一些时间来被处理
	time.Sleep(10 * time.Second)

	// 检查消息是否已被正确处理
	if handler.lastMessage != message {
		t.Errorf("Expected message %s; got %s", message, handler.lastMessage)
	}
}

type simpleHandler struct {
	lastMessage string
}

func (h *simpleHandler) HandleMessage(message IMessage) (commit bool) {
	h.lastMessage = string(message.GetMessageData())
	return true
}

func (h *simpleHandler) HandleError(message IMessage, err error) {
	panic("implement me")
}

// 吞吐量测试
func TestKafkaThroughput(t *testing.T) {
	client, _ := NewKafka(kafkaConfig)
	defer client.Close()

	var wg sync.WaitGroup
	messageCount := 10000
	producerCount := 50
	message := strings.Repeat("A", 1024) // 1KB 消息

	start := time.Now()
	for i := 0; i < producerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < messageCount; j++ {
				client.Publish(topic, message)
			}
		}()
	}
	wg.Wait()
	duration := time.Since(start)
	totalMessages := producerCount * messageCount
	fmt.Printf("Total messages: %d, Duration: %v, Messages per second: %f\n", totalMessages, duration, float64(totalMessages)/duration.Seconds())
}

type latencyHandler struct {
	receivedTimes []time.Time
	mu            sync.Mutex
	wg            *sync.WaitGroup
	t             *testing.T
}

func (h *latencyHandler) HandleMessage(message IMessage) (commit bool) {
	h.mu.Lock()
	h.receivedTimes = append(h.receivedTimes, time.Now())
	h.mu.Unlock()
	h.t.Logf("Received message: %s", message.GetMessageData())
	h.wg.Done() // 确认处理完成
	return true
}

func (h *latencyHandler) HandleError(message IMessage, err error) {
	panic("implement	me")
}

func TestMessageLatency(t *testing.T) {
	client, err := NewKafka(kafkaConfig)
	if err != nil {
		t.Fatalf("Failed to create Kafka client: %s", err)
	}
	defer client.Close()

	const topic = "integration-test-latency3" // 使用不同的主题以避免与其他测试干扰

	var wg sync.WaitGroup
	messageCount := 100
	handler := &latencyHandler{wg: &wg, t: t}
	client.Subscribe(topic, "test-group-latency", handler)

	wg.Add(messageCount) // 在发送任何消息之前设置正确的count
	startTimes := make([]time.Time, messageCount)

	for i := 0; i < messageCount; i++ {
		startTimes[i] = time.Now()
		if err := client.Publish(topic, fmt.Sprintf("Test latency message %d", i)); err != nil {
			t.Errorf("Failed to publish message %d: %s", i, err)
			wg.Done() // 发布失败时也要调用Done，以避免阻塞
		}
	}

	wg.Wait() // 等待所有消息都被处理

	// 计算延迟
	var totalLatency time.Duration
	handler.mu.Lock()
	for i, receivedTime := range handler.receivedTimes {
		if i < len(startTimes) {
			latency := receivedTime.Sub(startTimes[i])
			totalLatency += latency
			fmt.Printf("Message %d latency: %v\n", i, latency)
		}
	}
	handler.mu.Unlock()

	avgLatency := totalLatency / time.Duration(len(handler.receivedTimes))
	fmt.Printf("Average message latency: %v\n", avgLatency)
}

// 测试Kafka服务不可用的情况
func TestKafkaServiceUnavailable(t *testing.T) {
	// 修改为一个错误或不存在的Broker地址
	invalidConfig := &KafkaConfig{
		Brokers:    "invalid:9092",
		Username:   kafkaConfig.Username,
		Password:   kafkaConfig.Password,
		Mechanisms: kafkaConfig.Mechanisms,
		Protocol:   kafkaConfig.Protocol,
	}

	_, err := NewKafka(invalidConfig)
	if err == nil {
		t.Fatal("Expected error when creating client with invalid broker, got none")
	}
	t.Log("Received expected error:", err)
}

// 多并发发布和订阅
func TestConcurrentProducers(t *testing.T) {
	client, _ := NewKafka(kafkaConfig)
	defer client.Close()

	var wg sync.WaitGroup
	producerCount := 10
	messageCountPerProducer := 100

	start := time.Now()
	for i := 0; i < producerCount; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < messageCountPerProducer; j++ {
				message := fmt.Sprintf("Message %d from producer %d", j, producerID)
				if err := client.Publish(topic, message); err != nil {
					t.Error("Failed to publish message:", err)
				}
			}
		}(i)
	}
	wg.Wait()
	duration := time.Since(start)
	totalMessages := producerCount * messageCountPerProducer
	t.Logf("Total messages: %d, Duration: %v", totalMessages, duration)
}

//// 主题和分区的动态操作
//func TestTopicCreateDelete(t *testing.T) {
//	client, _ := NewKafka(kafkaConfig)
//	defer client.Close()
//
//	// 假设有一个可以创建和删除主题的方法
//	topicName := "new-test-topic"
//	if err := client.CreateTopic(topicName, 1, 1); err != nil { // 假设参数分别为主题名、分区数、副本数
//		t.Fatalf("Failed to create topic: %s", err)
//	}
//	t.Log("Topic created successfully")
//
//	if err := client.DeleteTopic(topicName); err != nil {
//		t.Fatalf("Failed to delete topic: %s", err)
//	}
//	t.Log("Topic deleted successfully")
//}
