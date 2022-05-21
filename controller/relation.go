package controller

import (
	"douyin/mq"
	"douyin/services"
	"douyin/tools"
	"fmt"
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
	//参数验证，to_user_id是否存在，type值是否合法
	toUserId, err := strconv.Atoi(c.Query("to _user_id"))
	actionType := c.Query("action_type")
	if (actionType != "1" && actionType != "2") || err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "操作检验错误！"})
		return
	}
	//用户存在验证
	_, err = services.GetUserService().UserInfo(toUserId)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "对应用户不存在！"})
		return
	}
	//redis处理缓存关注数量，hincrby cacheHashWriter to_user_id 1 or -1
	if actionType == "2" {
		actionType = "-1"
	}
	err = tools.RedisDo("HINCRBY", services.GetFollowWrite(), toUserId, actionType)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "系统错误!请稍后重试！"})
		return
	}
	//redis缓存关注，hset user_id to_user_id actionType
	err = tools.RedisDo("HSET", userId, toUserId, actionType)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "系统错误!请稍后重试！"})
		//处理之前加入的缓存，hdecrby cacheHashWrite to_user_id 1 or -1
		err := tools.RedisDo("HDECRBY", services.GetFollowWrite(), toUserId, actionType)
		if err != nil {
			fmt.Println(userId + "在redis中多了一个“意外的”关注")
		}
		return
	}
	//mq信息处理
	rabbitmq := mq.GetFollowMQ()
	msg := strings.Join([]string{userId, strconv.Itoa(toUserId), actionType}, "_")
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
