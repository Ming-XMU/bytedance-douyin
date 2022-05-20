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
	CreatVideoList(user int) []models.VOVideo
	GetAuthor(id int) models.VOUser
}
type FeedServiceImpl struct {
	feedDao daos.FeedDao
}

func (f *FeedServiceImpl) PublishAction(c *gin.Context) (err error) {
	//verify title
	title := c.PostForm("title")
	if tools.VerifyParamsEmpty(title){
		err = errors.New("title is empty..")
		return
	}
	//get user id from token
	token := c.PostForm("token")
	tokenKey, err := tools.JwtParseTokenKey(token)
	user, err := tools.RedisTokenKeyValue(tokenKey)
	userId := user.UserId
	file, err := c.FormFile("data")
	//create play_url
	filename := filepath.Base(file.Filename)
	finalName := fmt.Sprintf("%d_%s", userId, filename)
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
		Title: title,
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
//author: wechan
func (f *FeedServiceImpl) GetJsonFeeCache() (VideoList []models.Video, err error) {
	VideoList = make([]models.Video, 0, 31)
	//连接redis
	rec, err := redis.Dial("tcp", "120.78.238.68:6379")
	if err != nil {
		log.Println("redis dial failed,err:", err.Error())
		return nil, err
	}
	//从redis获取数据
	videoCache, err := redis.Values(rec.Do("Lrange", "video_cache", 0, -1))
	if err != nil {
		log.Println("get redis video_cache failed,err:", err.Error())
		return nil, err
	}
	if videoCache == nil || len(videoCache) < 1 { //读不到redis数据
		log.Println("video cache:", videoCache)
		log.Println("redis no data")
		return nil, err
	}
	//遍历数据反序列化
	for _, val := range videoCache {
		var video models.Video
		err = json.Unmarshal(val.([]byte), &video)
		VideoList = append(VideoList, video)
	}
	return VideoList, err
}

// CreatVideoList 获取视频流列表
// author:wechan
func (f *FeedServiceImpl) CreatVideoList(user int) (videolist []models.VOVideo) {
	videolist = make([]models.VOVideo, 0, 31)
	var videoret models.VOVideo
	videos, err := f.GetJsonFeeCache()
	if err != nil || videos == nil {
		//fmt.Println("create video list get redis cache failed,err:", err.Error())
		//fmt.Println("len of video cache: ", len(videos))
		return models.VODemoVideos //获取不到redis缓存数据，直接返回demovideos
	}
	for _, singlevideo := range videos {
		videoret.Id = singlevideo.ID
		videoret.CoverUrl = singlevideo.CoverUrl
		videoret.PlayUrl = singlevideo.PlayUrl
		videoret.CommentCount = singlevideo.CommentCount
		videoret.FavoriteCount = singlevideo.FavoriteCount
		videoret.Author = f.GetAuthor(int(singlevideo.UserId))
		if user == 0 {
			videoret.IsFavorite = false
		} else {
			videoret.IsFavorite = GetFavoriteService().FavoriteJudge(user, int(singlevideo.ID))
		}
		videoret.Title = singlevideo.Title
		videolist = append(videolist, videoret)
	}
	return videolist
}

func (f *FeedServiceImpl) GetAuthor(id int) (Author models.VOUser) {
	getuser, err := GetUserService().UserInfo(id)
	if err != nil {
		fmt.Println("get authors failed,err: ", err.Error())
		return models.VODemoUser
	}
	Author.Id = getuser.Id
	Author.Name = getuser.Name
	Author.FollowCount = getuser.FollowCount
	Author.FollowerCount = getuser.FollowerCount
	Author.IsFollow = false // 要查表，未完成
	return Author
}
