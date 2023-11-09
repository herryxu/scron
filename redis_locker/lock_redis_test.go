package redis_locker

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewRedisLock(t *testing.T) {
	locker := NewRedisLocker(context.Background(), NewRedisClient(), "test_redis", WithTimeout(60*time.Second), WithAutoRenew())
	err := locker.Lock()
	fmt.Println(err)
	defer locker.UnLock()
	time.Sleep(10 * time.Second)
}
func TestNewRedisLockUnlock(t *testing.T) {
	locker := NewRedisLocker(context.Background(), NewRedisClient(), "test_redis", WithTimeout(60*time.Second), WithAutoRenew())
	err := locker.UnLock()
	fmt.Println(err)
}
