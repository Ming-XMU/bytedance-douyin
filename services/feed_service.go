package services

import (
	"douyin/models"
	"douyin/tools"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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
	//token := c.PostForm("token")
	userId := "1"
	//TODO Verify Token
	file, err := c.FormFile("data")
	//Verify User_id
	if tools.VerifyParamsEmpty(userId) {
		//TODO log format
		fmt.Println("user_id is empty....")
		err = errors.New("user_id can not be empty...")
		return
	}
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
	cmd := exec.Command("ffmpeg", "-i", Play_Url_Path+finalName, "-ss", "00:00:00", "-frames:v", "1", Cover_Url_Path+coverFile)
	err = cmd.Run()
	if err != nil {
		//TODO log format
		fmt.Println("create cover failed:", err.Error())
		err = errors.New("create cover failed..")
		return
	}
	//Save Db
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		//TODO log format
		fmt.Println("user_id validate..")
		return
	}
	video := models.Video{
		UserId:        id,
		PlayUrl:       saveFile,
		CoverUrl:      coverFile,
		CommentCount:  0,
		FavoriteCount: 0,
	}
	//TODO wait test connect db
	_, err = f.feedDao.CreateFeed(video)
	if err != nil {
		//TODO log format
		fmt.Println("create feed record failed : ", err.Error())
		return
	}
	return
}

//single create
var (
	feedService     FeedService
	feedServiceOnce sync.Once
)

func GetVideoService() FeedService {
	feedServiceOnce.Do(func() {
		feedService = &FeedServiceImpl{
			feedDao: daos.GetFeeDao(),
		}
	})
	return feedService
}
