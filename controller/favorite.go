package controller

import (
	"douyin/services"
	"douyin/tools"
	"github.com/anqiansong/ketty/console"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var (
	FavouriteService services.FavoriteService
)

// FavoriteAction @author cwh
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
	err = FavouriteService.FavoriteAction(loginInfo.UserId, videoId, actionType)
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
	//token验证
	_, err := tools.VeifyToken(c.Query("token"))
	if err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: "请先登录"},
		})
		return
	}
	//参数获取
	userId := c.Query("user_id")
	//get user favourite video
	uId, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		console.Error(err)
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		})
	}
	list, err := FavouriteService.GetUserFavoriteVideoList(uId)
	if err != nil {
		console.Error(err)
		c.JSON(http.StatusOK, Response{
			StatusCode: -1,
			StatusMsg:  err.Error(),
		})
	}
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: list,
	})
}
