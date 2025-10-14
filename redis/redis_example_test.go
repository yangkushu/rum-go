package redis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRedisClusterClient(t *testing.T) {

	cfg := &Config{
		Addrs: ":7001,:7002,:7003,:7004,:7005,:7006",
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 100; i++ {

		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)

		result := client.Set(context.Background(), key, value, time.Minute)
		if result.Err() != nil {
			t.Fatal(result.Err())
		}

		v, err := client.Get(context.Background(), key).Result()
		if err != nil {
			t.Fatal(err)
		}

		if v != value {
			t.Fatal("value not equal")
		}
	}

}
