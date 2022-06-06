package services

import (
	"douyin/daos"
	"douyin/models"
	"douyin/tools"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"sync"
)

var (
	userService     UserService
	userServiceOnce sync.Once
)

const (
	DEAULT_AVATAR     = "https://static-1304359512.cos.ap-guangzhou.myqcloud.com/img/default_avatar.jpg"
	DEAULT_BACKGROUND = "https://static-1304359512.cos.ap-guangzhou.myqcloud.com/img/default_background.jpg"
	IMG_Url = "https://static-1304359512.cos.accelerate.myqcloud.com/img/"
)


type UserService interface {
	UserLogin(username string, password string) (*models.User, error)
	UserRegist(username string, password string, userId int64, salt string) error
	UserInfo(id string) (*models.User, error)
	FindLastUserId() int64
	UserUpdate(c *gin.Context) error
}
type UserServiceImpl struct {
	userDao daos.UserDao
}

func GetUserService() UserService {
	userServiceOnce.Do(func() {
		userService = &UserServiceImpl{
			userDao: daos.GetUserDao(),
		}
	})
	return userService
}

func (u *UserServiceImpl) UserLogin(username string, password string) (*models.User, error) {
	user, err := u.userDao.FindByName(username)
	if err != nil {
		return nil, errors.New("user does not exist")
	}
	//校验密码
	fmt.Println(user)
	//MD5加密验证
	password = tools.Md5Util(password, user.Salt)
	if password != user.Pwd {
		return nil, errors.New("password error")
	}
	return user, nil
}

// UserRegist 注册用户
//1.先判断表里有没有用户 如果有就提示用户存在
//2.判断用户名是否违法或者合规（暂未实现）
//3.注册用户
func (u *UserServiceImpl) UserRegist(username string, password string, userId int64, salt string) error {
	//判断用户是否已经注册
	_, err := u.userDao.FindByName(username)
	if err == nil {
		return errors.New("user exist")
	}
	//加入判断用户名是否合规的方法(未实现)
	//添加用户
	user := models.User{
		Id:            userId,
		Name:          username,
		Pwd:           password,
		Salt:          salt,
		FollowCount:   0,
		FollowerCount: 0,
		Avatar: DEAULT_AVATAR,
		BackGround: DEAULT_BACKGROUND,
	}
	e := u.userDao.AddUser(&user)
	if e != nil {
		return errors.New("user regist failed")
	}
	return nil
}

// FindLastUserId 返回当前最大的用户ID
func (u *UserServiceImpl) FindLastUserId() int64 {
	return u.userDao.LastId()
}

// UserInfo @author cwh
//根据id查询对应的对象
func (u *UserServiceImpl) UserInfo(userId string) (*models.User, error) {
	id, err := strconv.Atoi(userId)
	if err != nil {
		return nil, err
	}
	return u.userDao.FindById(id)
}

// UserUpdate
// @author zia
// @Description: 用户更新
// @receiver u
// @param c
// @return error
func (u *UserServiceImpl) UserUpdate(c *gin.Context) error {
	token, _ := c.GetPostForm("token")
	loginUser, err := tools.VeifyToken(token)
	if loginUser == nil || err != nil {
		return errors.New("请先登录")
	}
	user := models.User{Id: loginUser.UserId}
	avatar, _ := c.FormFile("avatar")
	background, _ := c.FormFile("back_ground")
	//上传头像
	if avatar != nil {
		file, err1 := avatar.Open()
		if err1 != nil {
			return errors.New("avatar upload fail")
		}
		err1 = tools.UploadFileObjectToCos(file, avatar.Filename, "img", "image/jpeg")
		if err1 != nil {
			return errors.New("avatar upload fail")
		}
		//记录user的头像地址
		user.Avatar = IMG_Url + avatar.Filename
	}
	//上传背景图
	if background != nil {
		file, err1 := background.Open()
		if err1 != nil {
			return errors.New("background upload fail")
		}
		err1 = tools.UploadFileObjectToCos(file, background.Filename, "img", "image/jpeg")
		if err1 != nil {
			return errors.New("background upload fail")
		}
		user.BackGround = IMG_Url + background.Filename
	}
	//更新数据
	err = u.userDao.UpdateUser(&user)
	if err != nil {
		return errors.New("更新失败")
	}
	return nil
}
