package controller

import (
	"douyin/services"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []Video `json:"video_list,omitempty"`
	NextTime  int64   `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	var EarliestTime = time.Now()
	//vi,_:=redis.String(models.GetRec().Do("Get", "video"))
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "success"},
		VideoList: CreatVideoList(&EarliestTime),
		NextTime:  EarliestTime.Unix(),
	})
}

func CreatVideoList(EarliestTime *time.Time) (videolist []Video) {
	var videoret Video

	videos := services.GetJsonFeeCache()
	for _, singlevideo := range videos {
		videoret.Id = singlevideo.ID
		videoret.CoverUrl = singlevideo.CoverUrl
		videoret.PlayUrl = singlevideo.PlayUrl
		videoret.CommentCount = singlevideo.CommentCount
		videoret.FavoriteCount = singlevideo.FavoriteCount
		videoret.Author = getAuthor(int(singlevideo.UserId)) //TODO
		videoret.IsFavorite = false                          //TODO
		if EarliestTime.Unix() > singlevideo.CreateAt.Unix() {
			EarliestTime = &singlevideo.CreateAt
		}
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
	Author.IsFollow = false //TODO 要查表，未完成
	return Author
}

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
