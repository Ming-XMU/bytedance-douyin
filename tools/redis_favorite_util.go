package tools

import (
	"douyin/models"
	"github.com/gomodule/redigo/redis"
)

/**
 * @Author: Ember
 * @Date: 2022/5/21 13:47
 * @Description: TODO
 **/
var (
	//video favorite cache
	DefaultVideoFavoriteCaches = map[int]string{
		1 : "video_cache_1",
		2 : "video_cache_2",
		3 : "video_cache_3",
	}
	//Video Favorite Caches Length
	DefaultVideoFavoriteCacheLength int = 3;
	//User Favorite Video Cache prefix
	DefaultUserFavoriteVideoCachePrefix = "user_favorite_video_cache:"
)

//cache favorite
func RedisCacheFavorite(favorite *models.Favorite) (interface{},error){
	//calculate hash
	key := hashKey(favorite.VideoId)
	//get cache name
	cacheName := DefaultVideoFavoriteCaches[key]
	favoriteVideoName := packageUserFavoriteCacheName(favorite.UserId)
	luaScript := "if redis.call(\"SISMEMBER\",KEYS[3],KEYS[2]) == 1 then" +
		"\n    return 2\nelse\n    redis.call(\"SADD\",KEYS[3],KEYS[2])\n    " +
		"if redis.call(\"EXISTS\",KEYS[1]) == 0 " +
		"or redis.call(\"HEXISTS\",KEYS[1],KEYS[2]) == 0 then" +
		"\n    \tredis.call(\"HSET\",KEYS[1],KEYS[2],1)\n\telse\n    \t" +
		"redis.call(\"HINCRBY\",KEYS[1],KEYS[2],1)\n        end\n    return 1\n    end\nend"
	conn := models.GetRec()
	script := redis.NewScript(2, luaScript)
	return script.Do(conn, cacheName, favorite.VideoId,favoriteVideoName)
}

//cancel favorite
func RedisCacheCancelFavorite(favorite *models.Favorite)(interface{},error){
	//calculate hash
	key := hashKey(favorite.VideoId)
	//get cache name
	cacheName := DefaultVideoFavoriteCaches[key]
	favoriteVideoName := packageUserFavoriteCacheName(favorite.UserId)
	luaScript := "if redis.call(\"SISMEMBER\",KEYS[3],KEYS[2]) == 0 then" +
		"\n    return 2\nelse\n    redis.call(\"SREM\",KEYS[3],KEYS[2])\n    " +
		"if redis.call(\"EXISTS\",KEYS[1]) == 0 or " +
		"redis.call(\"HEXISTS\",KEYS[1],KEYS[2]) == 0 then\n    \t" +
		"redis.call(\"HSET\",KEYS[1],KEYS[2],-1)\n\telse\n    \t" +
		"redis.call(\"HINCRBY\",KEYS[1],KEYS[2],-1)\n        end\n    return 1\n    end\nend"
	conn := models.GetRec()
	script := redis.NewScript(2, luaScript)
	return script.Do(conn, cacheName, favorite.VideoId,favoriteVideoName)
}
//calculate hash key
func hashKey(videoId int)int{
	return videoId % DefaultVideoFavoriteCacheLength
}
//get user favorite video cache name
func packageUserFavoriteCacheName(userId int) string{
	return DefaultUserFavoriteVideoCachePrefix + string(userId)
}