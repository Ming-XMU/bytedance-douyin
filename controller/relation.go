package controller

import (
	"douyin/mq"
	"douyin/services"
	"douyin/tools"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var (
	FollowSerivce services.FollowService
)

type UserListResponse struct {
	Response
	UserList []User `json:"user_list"`
}

// RelationAction no practical effect, just check if token is valid
func RelationAction(c *gin.Context) {
	userId := c.Query("user_id")
	toUserId := c.Query("to_user_id")
	actionType := c.Query("action_type")
	token := c.Query("token")
	_, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "token is expire or empty"})
		return
	}
	//redis缓存关注 | 取关操作
	//redis处理缓存关注 | 取关数量
	rabbitmq := mq.GetFollowMQ()
	msg := strings.Join([]string{userId, toUserId, actionType}, "_")
	err = rabbitmq.PublishSimple(msg)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "action fail"})
	} else {
		c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "action success"})
	}
}

// FollowList all users have same follow list
func FollowList(c *gin.Context) {
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
		},
		UserList: []User{DemoUser},
	})
}

// FollowerList all users have same follower list
func FollowerList(c *gin.Context) {
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
		},
		UserList: []User{DemoUser},
	})
}
