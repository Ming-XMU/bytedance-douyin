package services

import (
	"douyin/daos"
	"douyin/models"
	"douyin/tools"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

func GetUserService(db *gorm.DB) *UserService {
	return &UserService{
		UserDao: daos.GetUserDao(db),
	}
}

type UserService struct {
	UserDao *daos.UserDao
}

func (u *UserService) UserLogin(username string, password string) (*models.User, error) {
	user, err := u.UserDao.FindByName(username)
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

//注册用户
//1.先判断表里有没有用户 如果有就提示用户存在
//2.判断用户名是否违法或者合规（暂未实现）
//3.注册用户
func (u *UserService) UserRegist(username string, password string, userId int64, salt string) error {
	//判断用户是否已经注册
	_, err := u.UserDao.FindByName(username)
	if err == nil {
		return errors.New("user does not exist")
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
	}
	e := u.UserDao.AddUser(&user)
	if e != nil {
		return errors.New("user regist failed")
	}
	return nil
}
