package sys_info

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"math"
	"strconv"
	"time"

	"github.com/olaola-chat/slp-tools/common"
)

var (
	serverLoadInfoKey = "slp_tools_server_load_info"
)

func (s *SysInfo) ReportEnv(env string) {
	if env != "" {
		common.RunMode = env
	}
}

// 上报当前服务器资源情况
func (s *SysInfo) Report() {
	if s.Ip == "" {
		return
	}
	ctx := context.Background()
	// score 通过日期+cpu+mem 就可以得到一个活跃的最佳服务分值
	var score = time.Now().Format(common.DatePrettyFormat)
	redisClient := common.NewRedisClient()
	if s.Cpu >= 0 {
		cpu := int(math.Ceil(s.Cpu))
		score = fmt.Sprintf("%s%s", score, formatPercent(cpu))
	}
	if s.Memory >= 0 {
		mem := int(math.Ceil(s.Memory))
		score = fmt.Sprintf("%s%s", score, formatPercent(mem))
	}
	scoreFloat, _ := strconv.ParseFloat(score, 64)
	if err := redisClient.ZAdd(ctx, serverLoadInfoKey, &redis.Z{
		Score:  scoreFloat,
		Member: s.Ip,
	}).Err(); err != nil {
		panic("report err")
	}
}

// 将取值区间固定在两位数
func formatPercent(number int) string {
	if number < 10 {
		return fmt.Sprintf("0%d", number)
	} else if number == 100 {
		return fmt.Sprintf("%d", number-1)
	}
	return fmt.Sprintf("%v", number)
}

// 拿到状态最好的服务器信息
func (s *SysInfo) BestServerIp() string {
	redisClient := common.NewRedisClient()
	ctx := context.Background()
	// todo 升序 去时间
	var score = time.Now().Format(common.DatePrettyFormat)
	if result, err := redisClient.ZRangeByScore(ctx, serverLoadInfoKey, &redis.ZRangeBy{
		Min:    fmt.Sprintf("%v0000", score),
		Max:    fmt.Sprintf("%v99999", score),
		Offset: 0,
		Count:  1,
	}).Result(); err == nil {
		if len(result) > 0 {
			return result[0]
		}
	}
	return ""
}
