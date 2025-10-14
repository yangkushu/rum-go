package iface

// IDistributedLock 分布式锁
type IDistributedLock interface {
	Lock() error
	Unlock() (bool, error)
	TryLock() error
}
