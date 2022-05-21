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
	toUserId, err := strconv.Atoi(c.Query("to_user_id"))
	_, err = services.GetUserService().UserInfo(toUserId)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "对应用户不存在！"})
		return
	}
	//缓存查询，不存在便加载
	err = services.GetFollowService().FollowListCdRedis(int(user.UserId))
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "缓存错误！"})
		return
	}

	actionType := c.Query("action_type")
	var action string
	var add int
	if actionType == "1" {
		//关注操作，redis缓存列表添加toUserId，对应toUserId一个关注数
		action = "SADD"
		add = -1
	} else if actionType == "2" {
		//取关操作，redis缓存列表添加toUserId，对应toUserId一个关注数
		action = "SREM"
		add = -1
	} else {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "操作类型错误！"})
		return
	}
	//sadd userId touserId
	if tools.RedisDoKV(action, userId, toUserId) != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "系统错误！，请稍后重试"})
		return
	}
	//hincrby cacheHashWrite toUserId 1
	_ = tools.RedisDoHash("HINCRBY", services.GetFollowWrite(), toUserId, add)
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
