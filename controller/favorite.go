package controller

import (
	"douyin/models"
	"douyin/services"
	"douyin/tools"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

//@author cwh
// FavoriteAction no practical effect, just check if token is valid--点赞
func FavoriteAction(c *gin.Context) {
	//权限鉴定
	token := c.Query("token")
	loginInfo, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "请先登录！"},
		})
		return
	}
	//参数处理，3个参数转为int，检查videoId是否存在
	videoId, e2 := strconv.ParseInt(c.Query("video_id"), 10, 64)
	exist := services.GetVideoService().VideoExist(videoId)
	actionType, e3 := strconv.Atoi(c.Query("action_type"))
	if e2 != nil || e3 != nil || (actionType != 1 && actionType != 2) || !exist {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "参数错误！"})
		return
	}
	//调用service进行操作
	err = services.GetFavoriteService().FavoriteAction(loginInfo.UserId, videoId, actionType)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, Response{
			StatusCode: 0,
			StatusMsg:  "操作成功",
		})
	}
}

// FavoriteList all users have same favorite video list
func FavoriteList(c *gin.Context) {
	token := c.Query("token")
	loginUser, err := tools.VeifyToken(token)
	if err != nil{
		c.JSON(http.StatusOK,Response{
			StatusCode: -1,
			StatusMsg: err.Error(),
		})
	}
	list, err := services.GetFavoriteService().GetUserFavoriteVideoList(loginUser.UserId)
	if err != nil{
		c.JSON(http.StatusOK,Response{
			StatusCode: -1,
			StatusMsg: err.Error(),
		})
	}
	//vo to dto
	videoList := make([]Video,len(list))
	//get relation ship
	for _,favorite := range(list){
		video := Video{
			Id: favorite.Video.ID,
			PlayUrl: favorite.Video.PlayUrl,
			CoverUrl: favorite.Video.CoverUrl,
			FavoriteCount: favorite.Video.FavoriteCount,
			CommentCount: favorite.Video.CommentCount,
			IsFavorite: true,
			Author: models.UserMessage{
				Id: favorite.UserId,
				Name: favorite.Author.Name,
				FollowCount: favorite.Author.FollowCount,
				FollowerCount: favorite.Author.FollowerCount,
			},
		}
		videoList = append(videoList,video)
	}
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: videoList,
	})
}
