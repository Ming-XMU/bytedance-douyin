package services

import (
	"douyin/models"
	"douyin/tools"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anqiansong/ketty/console"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)
import "douyin/daos"

/**
 * @Author: Ember
 * @Date: 2022/5/16 10:02
 * @Description: TODO
 **/
const (
	Play_Url_Path  = "/go/src/simple-demo/public/video/"
	Cover_Url_Path = "/go/src/simple-demo/public/img/"

	Show_Play_Url_Prefix  = "https://static-1304359512.cos.accelerate.myqcloud.com/video/"
	Show_Cover_Url_Prefix = "https://static-1304359512.cos.accelerate.myqcloud.com/img/"

	MaxTitleLength = 100
	MinTitleLength = 10
)

type FeedService interface {
	// PublishAction public video
	PublishAction(c *gin.Context) error
	CreatVideoList(user int, latestTime int64) ([]models.VOVideo, int64)
	GetAuthor(user, id int) (Author models.VOUser)
	GetUserAllPublishVideos(userId int64) (videoList []models.VideoVo, err error)
	// FlushRedisFavouriteActionCache flush redis favourite
	FlushRedisFavouriteActionCache(videoId int64, count int) error
	FlushRedisFavouriteCount()

	VideoCacheCdRedis(time int64) error
}
type FeedServiceImpl struct {
	feedDao daos.FeedDao
}

func (f *FeedServiceImpl) verifyTitle(title string) error {
	if tools.VerifyParamsEmpty(title) {
		return errors.New("title is empty")
	}
	//check title length
	if len(title) > MaxTitleLength || len(title) < MinTitleLength {
		return errors.New("title length is limit 10 ~ 100")
	}
	//check invalid input
	_, isReplaced := tools.Replace(title, "*")
	if isReplaced {
		return errors.New("input vaild")
	}
	return nil
}

func (f *FeedServiceImpl) PublishAction(c *gin.Context) (err error) {
	//verify title
	title := c.PostForm("title")
	err = f.verifyTitle(title)
	if err != nil {
		return
	}
	//get user id from token
	token := c.PostForm("token")
	tokenKey, err := tools.JwtParseTokenKey(token)
	user, err := tools.RedisTokenKeyValue(tokenKey)
	userId := user.UserId
	file, err := c.FormFile("data")
	if err != nil {
		console.Error(err)
		return errors.New("get video data failed")
	}
	//create play_url
	filename := filepath.Base(file.Filename)
	//check invalid input
	_, isReplaced := tools.Replace(filename, "*")
	if isReplaced {
		return errors.New("出现违法词汇，请检查视频名称")
	}
	finalName := fmt.Sprintf("%d_%s", userId, filename)
	saveFile := filepath.Join("./public/video", finalName)
	//check multiply
	savePlayUrl := Show_Play_Url_Prefix + finalName
	rowsAffected, err := f.feedDao.FindVideoByPlayUrl(savePlayUrl)
	if rowsAffected > 0 {
		err = errors.New("video is existed")
		console.Warn("videoName:%s is existed", savePlayUrl)
		return
	}
	//create video
	if err = c.SaveUploadedFile(file, saveFile); err != nil {
		console.Error(err)
		err = errors.New("create video failed")
		return
	}
	//create Minio Video
	//err = tools.UploadFileObjectToMinio("video", finalName, file, "video/mp4")
	//if err != nil {
	//	console.Error(err)
	//	err = errors.New("upload video to minio failed")
	//	return
	//}
	//Create CoverUrl
	//cmd format :ffmpeg -i  1_mmexport1652668404330.mp4 -ss 00:00:00 -frames:v 1 out.jpg
	coverFile := finalName + ".jpg"
	playLocalUrl := Play_Url_Path + finalName
	coverLocalUrl := Cover_Url_Path + coverFile
	cmd := exec.Command("ffmpeg", "-i", playLocalUrl, "-ss", "00:00:00", "-frames:v", "1", coverLocalUrl)
	err = cmd.Run()
	if err != nil {
		console.Error(err)
		//del file
		err = os.Remove(saveFile)
		if err != nil {
			console.Error(err)
		}
		err = errors.New("create cover failed")
		return
	}
	//上传封面到minio
	//err = tools.UploadFileToMinio("img", coverFile, coverLocalUrl, "image/jpeg")
	//if err != nil {
	//	console.Error(err)
	//	err = errors.New("cover upload to minio failed")
	//	return
	//}
	//上传视频和封面到cos存储桶
	go func() {
		err = tools.UploadFileToCos(playLocalUrl, finalName, "video")
		if err != nil {
			err = errors.New("upload video to cos failed")
			logrus.Println(err)
			return
		}
		err = tools.UploadFileToCos(coverLocalUrl, coverFile, "img")
		if err != nil {
			err = errors.New("upload img to cos failed")
			logrus.Println(err)
			return
		}
		//如果视频封面成功上传到minio，移除本地视频和封面
		err = os.Remove(coverLocalUrl)
		if err != nil {
			//删除失败
			logrus.Errorln(err)
			return
		}
		err = os.Remove(playLocalUrl)
		if err != nil {
			//删除失败
			logrus.Println(err)
			return
		}
	}()
	saveCoverUrl := Show_Cover_Url_Prefix + coverFile
	//Save Db
	video := &models.Video{
		UserId:        userId,
		PlayUrl:       savePlayUrl,
		CoverUrl:      saveCoverUrl,
		CommentCount:  0,
		FavoriteCount: 0,
		Title:         title,
	}
	_, err = f.feedDao.CreateFeed(video)
	if err != nil {
		console.Error(err)
		return
	}
	//cache
	err = tools.RedisCacheFeed(video)
	if err != nil {
		console.Error(err)
		return
	}
	return
}

func (f *FeedServiceImpl) GetUserAllPublishVideos(userId int64) (videoList []models.VideoVo, err error) {

	videos, err := f.feedDao.GetUserVideos(userId)
	if err != nil {
		console.Error(err)
		return
	}
	videoList = make([]models.VideoVo, 0)
	info, err := GetUserService().UserInfo(fmt.Sprint(userId))
	for _, v := range videos {
		videoList = append(videoList, models.VideoVo{
			Id: v.ID,
			Author: models.UserMessage{
				Id:            info.Id,
				Name:          info.Name,
				FollowCount:   info.FollowCount,
				FollowerCount: info.FollowerCount,
				IsFollow:      false,
			},
			PlayUrl:       v.PlayUrl,
			CoverUrl:      v.CoverUrl,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			Title:         v.Title,
		})
	}
	return videoList, nil
}

func (f *FeedServiceImpl) FlushRedisFavouriteActionCache(videoId int64, count int) error {
	return f.feedDao.UpdateVideoFavoriteCount(videoId, count)
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

func (f *FeedServiceImpl) VideoCacheCdRedis(time int64) error {
	if tools.RedisKeyExists("video_cache_set") {
		return tools.RedisKeyFlush("video_cache_set")
	}
	videolist, err := f.feedDao.GetVideosByCreateAt(time)
	if err != nil {
		logrus.Errorln("get videos by creatAt failed", err.Error())
		return err
	}
	for _, video := range videolist {
		err = tools.RedisCacheFeed(&video)
		if err != nil {
			console.Error(err)
			return err
		}
	}
	_ = tools.RedisDoKV("EXPIRE", "video_cache_set", 1800)
	//	return nil
	return err
}

//GetJsonFeeCache 获取redis中缓存的视频数据
//author: wechan
func (f *FeedServiceImpl) GetJsonFeeCache(latestTime int64) (VideoList []models.Video, err error) {
	VideoList = make([]models.Video, 0, 31)
	//连接redis
	rec := models.GetRec()
	//从redis获取数据
	err = f.VideoCacheCdRedis(latestTime) //如果redis中没有数据，就从数据库重新加载redis缓存
	if err != nil {
		logrus.Errorln("VideoCacheCdRedis failed,err", err.Error())
	}
	videoCache, err := redis.Values(rec.Do("ZRevRangeByScore", "video_cache_set", latestTime, 0, "limit", 0, 29))
	if err != nil {
		//redis读取出错，从数据库读取
		logrus.Errorln("get redis video_cache failed,err:", err.Error())
		VideoList, err = f.feedDao.GetVideosByCreateAt(latestTime)
		if err != nil || len(VideoList) < 1 {
			return nil, err
		}
		return VideoList, err
	}
	if videoCache == nil || len(videoCache) < 1 {
		//读不到redis数据，从数据库读取
		logrus.Warnln("video cache is no get:", videoCache)
		VideoList, err = f.feedDao.GetVideosByCreateAt(latestTime)
		if err != nil || len(VideoList) < 1 {
			return nil, err
		}
		return VideoList, err
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
func (f *FeedServiceImpl) CreatVideoList(user int, latestTime int64) (videolist []models.VOVideo, newestTime int64) {
	videolist = make([]models.VOVideo, 0, 31)
	var videoret models.VOVideo
	videos, err := f.GetJsonFeeCache(latestTime)
	if err != nil || videos == nil {
		//fmt.Println("create video list get redis cache failed,err:", err.Error())
		//fmt.Println("len of video cache: ", len(videos))
		return models.VODemoVideos, time.Now().Unix() //从redis和mysql都获取不到数据，直接返回demovideos
	}
	for _, singlevideo := range videos {
		videoret.Id = singlevideo.ID
		videoret.CoverUrl = singlevideo.CoverUrl
		videoret.PlayUrl = singlevideo.PlayUrl
		videoret.CommentCount = GetCommentService().GetCommentCount(singlevideo.ID)
		videoret.FavoriteCount = tools.GetFavoriteCount(singlevideo.ID)
		//getuser, err := GetUserService().UserInfo(string(singlevideo.UserId))
		//if err != nil {
		//	fmt.Println("get authors failed,err: ", err.Error())
		//	videoret.Author=models.VODemoUser
		//}else {
		//	videoret.Author=followService.UserFollowInfo(getuser,string(singlevideo.UserId))
		//}//上面那接口用起来有点难改....
		videoret.Author = f.GetAuthor(user, int(singlevideo.UserId))
		//videoret.Author=followService.UserFollowInfo(, strconv.Itoa(user))
		if user == 0 { //未登录用户，是否点赞即为默认值未点赞
			videoret.IsFavorite = false
		} else {
			//videoret.IsFavorite = GetFavoriteService().FavoriteJudge(user, int(singlevideo.ID))
			getIs := tools.JudgeisFavoriteByredis(singlevideo.ID, int64(user))
			if getIs == 1 {
				videoret.IsFavorite = true
			} else {
				videoret.IsFavorite = false
			}
		}
		videoret.Title = singlevideo.Title

		videolist = append(videolist, videoret)
	}
	newestTime = time.Now().Unix()
	if len(videos) > 0 {
		newestTime = videos[0].CreateAt.Unix()
	}
	return videolist, newestTime
}

func (f *FeedServiceImpl) GetAuthor(user, id int) (Author models.VOUser) {
	getuser, err := GetUserService().UserInfo(strconv.Itoa(id))
	if err != nil {
		logrus.Errorln("get authors failed,err: ", user, err.Error())
		return models.VODemoUser
	}
	Author.Id = getuser.Id
	Author.Name = getuser.Name
	Author.FollowCount = getuser.FollowCount
	Author.FollowerCount = getuser.FollowerCount
	if user == 0 { //未登录用户，关注即为默认值未关注
		Author.IsFollow = false
	} else { //user-登录用户id getuser.id-视频作者id，前后关系!
		//Author.IsFollow, err = daos.GetFollowDao().JudgeIsFollow(user, int(getuser.Id))
		getIsFollow, err := tools.RedisDo("sismember", getFollowKey(strconv.Itoa(user)), getuser.Id)
		if err != nil {
			logrus.Errorln("feed 数据库读取follow出错", err.Error())
			Author.IsFollow = false //数据库读取follow出错时，使用默认值false
		} else {
			Author.IsFollow, _ = redis.Bool(getIsFollow, err)
		}
		//如果视频作者为自己，默认设置为已关注状态
		str := strconv.FormatInt(getuser.Id, 10)
		toId, _ := strconv.Atoi(str)
		if user == toId {
			Author.IsFollow = true
		}
	}
	return Author
}

// FlushRedisFavouriteCount flush redis favourite cache
func (f *FeedServiceImpl) FlushRedisFavouriteCount() {
	console.Info("==========flush redis favourite count cache==========")
	for _, cacheName := range tools.DefaultVideoCacheList {
		kv, err := tools.GetAllKV(cacheName)
		if err != nil {
			logrus.Errorln("flush error occured,cacheName :", cacheName)
		}
		for k, v := range kv {
			//parse int
			videoId, err := strconv.ParseInt(k, 10, 64)
			if err != nil {
				//TODO 出现局部出错
				console.Error(err)
				continue
			}
			count, err := strconv.Atoi(v)
			if err != nil {
				//TODO 出现局部出错
				console.Error(err)
				continue
			}
			err = f.FlushRedisFavouriteActionCache(videoId, count)
			if err != nil {
				//TODO 出现局部出错
				console.Error(err)
				continue
			}
		}
	}
}
