package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	redsyncGoRedis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goRedis "github.com/redis/go-redis/v9"
	"github.com/yangkushu/rum-go/iface"
	"strings"
	"time"
)

type Client struct {
	goRedis.UniversalClient
	config *Config
}

func NewClient(config *Config) (*Client, error) {

	c := &Client{
		config: config,
	}

	// 解析地址
	addrs := c.parseAddrs(config)
	//log.Info("redis addr", log.Any("addrs", addrs))

	if len(addrs) == 0 {
		return nil, errors.New("redis address is empty")
	}

	// 使用通用的客户端,当addrs为多个时，自动使用redis cluster
	opts := &goRedis.UniversalOptions{
		Addrs:    addrs,
		Password: config.Password,
	}
	// 如果配置了用户名，就使用用户名
	if config.User != "" {
		opts.Username = config.User
	}
	c.UniversalClient = goRedis.NewUniversalClient(opts)

	// 检查是否能够连通
	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis client ping failed:%w", err)
	}

	return c, nil
}

// NewClientWithHook 创建一个带Hook的Client
func NewClientWithHook(cfg *Config, hooks []goRedis.Hook) (*Client, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if hooks != nil {
		for _, hook := range hooks {
			client.AddHook(hook)
		}
	}
	return client, nil
}

//// GetConfig 获取配置
//func (r *Client) GetConfig() *Config {
//	return r.config
//}

// NewLock 创建一个分布式锁,使用默认的过期时间[8s]和重试次数[32]
func (r *Client) NewLock(name string) iface.IDistributedLock {
	rs := r.newRedsync()
	return rs.NewMutex(name)
}

// NewLockWithExpiryTries 创建一个带过期时间和重试次数的分布式锁
func (r *Client) NewLockWithExpiryTries(name string, expiry time.Duration, tries int) iface.IDistributedLock {
	return r.newMutexWithOptions(name, redsync.WithExpiry(expiry), redsync.WithTries(tries))
}

// NewLockWithExpiry 创建一个带过期时间的分布式锁
func (r *Client) NewLockWithExpiry(name string, expiry time.Duration) iface.IDistributedLock {
	return r.newMutexWithOptions(name, redsync.WithExpiry(expiry))
}

//// NewLockWithOptions 创建一个带选项的分布式锁
//func (r *Client) NewLockWithOptions(name string, options ...redsync.Option) iface.IDistributedLock {
//	return r.newMutexWithOptions(name, options...)
//}

func (r *Client) newMutexWithOptions(name string, options ...redsync.Option) *redsync.Mutex {
	rs := r.newRedsync()
	return rs.NewMutex(name, options...)
}

func (r *Client) newRedsync() *redsync.Redsync {
	pool := redsyncGoRedis.NewPool(r)
	return redsync.New(pool)
}

// parseAddrs 从addrs、addr、port 解析地址
func (r *Client) parseAddrs(config *Config) []string {
	// 先取 addrs，如果配置了多个地址，就使用cluster方式连接，返回地址列表 比如 127.0.0.1:7000,127.0.0.1:7001,127.0.0.1:7002
	if config.Addrs != "" {
		// 如果配置了多个地址，就使用cluster方式连接
		return strings.Split(config.Addrs, ",")
	}

	// 如果没有配置 addrs，就使用 addr 和 port
	// 如果addr包含端口，就直接使用
	if strings.Contains(config.Addr, ":") {
		return []string{config.Addr}
	}

	// 如果addr不包含端口，就使用addr和port拼接
	if config.Port != 0 {
		return []string{fmt.Sprintf("%s:%d", config.Addr, config.Port)}
	}
	return []string{config.Addr}
}

// Close 关闭连接
func (r *Client) Close() error {
	return r.UniversalClient.Close()
}

// IsLockAlreadyExist 判断是否是锁已经存在
func IsLockAlreadyExist(err error) bool {
	return errors.Is(err, redsync.ErrFailed)
}

// IsKeyNotExist 判断是否是key不存在。如果传入err=nil，也会返回false
func IsKeyNotExist(err error) bool {
	return errors.Is(err, goRedis.Nil)
}
