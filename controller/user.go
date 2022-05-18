package controller

import (
	"douyin/services"
	"douyin/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync/atomic"
)

var (
	UserSerivce    services.UserService
	userIdSequence int64
)

type UserLoginResponse struct {
	Response
	UserId int64  `json:"user_id,omitempty"`
	Token  string `json:"token"`
}

type UserResponse struct {
	Response
	User User `json:"user"`
}

//注册功能
//失败返回错误信息
func Register(c *gin.Context) {
	//随机生成salt
	salt := tools.RandomStringUtil()
	username := c.Query("username")
	//密码MD5加密
	password := tools.Md5Util(c.Query("password"), salt)
	//更新用户ID
	userIdSequence = services.GetUserService().FindLastUserId()
	//注册用户
	if err := services.GetUserService().UserRegist(username, password, userIdSequence, salt); err != nil {
		//注册失败返回错误信息
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: err.Error()},
		})
	} else {
		//成功注册
		atomic.AddInt64(&userIdSequence, 1)
		user, err2 := services.GetUserService().UserInfo(int(userIdSequence))
		if err2 != nil {
			return
		}
		token, err2 := tools.CreateToken(user)
		if err2 != nil {
			return
		}
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0, StatusMsg: "regist success"},
			UserId:   userIdSequence,
			Token:    token,
		})
	}
}

// Login 登录接口
// username: 用户名  password:密码
func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	//登录验证失败
	//返回：msg:user does not exist | password error
	if user, err := services.GetUserService().UserLogin(username, password); err != nil {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: err.Error()},
		})
	} else {
		//登陆成功
		//更新令牌
		token, err := tools.CreateToken(user)
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 0, StatusMsg: "login success"},
			UserId:   user.Id,
			Token:    token,
		})
	}
}

//@author cwh
//根据id获取用户信息
//缺失token验证
//缺失是否关注的查询
func UserInfo(c *gin.Context) {
	id, parse := strconv.Atoi(c.Query("user_id"))
	if parse != nil {
		fmt.Println("??")
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "参数类型错误！"},
		})
		return
	}
	if user, err := services.GetUserService().UserInfo(id); err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: err.Error()},
		})
	} else {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "获取用户信息成功！"},
			User: User{
				Id:            user.Id,
				Name:          user.Name,
				FollowCount:   user.FollowCount,
				FollowerCount: user.FollowerCount,
				IsFollow:      false,
			},
		})
	}

}
