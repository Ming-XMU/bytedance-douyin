package controller

import (
	"douyin/services"
	"douyin/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	var user = 1
	//权限鉴定
	token := c.Query("token")
	t := c.Query("latest_time")

	latestTime, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		fmt.Println("时间戳转换失败")

	}
	loginuser, err := tools.VeifyToken(token)
	if err != nil { //未登录用户
		user = 0 // 后面在考虑用什么好
	} else {
		user = int(loginuser.UserId)
	}

	//vi,_:=redis.String(models.GetRec().Do("Get", "video"))
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "success"},
		VideoList: CreatVideoList(user),
		NextTime:  latestTime,
	})
}

// 下面的改完bug之后要放到services

func CreatVideoList(user int) (videolist []Video) {
	var videoret Video
	videos := services.GetJsonFeeCache()
	//redis缓存没有时未完成
	for _, singlevideo := range videos {
		videoret.Id = singlevideo.ID
		videoret.CoverUrl = singlevideo.CoverUrl
		videoret.PlayUrl = singlevideo.PlayUrl
		videoret.CommentCount = singlevideo.CommentCount
		videoret.FavoriteCount = singlevideo.FavoriteCount
		videoret.Author = getAuthor(int(singlevideo.UserId)) //TODO
		videoret.IsFavorite = services.GetFavoriteService().FavoriteJudge(user, int(singlevideo.ID))
		videolist = append(videolist, videoret)
	}
	return videolist
}

func getAuthor(id int) (Author User) {
	getuser, err := services.GetUserService().UserInfo(id)
	if err != nil {
		fmt.Println("get authors failed,err: ", err.Error())
	}
	Author.Id = getuser.Id
	Author.Name = getuser.Name
	Author.FollowCount = getuser.FollowCount
	Author.FollowerCount = getuser.FollowerCount
	Author.IsFollow = false // 要查表，未完成
	return Author
}

//func judge(userId, videoId int) bool {
//	is, _ := daos.GetFavoriteDao().JudgeIsFavorite(userId, videoId)
//
//	return is
//}

//var DemoUser = User{
//	Id:            1,
//	Name:          "TestUser",
//	FollowCount:   0,
//	FollowerCount: 0,
//	IsFollow:      false,
//}
//var DemoVideos = []Video{
//	{
//		Id:            1,
//		Author:        DemoUser,
//		PlayUrl:       "https://www.w3schools.com/html/movie.mp4",
//		CoverUrl:      "https://cdn.pixabay.com/photo/2016/03/27/18/10/bear-1283347_1280.jpg",
//		FavoriteCount: 0,
//		CommentCount:  0,
//		IsFavorite:    false,
//	},
//}
