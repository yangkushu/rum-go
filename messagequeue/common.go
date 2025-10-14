package messagequeue

// SubscriptionType 订阅类型
type SubscriptionType int

// Topic 主题
type Topic string

const (
	Exclusive SubscriptionType = iota // 独占模式
	Shared                            // 共享模式，通过轮询只有一个消费者收到消息
	Failover                          // 灾备模式，只给给一个消费者，这个消费者下线后，会确定下一个消费者
	KeyShared
)
