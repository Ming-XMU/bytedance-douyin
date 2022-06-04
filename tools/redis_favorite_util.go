package tools

import (
	"douyin/daos"
	"douyin/models"
	"errors"
	"github.com/gomodule/redigo/redis"
	"log"
	"strconv"
	"strings"
	"time"
)

/**
 * @Author: Ember
 * @Date: 2022/5/21 13:47
 * @Description: TODO
 **/
var (
	// DefaultVideoFavoriteCaches video favorite cache
	DefaultVideoFavoriteCaches = map[int64]string{
		0: "video_cache_favorite_0",
		1: "video_cache_favorite_1",
		2: "video_cache_favorite_2",
	}
	DefaultVideoCacheList = []string{"video_cache_favorite_0", "video_cache_favorite_1", "video_cache_favorite_2"}
	// DefaultVideoFavoriteCacheLength Video Favorite Caches Length
	DefaultVideoFavoriteCacheLength int64 = 3

	// DefaultUserFavoriteVideoCachePrefix User Favorite Video Cache prefix
	DefaultUserFavoriteVideoCachePrefix = "user_favorite_video_cache:"

	// DefaultFavouriteActionLimitCaches favorite action record cache
	DefaultFavouriteActionLimitCaches = map[int64]string{
		0: "user_cache_limit_rate_favorite_action_0",
		1: "user_cache_limit_rate_favorite_action_1",
		2: "user_cache_limit_rate_favorite_action_2",
	}
	DefaultFavouriteActionLimitList = []string{
		"user_cache_favorite_limit_rate_action_0",
		"user_cache_favorite_limit_rate_action_1",
		"user_cache_favorite_limit_rate_action_2",
	}
	DefaultFavouriteActionLimitLength int64 = 3
	DefaultFavouriteActionLimitPrefix       = "user_favourite_limit:"
)

// RedisCacheFavorite cache favorite
func RedisCacheFavorite(favorite *models.Favorite) (interface{}, error) {
	//calculate hash
	key := hashKey(favorite.VideoId)
	//get cache name
	cacheName := DefaultVideoFavoriteCaches[key]
	favoriteVideoName := PackageUserFavoriteCacheName(favorite.UserId)
	luaScript := "if redis.call(\"SISMEMBER\",KEYS[3],KEYS[2]) == 1 then\n    return 2" +
		"\nelse\n    redis.call(\"SADD\",KEYS[3],KEYS[2])\n    " +
		"if redis.call(\"EXISTS\",KEYS[1]) == 0 or redis.call(\"HEXISTS\",KEYS[1],KEYS[2]) == 0 " +
		"then\n    \tredis.call(\"HSET\",KEYS[1],KEYS[2],1)\n        " +
		"return 1\n\telse\n    \tredis.call(\"HINCRBY\",KEYS[1],KEYS[2],1)\n    \treturn 1\n\tend\nend"
	conn := models.GetRec()
	defer CloseConn(conn)
	script := redis.NewScript(3, luaScript)
	return script.Do(conn, cacheName, favorite.VideoId, favoriteVideoName)
}

// RedisGetUserFavouriteCache get user favourite cache
func RedisGetUserFavouriteCache(userId int64) (favouriteList []int64, err error) {
	rec := models.GetRec()
	defer CloseConn(rec)
	cacheName := PackageUserFavoriteCacheName(userId)
	values, err := redis.Values(rec.Do("smembers", cacheName))
	if err != nil {
		return
	}
	list := make([]int64, len(values))
	for _, v := range values {
		list = append(list, v.(int64))
	}
	return list, nil
}

// RedisCacheCancelFavorite cancel favorite
func RedisCacheCancelFavorite(favorite *models.Favorite) (interface{}, error) {
	//calculate hash
	key := hashKey(favorite.VideoId)
	//get cache name
	cacheName := DefaultVideoFavoriteCaches[key]
	favoriteVideoName := PackageUserFavoriteCacheName(favorite.UserId)
	luaScript := "if redis.call(\"SISMEMBER\",KEYS[3],KEYS[2]) == 0 then\n    return 2" +
		"\nelse\n    redis.call(\"SREM\",KEYS[3],KEYS[2])\n    " +
		"if redis.call(\"EXISTS\",KEYS[1]) == 0 or redis.call(\"HEXISTS\",KEYS[1],KEYS[2]) == 0 " +
		"then\n    \tredis.call(\"HSET\",KEYS[1],KEYS[2],-1)\n        " +
		"return 1\n\telse\n    \tredis.call(\"HINCRBY\",KEYS[1],KEYS[2],-1)\n   \t\treturn 1\n    end\nend"
	conn := models.GetRec()
	defer CloseConn(conn)
	script := redis.NewScript(3, luaScript)
	return script.Do(conn, cacheName, favorite.VideoId, favoriteVideoName)
}

//calculate hash key
func hashKey(videoId int64) int64 {
	return videoId % DefaultVideoFavoriteCacheLength
}

// PackageUserFavoriteCacheName get user favorite video cache name
func PackageUserFavoriteCacheName(userId int64) string {
	formatUserId := strconv.FormatInt(userId, 10)
	return DefaultUserFavoriteVideoCachePrefix + formatUserId
}

// UnPackUserFavoriteCacheName get userid from cacheName
func UnPackUserFavoriteCacheName(cacheName string) (userid int64, err error) {
	split := strings.Split(cacheName, ":")
	if len(split) < 2 {
		err = errors.New("cacheName is not corrept")
		return
	}
	str := split[1]
	userid, err = strconv.ParseInt(str, 10, 64)
	return
}

// FavouriteRateLimit use favourite rate limit
//限流规则：5分钟10次、一天不超过100
func FavouriteRateLimit(userId int64) (result interface{}, err error) {
	rec := models.GetRec()
	defer CloseConn(rec)
	luaScript := "if redis.call(\"EXISTS\",KEYS[1]) ~= 0 and " +
		"tonumber(redis.call(\"HGET\",KEYS[1],KEYS[2])) > 100 " +
		"then\n    return 0\nend\nredis.call(\"HINCRBY\",KEYS[1],KEYS[2],1)\n" +
		"if redis.call(\"LLEN\",KEYS[3]) < 10 then\n    " +
		"redis.call(\"LPUSH\",KEYS[3],KEYS[4])\n    " +
		"return 1\nend\nif tonumber(KEYS[4]) - tonumber(redis.call(\"LINDEX\",KEYS[3],9)) < 300 " +
		"then\n    return 0\nend\nredis.call(\"RPOP\",KEYS[3])\n" +
		"redis.call(\"LPUSH\",KEYS[3],KEYS[4])\nredis.call(\"EXPIRE\",KEYS[3],300)\nreturn 1"
	cacheName := GetFavouriteRateLimitCache(userId)
	limitListName := PackageFavouriteRateLimitListName(userId)
	script := redis.NewScript(4, luaScript)
	return script.Do(rec, cacheName, userId, limitListName, time.Now().Unix())
}

func GetFavouriteRateLimitCache(userId int64) string {
	return DefaultFavouriteActionLimitCaches[userId%DefaultFavouriteActionLimitLength]
}
func PackageFavouriteRateLimitListName(userId int64) string {
	formatUserId := strconv.FormatInt(userId, 10)
	return DefaultFavouriteActionLimitPrefix + formatUserId
}
func FavouriteRateLimitDel() error {
	rec := models.GetRec()
	defer CloseConn(rec)
	luascript := "redis.call(\"DEL\",KEYS[1])\nredis.call(\"DEL\",KEYS[2])\nredis.call(\"DEL\",KEYS[3])"
	script := redis.NewScript(3, luascript)
	_, err := script.Do(rec, "user_cache_favorite_limit_rate_action_0",
		"user_cache_favorite_limit_rate_action_1",
		"user_cache_favorite_limit_rate_action_2")
	return err
}

// JudgeisFavoriteByredis 判断是否点赞
// Author: wechan
func JudgeisFavoriteByredis(VideoId, UserId int64) int64 {
	conn := models.GetRec()
	defer CloseConn(conn)
	favoriteVideoName := PackageUserFavoriteCacheName(UserId)
	ret, err := conn.Do("sismember", favoriteVideoName, VideoId)
	if err != nil { //读取缓存出错，默认为未点赞
		return 0
	}
	return ret.(int64)
}

// GetFavoriteCount 获取当前视频的点赞量
// Author: wechan
func GetFavoriteCount(VideoId int64) int64 {
	conn := models.GetRec()
	defer CloseConn(conn)
	//calculate hash
	key := hashKey(VideoId)
	//get cache name
	cacheName := DefaultVideoFavoriteCaches[key]
	ret, err := conn.Do("HGET", cacheName, VideoId)
	if ret == nil || err != nil { //获取点赞量失败
		log.Println("HGET favoriteCount failed,error:", err.Error())
		//从数据库中获取
		var video *models.Video
		video, err = daos.GetVideoDao().FindById(VideoId)
		if err != nil {
			return 0
		}
		return video.FavoriteCount
	}
	return ret.(int64)
}
