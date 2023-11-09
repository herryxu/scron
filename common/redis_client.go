package common

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/olaola-chat/slp-tools/redis_locker"
)

var redisClient = &redis.Client{}
var once sync.Once

type redisConfig struct {
	Host     string
	Port     int
	Password string
}

func NewRedisLocker(key, taskKey string, ttl int, client *redis.Client) redis_locker.RedisLockInter {
	return redis_locker.NewCronLock(context.Background(), client, key, taskKey,
		redis_locker.WithAutoRenew(),
		redis_locker.WithTimeout(time.Duration(ttl)*time.Second))
}

// RedisClient 根据name实例化redis对象
func NewRedisClient() *redis.Client {
	//instanceKey := "slp-tools-go-redis"
	once.Do(func() {
		config := redisConfig{}
		//err := g.Cfg().GetStruct(fmt.Sprintf("go-redis.%s", name), &config)
		err := getConf(&config)
		if err != nil {
			panic(fmt.Errorf("NewRedisClient config err:%v", err))
		}
		addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
		options := redis.Options{
			Addr:               addr,
			Dialer:             nil,
			OnConnect:          nil,
			Password:           config.Password,
			DB:                 0,
			MaxRetries:         3,
			MinRetryBackoff:    0,
			MaxRetryBackoff:    0,
			DialTimeout:        0,
			ReadTimeout:        0,
			WriteTimeout:       0,
			PoolSize:           20,
			MinIdleConns:       4,
			MaxConnAge:         0,
			PoolTimeout:        0,
			IdleTimeout:        0,
			IdleCheckFrequency: 0,
			TLSConfig:          nil,
		}
		// 新建一个client
		redisClient = redis.NewClient(&options)
		log.Println("new redisClient:", redisClient)
	})
	// 告警
	if redisClient != nil {
		return redisClient
	}
	//alarmIns := alarm.GetAlarmInstance()
	// 告警
	//alarmIns.SendAlarm(fmt.Sprintf("slp-tools.NewRedisClient错误,请及时处理！@all"), "slp-tools.redis.alarm", 5*time.Minute)
	panic(fmt.Errorf("NewRedisClient err,redisClient:%v", redisClient))
}
func getConf(config *redisConfig) error {
	if RunMode == "prod" {
		config.Host = "r-bp13qyr3ykmuvb7jv5.redis.rds.aliyuncs.com"
		config.Port = 6379
		config.Password = "slp_tools_rds2023"
	} else {
		config.Host = "127.0.0.1"
		config.Port = 6379
		config.Password = ""
	}
	return nil
}
