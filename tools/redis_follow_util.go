package tools

import (
	"douyin/models"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"time"
)

const (
	followActionLimitPrefix = "follow_action_limit"
	followLimitTime         = 60 //指定时间区间30s内
	followLimitNum          = 30 //限制关注次数
	followExpireTime        = 60 //60s过期
)

func getFollowLimitKey(userId int64) string {
	key := fmt.Sprintf("%s_%d", followActionLimitPrefix, userId)
	return key
}

// IsActionLimit
// @author zia
// @Description: 定义关注限流方法 | 维护时间区间内的关注次数
// @param userId 用户id
// @return bool false 放行 | true 拦截
// @return error
func IsActionLimit(userId int64) (bool, error) {
	rec := models.GetRec()
	defer CloseConn(rec)
	uid, err := uuid.NewUUID()
	if err != nil {
		return false, err
	}
	cur := time.Now().Unix()
	pre := time.Now().Unix() - followLimitTime
	luaScript := "redis.call('zremrangeByScore', KEYS[1], 0, ARGV[1])\n" +
		"local res = redis.call('zcard', KEYS[1])\n" +
		"if (res == nil) or (res < tonumber(ARGV[3])) then\n" +
		"	redis.call('zadd', KEYS[1], ARGV[2], ARGV[4])\n" +
		"	redis.call('expire',KEYS[1],ARGV[5])" +
		"	return 0\n" +
		"else \n" +
		"	redis.call('expire',KEYS[1],ARGV[5])\n" +
		"	return 1\n" +
		"end"
	script := redis.NewScript(1, luaScript)
	res, err := script.Do(rec, getFollowLimitKey(userId), pre, cur, followLimitNum, uid.String(), followExpireTime)
	if err != nil {
		return false, err
	}
	if res.(int64) == 1 {
		return true, nil
	}
	return false, nil
}
