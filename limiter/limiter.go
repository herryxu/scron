package limiter

import (
	"sync"
	"time"
)

var (
	alarmMap = sync.Map{}
)

func CheckLimiter(key string, maxTime int64) bool {
	var timeStamp int64
	timeStr, ok := alarmMap.Load(key)
	if ok {
		timeStamp = timeStr.(int64)
	}
	if time.Now().Unix()-timeStamp < maxTime {
		return false
	}
	alarmMap.Store(key, time.Now().Unix())
	return true
}
