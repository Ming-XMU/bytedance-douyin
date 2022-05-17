package tools

import (
	"douyin/models"
	"encoding/json"
	"github.com/gomodule/redigo/redis"
)

/**
 * @Author: Ember
 * @Date: 2022/5/17 12:23
 * @Description: TODO
 **/
const(
	//video cache name
	VideoCacheName = "video_cache"
	//range:0 ~ 29
	VideoCacheMaxLimit = 29
	VideoCacheMinLimit = 0
)
//do lua script
func RedisCacheFeed(video models.Video) (err error){
	con := models.GetRec()
	//create lua script
	script := "redis.call(\"LPUSH\",KEYS[1],KEYS[2])\nredis.call(\"LTRIM\",KEYS[1],KEYS[3],KEYS[4])"
	lua := redis.NewScript(4, script)
	//video json
	jsonResult, err := json.Marshal(video)
	if err != nil{
		return
	}
	_, err = lua.Do(con, VideoCacheName, jsonResult, VideoCacheMinLimit, VideoCacheMaxLimit)
	return err
}