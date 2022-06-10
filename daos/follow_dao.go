package daos

import (
	"douyin/models"
	"errors"
	"gorm.io/gorm"
	"sync"
)

//follow_dao
//Author: Wechan
//Description：关注和被关注的DAO操作

var (
	followDao     FollowDao
	followDaoOnce sync.Once
)

type FollowDao interface {
	AddFollow(follow *models.Follow) error
	DelFollow(follow *models.Follow) error
	FindFollow(followId string, followerId string) (*models.Follow, error)
	JudgeIsFollow(followId, followerId int) (is bool, err error)
	UserFollow(userId int) ([]int, error)
	UserFollower(userId int) ([]int, error)
	UpdateUserFollowCount(userId int, followCount int) error
	UpdateUserFollowerCount(userId int, followerCount int) error
}

type FollowDaoImpl struct {
	db *gorm.DB
}

func GetFollowDao() FollowDao {
	followDaoOnce.Do(func() {
		followDao = &FollowDaoImpl{
			db: models.GetDB(),
		}
	})
	return followDao
}

// JudgeIsFollow 判断是否有关注
func (f *FollowDaoImpl) JudgeIsFollow(followId, followerId int) (is bool, err error) {
	var exist models.Follow
	err = f.db.Where("follow_id=?&&follower_id=?", followId, followerId).Take(&exist).Error
	if exist.FollowId == 0 {
		return false, err
	}
	return true, err
}

// AddFollow
// @author zia
// @Description: 添加关注记录
// @receiver f
// @param follow
// @return error
func (f *FollowDaoImpl) AddFollow(follow *models.Follow) error {
	return f.db.Create(follow).Error
}

// DelFollow
// @author zia
// @Description: 删除关注记录
// @receiver f
// @param follow
// @return error
func (f *FollowDaoImpl) DelFollow(follow *models.Follow) error {
	return f.db.Where("follow_id = ? && follower_id = ?", follow.FollowId, follow.FollowerId).Delete(&models.Follow{}).Error
}

// FindFollow
// @author zia
// @Description: 查找关注记录
// @receiver f
// @param followId
// @param followerId
// @return *models.Follow
// @return error
func (f *FollowDaoImpl) FindFollow(followId string, followerId string) (*models.Follow, error) {
	var follow models.Follow
	if err := f.db.Where("follow_id = ? && follower_id = ?", followId, followerId).First(&follow).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &follow, nil
}

func (f *FollowDaoImpl) UserFollow(userId int) ([]int, error) {
	var res []int
	err := f.db.Table("follow").Select("follow_id").Where("follower_id = ?", userId).Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (f *FollowDaoImpl) UserFollower(userId int) ([]int, error) {
	var res []int
	err := f.db.Table("follow").Select("follower_id").Where("follow_id = ?", userId).Find(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (f *FollowDaoImpl) UpdateUserFollowCount(userId int, followCount int) error {
	return f.db.Model(&models.User{}).Where("id = ?", userId).Update("follow_count", followCount).Error
}

func (f *FollowDaoImpl) UpdateUserFollowerCount(userId int, followerCount int) error {
	return f.db.Model(&models.User{}).Where("id = ?", userId).Update("follower_count", followerCount).Error
}
