package log

import (
	"fmt"
	"sync"
	"time"
)

type ChannelWriteSyncer struct {
	C  chan<- []byte // 使用 byte slice 作为 channel 的类型
	mu sync.Mutex
}

func newChannelWriteSyncer(c chan<- []byte) *ChannelWriteSyncer {
	return &ChannelWriteSyncer{C: c}
}

func (cws *ChannelWriteSyncer) Write(p []byte) (n int, err error) {
	//fmt.Println("write:" + string(p))
	dataToSend := make([]byte, len(p))
	// 这里如果不复制一遍的话，在并发的情况下有可能读取到的[]byte不是完整的数据，具体原因还得排查
	copy(dataToSend, p)

	select {
	case cws.C <- dataToSend:
	case <-time.After(1 * time.Second):
		// 这里是为了防止阻塞，如果阻塞了超过1秒就丢弃日志
	}
	//cws.C <- p
	return len(p), nil
}

// 加锁也不好使
//func (cws *WriteSyncerChan) Write(p []byte) (n int, err error) {
//	cws.mu.Lock()
//	defer cws.mu.Unlock()
//	fmt.Println("write:" + string(p))
//	cws.C <- p
//	return len(p), nil
//}

func (cws *ChannelWriteSyncer) Sync() error {
	fmt.Println("sync")
	return nil // 在这里，我们不需要特别的同步操作
}
