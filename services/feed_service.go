package services

import (
	"douyin/models"
	"douyin/tools"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
)
import "douyin/daos"

/**
 * @Author: Ember
 * @Date: 2022/5/16 10:02
 * @Description: TODO
 **/
const (
	//TODO 绝对路径待修改
	//producing: /root/douyin/video/  /root/douyin/img
	Play_Url_Path  = "D:/goProject/src/simple-demo/public/video/"
	Cover_Url_Path = "D:/goProject/src/simple-demo/public/img/"
)

type FeedService interface {
	//public video
	PublishAction(c *gin.Context) error
}
type FeedServiceImpl struct {
	feedDao daos.FeedDao
}

func (f *FeedServiceImpl) PublishAction(c *gin.Context) (err error) {
	//TODO get user_id from token
	token := c.PostForm("token")
	tokenKey, err := tools.JwtParseTokenKey(token)
	user, err := tools.RedisTokenKeyValue(tokenKey)
	userId := user.UserId
	file, err := c.FormFile("data")
	//create play_url
	filename := filepath.Base(file.Filename)
	finalName := fmt.Sprintf("%s_%s", userId, filename)
	saveFile := filepath.Join("./public/video", finalName)
	//create video
	if err = c.SaveUploadedFile(file, saveFile); err != nil {
		//TODO log format
		fmt.Println("create video failed:", err.Error())
		err = errors.New("create video failed...")
		return
	}
	//TODO check multiply

	//Create CoverUrl
	//cmd format :ffmpeg -i  1_mmexport1652668404330.mp4 -ss 00:00:00 -frames:v 1 out.jpg
	coverFile := finalName + ".jpg"
	playUrl := Play_Url_Path + finalName
	coverUrl := Cover_Url_Path + coverFile
	cmd := exec.Command("ffmpeg", "-i", playUrl, "-ss", "00:00:00", "-frames:v", "1", coverUrl)
	err = cmd.Run()
	if err != nil {
		//TODO log format
		fmt.Println("create cover failed:", err.Error())
		err = errors.New("create cover failed..")
		return
	}
	//Save Db
	video := models.Video{
		UserId:        userId,
		PlayUrl:       playUrl,
		CoverUrl:      coverUrl,
		CommentCount:  0,
		FavoriteCount: 0,
	}
	_, err = f.feedDao.CreateFeed(video)
	if err != nil {
		//TODO log format
		fmt.Println("create feed record failed : ", err.Error())
		return
	}
	//cache
	err = tools.RedisCacheFeed(video)
	if err != nil {
		fmt.Println("cache feed failed:", err.Error())
		return
	}
	return
}

//single create
var (
	feedService     FeedService
	feedServiceOnce sync.Once
)

func GetFeedService() FeedService {
	feedServiceOnce.Do(func() {
		feedService = &FeedServiceImpl{
			feedDao: daos.GetFeeDao(),
		}
	})
	return feedService
}

//GetJsonFeeCache 获取redis中缓存的视频数据
func GetJsonFeeCache() (VideoList []models.Video) {
	//连接redis
	rec, err := redis.Dial("tcp", "120.78.238.68:6379")
	if err != nil {
		log.Println("redis dial failed,err:", err.Error())
		//TODO 错误处理未完成
	}
	//从redis获取数据
	videoCache, err := redis.Values(rec.Do("Lrange", "video_cache", 0, -1))
	if err != nil {
		log.Println("get redis video_cache failed,err:", err.Error())
		//TODO 错误处理未完成
	}
	//遍历数据反序列化
	for _, val := range videoCache {
		var video models.Video
		json.Unmarshal(val.([]byte), &video)
		VideoList = append(VideoList, video)
	}
	//
	return VideoList
}
