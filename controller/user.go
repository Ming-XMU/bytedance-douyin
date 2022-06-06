package controller

import (
	"douyin/models"
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
	User models.UserMessage `json:"user"`
}

// Register 注册功能
//失败返回错误信息
func Register(c *gin.Context) {
	//随机生成salt
	salt := tools.RandomStringUtil()
	username := c.Query("username")
	//密码MD5加密
	password := tools.Md5Util(c.Query("password"), salt)
	pwd := c.Query("password")
	//判断注册信息长度
	if len(username) > 32 || len(pwd) > 32 {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "用户注册信息过长"},
		})
		return
	}
	if len(username) == 0 {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "账号为空"},
		})
		return
	}
	if len(pwd) == 0 {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "密码为空"},
		})
		return
	}
	//判断用户名是否违规
	//err := tools.Init("./config/sensitive_word.txt")
	//if err != nil {
	//	c.JSON(http.StatusOK, UserLoginResponse{
	//		Response: Response{StatusCode: 1, StatusMsg: "敏感词库文件加载失败"},
	//	})
	//	return
	//}
	_, isReplaced := tools.Replace(username, "*")
	if isReplaced == true {
		c.JSON(http.StatusOK, UserLoginResponse{
			Response: Response{StatusCode: 1, StatusMsg: "用户名存在违规字符!"},
		})
		return
	}
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
		user, err2 := services.GetUserService().UserInfo(strconv.FormatInt(userIdSequence, 10))
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
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: 0, StatusMsg: "login fail"},
			})
		} else {
			c.JSON(http.StatusOK, UserLoginResponse{
				Response: Response{StatusCode: 0, StatusMsg: "login success"},
				UserId:   user.Id,
				Token:    token,
			})
		}
	}
}

// UserInfo @author cwh
//根据id获取用户信息
//缺失是否关注的查询
func UserInfo(c *gin.Context) {
	token, err := tools.VeifyToken(c.Query("token"))
	if err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: "请先登录"},
		})
		return
	}
	//id获取
	id := c.Query("user_id")
	//调用service进行查询
	if user, err := services.GetUserService().UserInfo(id); err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 1, StatusMsg: err.Error()},
		})
	} else {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "获取用户信息成功！"},
			User:     *services.GetFollowService().UserFollowInfo(user, strconv.FormatInt(token.UserId, 10)),
		})
	}
}

func UpdateUser(c *gin.Context) {
	err := services.GetUserService().UserUpdate(c)
	if err != nil {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: err.Error()})
	} else {
		c.JSON(http.StatusOK, Response{StatusCode: 0, StatusMsg: "上传成功"})
	}
}
