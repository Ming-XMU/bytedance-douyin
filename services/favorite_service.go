package services

import (
	"douyin/daos"
	"douyin/models"
	mq "douyin/mq"
	"douyin/tools"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var (
	favoriteService     FavoriteService
	favoriteServiceOnce sync.Once
)

func GetFavoriteService() FavoriteService {
	favoriteServiceOnce.Do(func() {
		favoriteService = &FavoriteServiceImpl{
			favoriteDao: daos.GetFavoriteDao(),
		}
	})
	return favoriteService
}

type FavoriteService interface {
	FavoriteAction(userId int64, videoId int64, acton int) error
	FavoriteJudge(userId, videoId int) bool
}

type FavoriteServiceImpl struct {
	favoriteDao daos.FavoriteDao
}

//@author cwh
//@userId 用户id
//@videoId 视频id
//@action 1点赞，2取消
func (f *FavoriteServiceImpl) FavoriteAction(userId int64, videoId int64, action int) error {
	favorite := &models.Favorite{
		UserId:  userId,
		VideoId: videoId,
	}
	//prepare msg to mq
	favoriteAction := &mq.FavoriteActionMsg{
		Favorite: favorite,
		Action:   action,
	}
	jsonMsg, err := json.Marshal(favoriteAction)
	if err != nil {
		return err
	}
	//cache
	if action == 1 {
		result, err := tools.RedisCacheFavorite(favorite)
		if err != nil {
			return err
		}
		if result == 2 {
			return errors.New("已经点赞过了")
		}
	} else if action == 2 {
		result, err := tools.RedisCacheCancelFavorite(favorite)
		if err != nil {
			fmt.Println(err)
			return err
		}
		if result == 2 {
			return errors.New("还没有进行点赞")
		}
	}

	rabbitMQSimple := mq.NewRabbitMQSimple("favoriteActionQueue", "amqp://admin:123456@120.78.238.68:5672/default_host")
	err = rabbitMQSimple.PublishSimple(string(jsonMsg))
	if err != nil {
		//Roll Back
		if action == 1 {
			tools.RedisCacheCancelFavorite(favorite)
		} else if action == 2 {
			tools.RedisCacheFavorite(favorite)
		}
		return err
	}
	return err
}

// FavoriteJudge 判断是否有点赞
// @author wechan
func (f *FavoriteServiceImpl) FavoriteJudge(userId, videoId int) bool {
	if userId == 0 {
		return false //未登录用户，直接返回false
	}
	is, err := f.favoriteDao.JudgeIsFavorite(userId, videoId)
	if err != nil {
		return false
	}
	return is
}
