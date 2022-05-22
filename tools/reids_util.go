package tools

import (
	"douyin/models"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
)

/**
 * @Author: Ember
 * @Date: 2022/5/17 12:23
 * @Description: TODO
 **/
const (
	//video cache name
	VideoCacheName = "video_cache_set"
	//range:0 ~ 29
	VideoCacheMaxLimit = 29
	VideoCacheMinLimit = 0
	//DefaultExpirationTime 默认过期时间30分钟
	DefaultExpirationTime int = 1800
)

//cache feed
func RedisCacheFeed(video *models.Video) (err error) {
	con := models.GetRec()
	//video json
	jsonResult, err := json.Marshal(&video)
	if err != nil {
		return
	}
	_, err = con.Do("ZADD", VideoCacheName, video.CreateAt.Unix(), jsonResult)
	return
}

// RedisCacheTokenKey
// @Description: 添加User结构体
// @param conn redis connection
// @param k key
// @param v	value
// @return error
func RedisCacheTokenKey(k string, u *LoginUser, t int) error {
	conn := models.GetRec()
	_, err := conn.Do("HMSET", redis.Args{}.Add(k).AddFlat(u)...)
	if err != nil {
		return err
	}
	_, err = conn.Do("EXPIRE", k, t)
	if err != nil {
		return err
	}
	return nil
}

// RedisTokenKeyValue
// @Description: 读取User结构体
// @param conn redis connection
// @param k	key
// @return u user
// @return err
func RedisTokenKeyValue(k string) (u *LoginUser, err error) {
	conn := models.GetRec()
	v, err := redis.Values(conn.Do("HGETALL", k))
	if err != nil {
		fmt.Println("redis.Values() err: ", err)
		return nil, err
	}
	t := new(LoginUser)
	if err = redis.ScanStruct(v, t); err != nil {
		fmt.Println(err)
		return nil, err
	}
	u = t
	return u, nil
}

// RedisKeyFlush
// @Description: 刷新过期时间
// @param conn	redis connection
// @param k	键值
// @return error
func RedisKeyFlush(k interface{}) error {
	conn := models.GetRec()
	_, err := conn.Do("expire", k, DefaultExpirationTime)
	if err != nil {
		return err
	}
	return err
}

// RedisCheckKey
// @Description: 判断键是否有效
// @param conn	redis connection
// @param k	键值
// @return bool	true:有效 | false:过期
// @return error
func RedisCheckKey(k string) (bool, error) {
	//当key不存在时，返回-2，当key存在但没有设置剩余生存时间时，返回-1。否则，以毫秒为单位，返回key的剩余生存时间
	conn := models.GetRec()
	r, err := redis.Int(conn.Do("TTL", k))
	if err != nil {
		return false, err
	}
	if r == -2 {
		return false, nil
	}
	return true, nil
}

// RedisDeleteKey
// @Description: 删除键
// @receiver rec
// @param k
// @return error
func RedisDeleteKey(k string) error {
	//当key不存在时，返回-2，当key存在但没有设置剩余生存时间时，返回-1。否则，以毫秒为单位，返回key的剩余生存时间
	conn := models.GetRec()
	_, err := conn.Do("DEL", k)
	if err != nil {
		return err
	}
	return nil
}

//@author cwh
//redis操作：action name value
func RedisDoKV(action string, name, value interface{}) error {
	con := models.GetRec()
	_, err := con.Do(action, name, value)
	if err != nil {
		return err
	}
	return nil
}

//@author cwh
//redis操作：action name key value
func RedisDoHash(action string, name, key, value interface{}) error {
	con := models.GetRec()
	_, err := con.Do(action, name, key, value)
	if err != nil {
		return err
	}
	return nil
}

//@author cwh
//key存在判断
func RedisKeyExists(key interface{}) bool {
	con := models.GetRec()
	do, _ := con.Do("EXISTS", key)
	if do == 1 {
		return true
	}
	return false
}
