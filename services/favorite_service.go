package services

import (
	"douyin/daos"
	"douyin/models"
	mq "douyin/mq"
	"douyin/tools"
	"encoding/json"
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
	FavoriteAction(userId, videoId, acton int) error
	FavoriteJudge(userId, videoId int) bool
}

type FavoriteServiceImpl struct {
	favoriteDao daos.FavoriteDao
}

//@author cwh
//@userId 用户id
//@videoId 视频id
//@action 1点赞，2取消
func (f *FavoriteServiceImpl) FavoriteAction(userId, videoId, action int) error {
	favorite := &models.Favorite{
		UserId:  userId,
		VideoId: videoId,
	}
	//cache
	if action == 1 {
		tools.RedisCacheFavorite(favorite)
	} else if action == 2 {
		tools.RedisCacheCancelFavorite(favorite)
	}
	//send msg to mq
	favoriteAction := &mq.FavoriteActionMsg{
		Favorite: favorite,
		Action: action,
	}
	jsonMsg, err := json.Marshal(favoriteAction)
	if err != nil{
		//TODO Roll Back
		return err
	}
	//TODO TEST
	rabbitMQSimple := mq.NewRabbitMQSimple("favoriteActionQueue", "amqp://admin:admin@192.168.160.134:5672/my_vhost")
	err = rabbitMQSimple.PublishSimple(string(jsonMsg))
	if err != nil{
		//TODO Roll Back
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
