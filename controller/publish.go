package controller

import (
	"douyin/models"
	"douyin/services"
	"github.com/anqiansong/ketty/console"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var (
	FeeSerivce services.FeedService
)

type VideoListResponse struct {
	Response
	VideoList []models.VideoVo `json:"video_list"`
}

// Publish check token then save upload file to public directory
func Publish(c *gin.Context) {
	err := FeeSerivce.PublishAction(c)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  " uploaded successfully",
	})
}

// PublishList all users have same public video list
func PublishList(c *gin.Context) {
	userId := c.Query("user_id")
	parseInt, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		console.Error(err)
	}
	videos, err := FeeSerivce.GetUserAllPublishVideos(parseInt)
	c.JSON(http.StatusOK, VideoListResponse{
		Response: Response{
			StatusCode: 0,
		},
		VideoList: videos,
	})
}
