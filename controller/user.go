package controller

import (
	"douyin/services"
	"douyin/tools"
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

// usersLoginInfo use map to store user info, and key is username+password for demo
// user data will be cleared every time the server starts
// test data: username=zhanglei, password=douyin
var usersLoginInfo = map[string]User{
	"root123456": {
		Id:            1,
		Name:          "root",
		FollowCount:   10,
		FollowerCount: 5,
		IsFollow:      true,
	},
}

// Register 注册功能
//失败返回错误信息
func Register(c *gin.Context) {
	//随机生成salt
	salt := tools.RandomStringUtil()
	username := c.Query("username")
	//密码MD5加密
	password := tools.Md5Util(c.Query("password"), salt)
	//更新用户ID
	userIdSequence = services.GetUserService().FindLastUserId()
	atomic.AddInt64(&userIdSequence, 1)

	//注册用户
	if err := services.GetUserService().UserRegist(username, password, userIdSequence, salt); err != nil {
		//注册失败返回错误信息
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: err.Error()},
		})
	} else {
		//成功注册
		user, err2 := services.GetUserService().UserInfo(int(userIdSequence))
		if err2 != nil {
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: 1, StatusMsg: err2.Error()},
			})
		}
		token, err3 := tools.CreateToken(user)
		if err3 != nil {
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: 1, StatusMsg: "create token failed"},
			})
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
//缺失是否关注的查询
func UserInfo(c *gin.Context) {
	/* 这个接口应该不用鉴权，先注释掉
	token := c.Query("token")
	if tools.VeifyToken(token) != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "请先登录！"},
		})
		return
	}*/
	//转换id类型
	id, parse := strconv.Atoi(c.Query("user_id"))
	if parse != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "参数类型错误！"},
		})
		return
	}
	//调用service进行查询
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
