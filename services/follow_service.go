package services

import (
	"douyin/daos"
	"douyin/models"
	"errors"
	"strconv"
	"sync"
)

var (
	followService     FollowService
	followServiceOnce sync.Once
)

type FollowService interface {
	Action(userId string, toUserId string, actionType string) error
}
type FollowServiceImpl struct {
	followDao daos.FollowDao
}

func GetFollowService() FollowService {
	followServiceOnce.Do(func() {
		followService = &FollowServiceImpl{
			followDao: daos.GetFollowDao(),
		}
	})
	return followService
}

func (f *FollowServiceImpl) Action(userId string, toUserId string, actionType string) error {
	//关注发起者:userId,被关注者:toUserId
	uid, err := strconv.Atoi(userId)
	if err != nil {
		return err
	}
	tuid, err := strconv.Atoi(toUserId)
	if err != nil {
		return err
	}
	follow := &models.Follow{FollowerId: uid, FollowId: tuid}
	switch actionType {
	case "1":
		{
			_, err = f.followDao.FindFollow(toUserId, userId)
			if err == nil {
				return errors.New("follow exist")

			}
			err = f.followDao.AddFollow(follow)
			if err != nil {
				return err
			}
			//清除redis的userId_toUserId关注操作缓存
			//...待补充
		}
	case "2":
		err = f.followDao.DelFollow(follow)
		if err != nil {
			return err
		}
		//清除redis的userId_toUserId取关操作缓存
		//...待补充
	}
	return nil
}
