package redis

import (
	"sync"
	"testing"
	"time"
)

func TestRedsync(t *testing.T) {

	cfg := &Config{
		Addrs: ":7001,:7002,:7003,:7004,:7005,:7006",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lock := client.NewLock("test_lock_123")
			err := lock.TryLock()
			if err != nil {
				if IsLockAlreadyExist(err) {
					t.Log("Failed to acquire lock ,client:", i, err)
				} else {
					t.Error(err)
					return
				}

			}

			t.Log("Acquired lock, client:", i)
			time.Sleep(time.Second * 1)
			lock.Unlock()

		}(i)
	}

	wg.Wait()
}
