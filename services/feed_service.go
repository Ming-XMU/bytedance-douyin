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
		return errors.New("??????????????????????????????????????????")
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
	//???????????????minio
	//err = tools.UploadFileToMinio("img", coverFile, coverLocalUrl, "image/jpeg")
	//if err != nil {
	//	console.Error(err)
	//	err = errors.New("cover upload to minio failed")
	//	return
	//}
	//????????????????????????cos?????????
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
		//?????????????????????????????????minio??????????????????????????????
		err = os.Remove(coverLocalUrl)
		if err != nil {
			//????????????
			logrus.Errorln(err)
			return
		}
		err = os.Remove(playLocalUrl)
		if err != nil {
			//????????????
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

//GetJsonFeeCache ??????redis????????????????????????
//author: wechan
func (f *FeedServiceImpl) GetJsonFeeCache(latestTime int64) (VideoList []models.Video, err error) {
	VideoList = make([]models.Video, 0, 31)
	//??????redis
	rec := models.GetRec()
	//???redis????????????
	err = f.VideoCacheCdRedis(latestTime) //??????redis?????????????????????????????????????????????redis??????
	if err != nil {
		logrus.Errorln("VideoCacheCdRedis failed,err", err.Error())
	}
	videoCache, err := redis.Values(rec.Do("ZRevRangeByScore", "video_cache_set", latestTime, 0, "limit", 0, 29))
	if err != nil {
		//redis?????????????????????????????????
		logrus.Errorln("get redis video_cache failed,err:", err.Error())
		VideoList, err = f.feedDao.GetVideosByCreateAt(latestTime)
		if err != nil || len(VideoList) < 1 {
			return nil, err
		}
		return VideoList, err
	}
	if videoCache == nil || len(videoCache) < 1 {
		//?????????redis???????????????????????????
		logrus.Warnln("video cache is no get:", videoCache)
		VideoList, err = f.feedDao.GetVideosByCreateAt(latestTime)
		if err != nil || len(VideoList) < 1 {
			return nil, err
		}
		return VideoList, err
	}
	//????????????????????????
	for _, val := range videoCache {
		var video models.Video
		err = json.Unmarshal(val.([]byte), &video)
		VideoList = append(VideoList, video)
	}
	return VideoList, err
}

// CreatVideoList ?????????????????????
// author:wechan
func (f *FeedServiceImpl) CreatVideoList(user int, latestTime int64) (videolist []models.VOVideo, newestTime int64) {
	videolist = make([]models.VOVideo, 0, 31)
	var videoret models.VOVideo
	videos, err := f.GetJsonFeeCache(latestTime)
	if err != nil || videos == nil {
		//fmt.Println("create video list get redis cache failed,err:", err.Error())
		//fmt.Println("len of video cache: ", len(videos))
		return models.VODemoVideos, time.Now().Unix() //???redis???mysql????????????????????????????????????demovideos
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
		//}//????????????????????????????????????....
		videoret.Author = f.GetAuthor(user, int(singlevideo.UserId))
		//videoret.Author=followService.UserFollowInfo(, strconv.Itoa(user))
		if user == 0 { //??????????????????????????????????????????????????????
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
	Author.Avatar = getuser.Avatar
	Author.Signature = getuser.Signature
	Author.Background = getuser.BackGround
	if user == 0 { //????????????????????????????????????????????????
		Author.IsFollow = false
	} else { //user-????????????id getuser.id-????????????id???????????????!
		//Author.IsFollow, err = daos.GetFollowDao().JudgeIsFollow(user, int(getuser.Id))
		getIsFollow, err := tools.RedisDo("sismember", getFollowKey(strconv.Itoa(user)), getuser.Id)
		if err != nil {
			logrus.Errorln("feed ???????????????follow??????", err.Error())
			Author.IsFollow = false //???????????????follow???????????????????????????false
		} else {
			Author.IsFollow, _ = redis.Bool(getIsFollow, err)
		}
		//????????????????????????????????????????????????????????????
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
				//TODO ??????????????????
				console.Error(err)
				continue
			}
			count, err := strconv.Atoi(v)
			if err != nil {
				//TODO ??????????????????
				console.Error(err)
				continue
			}
			err = f.FlushRedisFavouriteActionCache(videoId, count)
			if err != nil {
				//TODO ??????????????????
				console.Error(err)
				continue
			}
		}
	}
}
