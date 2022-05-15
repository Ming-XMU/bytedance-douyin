package daos

import (
	"douyin/models"
	"errors"
	"gorm.io/gorm"
)

func GetUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

type UserDao struct {
	db *gorm.DB
}

// AddUser 添加用户
// 参数 user User结构体指针
func (u *UserDao) AddUser(user *models.User) error {
	if err := u.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// FindByName 根据用户名查找用户
// 参数 name string类型 用户名
func (u *UserDao) FindByName(name string) (*models.User, error) {
	var user models.User
	if err := u.db.Where("name = ?", name).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &user, nil
}
