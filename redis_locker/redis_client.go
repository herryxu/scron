package redis_locker

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
)

var redisClient = &redis.Client{}
var once sync.Once

type redisConfig struct {
	Host     string
	Port     int
	Password string
}

// RedisClient 根据name实例化redis对象
func NewRedisClient() *redis.Client {
	once.Do(func() {
		config := redisConfig{}
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
	// 告警
	//alarmIns := alarm.GetAlarmInstance()
	//alarmIns.SendAlarm(fmt.Sprintf("slp-tools.NewRedisClient错误,请及时处理！@all"), "xxxxkeyname", 5*time.Minute)
	panic(fmt.Errorf("NewRedisClient err,redisClient:%v", redisClient))
}
func getConf(config *redisConfig) error {
	config.Host = "127.0.0.1"
	config.Port = 6379
	config.Password = ""
	return nil
}
