package daos

import (
	"douyin/models"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
	"sync"
)

var (
	userDao     UserDao
	userDaoOnce sync.Once
)

type UserDao interface {
	AddUser(user *models.User) error
	FindByName(name string) (*models.User, error)
	LastId() int64
	FindById(id int) (*models.User, error)
}
type UserDaoImpl struct {
	db  *gorm.DB
	rec redis.Conn
}

func GetUserDao() UserDao {
	userDaoOnce.Do(func() {
		userDao = &UserDaoImpl{
			db:  models.GetDB(),
			rec: models.GetRec(),
		}
	})
	return userDao
}

// AddUser 添加用户
// 参数 user User结构体指针
func (u *UserDaoImpl) AddUser(user *models.User) error {
	if err := u.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// FindByName 根据用户名查找用户
// 参数 name string类型 用户名
func (u *UserDaoImpl) FindByName(name string) (*models.User, error) {
	var user models.User
	if err := u.db.Where("name = ?", name).First(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &user, nil
}

//通过主键查询最后一条记录
//返回当前表内的最大ID
func (u *UserDaoImpl) LastId() int64 {
	fmt.Println("yes is dao")
	var user models.User
	if err := u.db.Last(&user).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		//表内没有数据默认为1
		return 1
	}
	return user.Id
}

//@author cwh
//根据id查询对应的user
func (u *UserDaoImpl) FindById(id int) (*models.User, error) {
	var user models.User
	err := u.db.Debug().Select("id", "name", "follow_count", "follower_count").Where("id = ?", id).Take(&user).Error
	fmt.Println(user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
