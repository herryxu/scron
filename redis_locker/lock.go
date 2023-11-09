package redis_locker

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
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

type CronLock struct {
	context.Context
	*redis.Client
	key             string
	Taskkey         string
	token           string
	lockTimeout     time.Duration
	isAutoRenew     bool
	autoRenewCtx    context.Context
	autoRenewCancel context.CancelFunc
	mutex           sync.Mutex
}

// 默认锁超时时间
const lockTime = 5 * time.Second

type Option func(lock *CronLock)

func NewCronLock(ctx context.Context, redisClient *redis.Client, lockKey, taskKey string, options ...Option) RedisLockInter {
	lock := &CronLock{
		Context:     ctx,
		Client:      redisClient,
		lockTimeout: lockTime,
	}
	for _, f := range options {
		f(lock)
	}

	lock.key = lockKey
	lock.Taskkey = taskKey
	// token 自动生成
	if lock.token == "" {
		lock.token = fmt.Sprintf("token_%d", time.Now().UnixNano())
	}

	return lock
}

// WithKey 设置锁的key
func WithKey(key string) Option {
	return func(lock *CronLock) {
		lock.key = key
	}
}

// WithTimeout 设置锁过期时间
func WithTimeout(timeout time.Duration) Option {
	return func(lock *CronLock) {
		lock.lockTimeout = timeout
	}
}

// WithAutoRenew 是否开启自动续期
func WithAutoRenew() Option {
	return func(lock *CronLock) {
		lock.isAutoRenew = true
	}
}

// WithToken 设置锁的Token
func WithToken(token string) Option {
	return func(lock *CronLock) {
		lock.token = token
	}
}
