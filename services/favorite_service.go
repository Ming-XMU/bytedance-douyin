package services

import (
	"douyin/daos"
	"douyin/models"
	"errors"
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
	if action == 1 {
		favorite := &models.Favorite{
			UserId:  userId,
			VideoId: videoId,
		}
		return f.favoriteDao.InsertFavorite(favorite)
	} else if action == 2 {
		return f.favoriteDao.DeleteFavorite(userId, videoId)
	}
	return errors.New("action is error")
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
