package scron

import (
	"context"
	"fmt"
	scron "github.com/olaola-chat/slp-tools/scron/cron_locker"
	"math"
	"time"

	"github.com/olaola-chat/slp-tools/alarm"
	"github.com/olaola-chat/slp-tools/common"
	"github.com/olaola-chat/slp-tools/sys_info"
)

var TaskLockError = "lock Taskkey failed"

// Status returns the status of entry.
func (entry *Entry) Status() int {
	return entry.status
}

// 获取睡眠时间
func getSleepTime(cpuFl, memoryFl float64) int {
	var (
		number = 100
		cpu    = 0
		mem    = 0
	)
	if cpuFl >= 0 {
		cpu = int(math.Ceil(cpuFl))
	}
	if memoryFl >= 0 {
		mem += int(math.Ceil(memoryFl))
	}
	if cpu+mem > number {
		return number
	}
	return cpu + mem
}
func (entry *Entry) getLock() bool {
	// 根据任务时间生成过期时间
	ttl := entry.getGapTime(entry.Next)
	sysInfo := sys_info.GetSysInfo()
	serverIp := sysInfo.BestServerIp()
	// 不是最佳服务ip
	if serverIp != sysInfo.Ip {
		// 根据负载生成一个睡眠时间
		time.Sleep(time.Duration(getSleepTime(sysInfo.Cpu, sysInfo.Memory)) * time.Millisecond)
	}

	key := entry.GetCronExecKey(entry.Next)
	taskKey := entry.GetTaskExecKey()
	// 理想状态下分配给状态最佳的服务
	redisLocker := scron.NewRedisLocker(key, taskKey, ttl, scron.NewRedisClient())
	if err := redisLocker.Lock(); err != nil {
		if TaskLockError == err.Error() {
			alarmIns := alarm.GetAlarmInstance()
			if common.RunMode == "prod" {
				alarmIns.SendAlarm(fmt.Sprintf("slp-tools.任务:%s,在下一个执行期未结束，请及时处理！@all", entry.Name), "slp-tools.cron.alarm:"+entry.Name, 5*time.Minute)
			}
		}
		return false
	}
	entry.Locker = redisLocker
	return true
}

// checkFinal
func (entry *Entry) checkFinal(ttl int) bool {
	// first start
	retryTimes := 0
	if entry.Prev.IsZero() {
		return true
	}
	key := entry.GetTaskExecKey()
	redisClient := common.NewRedisClient()
LoopTwice:
	if result, err := redisClient.SetNX(context.Background(), key, time.Now().Unix(), time.Duration(ttl)*time.Second).Result(); !result || err != nil {
		time.Sleep(100 * time.Millisecond)
		retryTimes++
		if retryTimes < 2 {
			goto LoopTwice
		}
		return false
	}
	return true
}

func (entry *Entry) getGapTime(now time.Time) int {
	nextTime := entry.Schedule.Next(now)
	subTime := nextTime.Unix() - now.Unix()
	if subTime <= 60*60*24 {
		return int(subTime)
	} else {
		// 如果执行时长超过一天
		return 60 * 60 * 24
	}
}

// 获取key 任务名+日期
func (entry *Entry) GetCronExecKey(now time.Time) string {
	return fmt.Sprintf("cron_%s%v", entry.Name, now.Format(common.SecondPrettyStrFormat))
}

// 获取key 任务名
func (entry *Entry) GetTaskExecKey() string {
	return fmt.Sprintf("exec_%s", entry.Name)
}

func (entry *Entry) releaseLock() {
	if err := entry.Locker.UnLock(); err != nil {
		if common.RunMode == "prod" {
			alarmIns := alarm.GetAlarmInstance()
			alarmIns.SendAlarm(fmt.Sprintf("cron:%v,err:%v,请及时处理！@lion(里奥)", entry.Name, err), "slp-tools.redis.alarm", 5*time.Minute)
		}
	}
}
