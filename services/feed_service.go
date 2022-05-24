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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	//TODO 上线修改为服务器的IP地址和端口
	Show_Play_Url_Prefix = "http://localhost:8080/static/video/"
	Show_Cover_Url_Prefix = "http://localhost:8080/static/img/"
)

type FeedService interface {
	//public video
	PublishAction(c *gin.Context) error
	CreatVideoList(user int) []models.VOVideo
	GetAuthor(user, id int) (Author models.VOUser)
	//flush redis favourite
	FlushRedisFavouriteActionCache(videoId int64,count int)error
	FlushRedisFavouriteCount()
}
type FeedServiceImpl struct {
	feedDao daos.FeedDao
}

func (f *FeedServiceImpl) PublishAction(c *gin.Context) (err error) {
	//verify title
	title := c.PostForm("title")
	if tools.VerifyParamsEmpty(title) {
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
	playLocalUrl := Play_Url_Path + finalName
	coverLocalUrl := Cover_Url_Path + coverFile
	cmd := exec.Command("ffmpeg", "-i", playLocalUrl, "-ss", "00:00:00", "-frames:v", "1", coverLocalUrl)
	err = cmd.Run()
	if err != nil {
		//TODO log format
		fmt.Println("create cover failed:", err.Error())
		//del file
		err = os.Remove(saveFile)
		if err != nil{
			fmt.Println("file remove failed...")
		}
		err = errors.New("create cover failed..")
		return
	}
	//Save Db
	video := &models.Video{
		UserId:        userId,
		PlayUrl:       Show_Play_Url_Prefix + finalName,
		CoverUrl:      Show_Cover_Url_Prefix + coverFile,
		CommentCount:  0,
		FavoriteCount: 0,
		Title:         title,
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

func(f *FeedServiceImpl)FlushRedisFavouriteActionCache(videoId int64,count int) error{
	return f.feedDao.UpdateVideoFavoriteCount(videoId,count)
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
	rec := models.GetRec()
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
		videoret.Author = f.GetAuthor(user, int(singlevideo.UserId))
		if user == 0 { //未登录用户，是否点赞即为默认值未点赞
			videoret.IsFavorite = false
		} else {
			videoret.IsFavorite = GetFavoriteService().FavoriteJudge(user, int(singlevideo.ID))
		}
		videoret.Title = singlevideo.Title
		videolist = append(videolist, videoret)
	}
	return videolist
}

func (f *FeedServiceImpl) GetAuthor(user, id int) (Author models.VOUser) {
	getuser, err := GetUserService().UserInfo(strconv.Itoa(id))
	if err != nil {
		fmt.Println("get authors failed,err: ", err.Error())
		return models.VODemoUser
	}
	Author.Id = getuser.Id
	Author.Name = getuser.Name
	Author.FollowCount = getuser.FollowCount
	Author.FollowerCount = getuser.FollowerCount
	if user == 0 { //未登录用户，关注即为默认值未关注
		Author.IsFollow = false
	} else { //user-登录用户id getuser.id-视频作者id，前后关系!
		Author.IsFollow, err = daos.GetFollowDao().JudgeIsFollow(user, int(getuser.Id))
		if err != nil {
			log.Println("feed 数据库读取follow出错", err.Error())
			Author.IsFollow = false //数据库读取follow出错时，使用默认值false
		}
	}
	return Author
}

//flush redis favourite cache
func (f *FeedServiceImpl)FlushRedisFavouriteCount(){
	for _,cacheName := range tools.DefaultVideoCacheList {
		kv, err := tools.GetAllKV(cacheName)
		if err != nil{
			fmt.Println("flush error occured,cacheName :" ,cacheName)
		}
		for k,v := range kv {
			//parse int
			videoId ,err:= strconv.ParseInt(k,10,64)
			if err != nil{
				//TODO 出现局部出错
				continue
			}
			count,err := strconv.Atoi(v)
			if err != nil{
				//TODO 出现局部出错
				continue
			}
			err = f.FlushRedisFavouriteActionCache(videoId, count)
			if err != nil{
				//TODO 出现局部出错
				continue
			}
		}
	}

}