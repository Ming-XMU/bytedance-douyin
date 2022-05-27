package controller

import (
	"douyin/models"
	"douyin/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type FeedResponse struct {
	Response
	VideoList []models.VOVideo `json:"video_list,omitempty"`
	NextTime  int64            `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c *gin.Context) {
	var user = 1
	//权限鉴定
	token := c.Query("token")
	loginuser, err := tools.VeifyToken(token)
	if err != nil { //未登录用户
		user = 0
	} else {
		user = int(loginuser.UserId)
	}
	t := c.Query("latest_time")
	latestTime, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		fmt.Println("时间戳转换失败,err:", err.Error())
		latestTime = time.Now().Unix()
	}
	c.JSON(http.StatusOK, FeedResponse{
		Response:  Response{StatusCode: 0, StatusMsg: "success"},
		VideoList: FeeSerivce.CreatVideoList(user),
		NextTime:  latestTime,
	})
}
