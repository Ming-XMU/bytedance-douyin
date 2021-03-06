package controller

import (
	"douyin/models"
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
	UserList []models.UserMessage `json:"user_list"`
}

// RelationAction @author cwh
// RelationAction no practical effect, just check if token is valid
func RelationAction(c *gin.Context) {
	//token验证
	token := c.Query("token")
	user, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "请先登录"})
		return
	}

	//请求没有携带userid参数，改为从token获取
	userId := strconv.FormatInt(user.UserId, 10)

	//用户存在验证
	toUserId := c.Query("to_user_id")
	_, err = services.GetUserService().UserInfo(toUserId)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "对应用户不存在！"})
		return
	}
	//不允许关注自己
	if userId == toUserId {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "不能关注自己！"})
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

// FollowList @author cwh
//获取关注列表用户
// FollowList all users have same follow list
func FollowList(c *gin.Context) {
	//获取token
	token := c.Query("token")
	_, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "请先登录"})
		return
	}
	//交给service层查询关注列表
	list, err := services.GetFollowService().UserFollowList(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "查询错误！"})
		return
	}
	//正常返回
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "查询成功",
		},
		UserList: list,
	})
}

// FollowerList @author cwh
//获取用户的粉丝列表
//all users have same follower list
func FollowerList(c *gin.Context) {
	//获取token
	token := c.Query("token")
	_, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "请先登录"})
		return
	}
	//交给service层查询关注列表
	list, err := services.GetFollowService().UserFollowerList(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "查询错误！"})
		return
	}
	//正常返回
	c.JSON(http.StatusOK, UserListResponse{
		Response: Response{
			StatusCode: 0,
			StatusMsg:  "查询成功",
		},
		UserList: list,
	})
}
