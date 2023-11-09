package redis_locker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

// Lock 加锁
func (lock *CronLock) Lock() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	result, err := lock.Client.SetNX(lock.Context, lock.key, lock.token, time.Duration(lock.lockTimeout.Seconds())*time.Second).Result()
	if !result || err != nil {
		return errors.New("lock key failed")
	}
	result, err = lock.Client.SetNX(lock.Context, lock.Taskkey, lock.token, time.Duration(lock.lockTimeout.Seconds())*time.Second+2).Result()
	if !result || err != nil {
		return errors.New("lock Taskkey failed")
	}
	if lock.isAutoRenew {
		lock.autoRenewCtx, lock.autoRenewCancel = context.WithCancel(lock.Context)
		go lock.autoRenew()
	}
	return nil
}

// UnLock 解锁
func (lock *CronLock) UnLock() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()

	// 如果已经创建了取消函数，则执行取消操作
	if lock.autoRenewCancel != nil {
		lock.autoRenewCancel()
	}
	if lock.Client.Get(lock.Context, lock.Taskkey).Val() == lock.token {
		if err := lock.Client.Del(lock.Context, lock.Taskkey).Err(); err != nil {
			return fmt.Errorf("failed to remove lock: %s", err)
		}
	}
	//result, err := lock.Client.Eval(lock.Context, unLockScript, []string{lock.Taskkey}, lock.token).Result()
	//if err != nil {
	//	return fmt.Errorf("failed to release lock: %w", err)
	//}
	//if result != "OK" {
	//	return errors.New("lock release failed")
	//}

	return nil
}

// SpinLock 自旋锁
func (lock *CronLock) SpinLock(timeout time.Duration) error {
	exp := time.Now().Add(timeout)
	for {
		if time.Now().After(exp) {
			return errors.New("spin lock timeout")
		}

		// 加锁成功直接返回
		err := lock.Lock()
		if err == nil {
			return nil
		}

		// 如果加锁失败，则休眠一段时间再尝试
		select {
		case <-lock.Context.Done():
			return lock.Context.Err() // 处理取消操作
		case <-time.After(100 * time.Millisecond):
			// 继续尝试下一轮加锁
		}
	}
}

// Renew 锁手动续期
func (lock *CronLock) Renew() error {
	lock.mutex.Lock()
	defer lock.mutex.Unlock()
	//res, err := lock.Client.Eval(lock.Context, renewScript, []string{lock.key}, lock.token, lock.lockTimeout.Seconds()).Result()
	//if err == redis.Nil {
	//	return nil
	//}
	if lock.Client.Get(lock.Context, lock.Taskkey).Val() == lock.token {
		if err := lock.Client.Expire(lock.Context, lock.Taskkey, time.Duration(lock.lockTimeout.Seconds()/3*2)*time.Second).Err(); err != nil {
			return fmt.Errorf("failed to renew lock: %s", err)
		}
	} else {
		return fmt.Errorf("failed to renew lock:")
	}
	return nil
}

// 锁自动续期
func (lock *CronLock) autoRenew() {
	ticker := time.NewTicker(lock.lockTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-lock.autoRenewCtx.Done():
			log.Println("autoRenew autoRenewCtx:")
			return
		case <-ticker.C:
			err := lock.Renew()
			if err != nil {
				log.Println("autoRenew failed:", err)
				return
			}
		}
	}
}
