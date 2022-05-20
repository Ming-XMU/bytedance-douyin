package daos

import (
	"douyin/models"
	"errors"
	"gorm.io/gorm"
	"sync"
)

var (
	followDao     FollowDao
	followDaoOnce sync.Once
)

type FollowDao interface {
	AddFollow(follow *models.Follow) error
	DelFollow(follow *models.Follow) error
	FindFollow(followId string, followerId string) (*models.Follow, error)
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

// AddFollow
// @author zia
// @Description: 添加关注记录
// @receiver f
// @param follow
// @return error
func (f *FollowDaoImpl) AddFollow(follow *models.Follow) error {
	return f.db.Debug().Create(follow).Error
}

// DelFollow
// @author zia
// @Description: 删除关注记录
// @receiver f
// @param follow
// @return error
func (f *FollowDaoImpl) DelFollow(follow *models.Follow) error {
	return f.db.Debug().Where("follow_id = ? && follower_id = ?", follow.FollowId, follow.FollowerId).Delete(&models.Follow{}).Error
}

func (f *FollowDaoImpl) FindFollow(followId string, followerId string) (*models.Follow, error) {
	var follow models.Follow
	if err := f.db.Where("follow_id = ? && follower_id = ?", followId, followerId).First(&follow).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return &follow, nil
}
