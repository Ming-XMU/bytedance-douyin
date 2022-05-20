package daos

import (
	"douyin/models"
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
	JudgeIsFollow(followId, followerId int) (is bool, err error)
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
	err = f.db.Debug().Where("follow_id=?&&follower_id=?", followId, followerId).Take(&exist).Error
	if exist.FollowId == 0 {
		return false, err
	}
	return true, err
}
