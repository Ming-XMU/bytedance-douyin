package controller

import (
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
	if tools.VeifyToken(token) != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "请先登录！"},
		})
		return
	}
	//参数处理，3个参数转为int，检查videoId是否存在
	userId, e1 := strconv.Atoi(c.Query("user_id"))
	videoId, e2 := strconv.Atoi(c.Query("video_id"))
	exist := services.GetVideoService().VideoExist(videoId)
	actionType, e3 := strconv.Atoi(c.Query("action_type"))
	if e1 != nil || e2 != nil || e3 != nil || (actionType != 1 && actionType != 2) || !exist {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "参数错误！"})
		return
	}
	//调用service进行操作
	err := services.GetFavoriteService().FavoriteAction(userId, videoId, actionType)
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
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: DemoVideos,
	})
}
