package models

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

var pool *redis.Pool

func GetRec() redis.Conn {
	return pool.Get()
}

func InitRedis(url string) {
	//con, err := redis.DialURL(url)
	//if err != nil {
	//	log.Fatalln(err)
	//	return
	//}
	//rec = con
	redisPool := &redis.Pool{
		//最大活跃连接数，0代表无限
		MaxActive: 888,
		//最大闲置连接数
		MaxIdle: 20,
		//闲置连接的超时时间
		IdleTimeout: time.Second * 100,
		//定义拨号获得连接的函数
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(url)
		},
	}
	pool = redisPool
}
