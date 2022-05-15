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
