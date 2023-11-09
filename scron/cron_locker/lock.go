package scron

import (
	"time"
)

type RedisLockInter interface {
	// Lock 加锁
	Lock() error

	// UnLock 解锁
	UnLock() error

	// SpinLock 自旋锁
	SpinLock(timeout time.Duration) error

	// Renew 手动续期
	Renew() error
}

const lockTime = 5 * time.Second
