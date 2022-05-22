package controller

import (
	"douyin/mq"
	"douyin/services"
	"douyin/tools"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

var (
	FollowSerivce services.FollowService
)

type UserListResponse struct {
	Response
	UserList []User `json:"user_list"`
}

//@author cwh
// RelationAction no practical effect, just check if token is valid
func RelationAction(c *gin.Context) {
	//token验证
	token := c.Query("token")
	user, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "前先登录！"})
		return
	}

	//参数userId跟token一致
	userId := c.Query("user_id")
	if strconv.FormatInt(user.UserId, 10) != userId {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "用户检验错误！"})
		return
	}

	//用户存在验证
	toUserId := c.Query("to_user_id")
	_, err = services.GetUserService().UserInfo(toUserId)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "对应用户不存在！"})
		return
	}

	//redis缓存处理
	actionType := c.Query("action_type")
	redisErr := services.GetFollowService().RedisAction(userId, toUserId, actionType)
	if redisErr != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: redisErr.Error()})
		return
	}
	//mq信息处理
	rabbitmq := mq.GetFollowMQ()
	msg := strings.Join([]string{userId, toUserId, actionType}, "_")
	err = rabbitmq.PublishSimple(msg)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "action fail"})
	} else {
		c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "action success"})
	}
}

//@author cwh
//获取关注列表
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
