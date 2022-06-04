package services

import (
	"douyin/daos"
	"douyin/models"
	mq "douyin/mq"
	"douyin/tools"
	"encoding/json"
	"errors"
	"github.com/anqiansong/ketty/console"
	"github.com/sirupsen/logrus"
	"strconv"
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
	// GetUserFavoriteVideoList Get User Favorite Video List
	GetUserFavoriteVideoList(videoId int64) (list []models.VideoVo, err error)
}

type FavoriteServiceImpl struct {
	favoriteDao daos.FavoriteDao
}

// FavoriteAction @author cwh
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
			console.Error(err)
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
			_, err := tools.RedisCacheCancelFavorite(favorite)
			if err != nil {
				logrus.Errorln("cache cancel favorite is false by", favorite)
			}
		} else if action == 2 {
			_, err := tools.RedisCacheFavorite(favorite)
			if err != nil {
				logrus.Errorln("cache favorite is false by", favorite)
			}
		}
		console.Error(err)
		return err
	}
	return err
}

func (f *FavoriteServiceImpl) GetUserFavoriteVideoList(userId int64) (list []models.VideoVo, err error) {
	favoriteList, err := f.GetUserFavoriteVideoListFromDB(userId)
	if err != nil {
		console.Error(err)
		return
	}
	//get follow list
	followList, err := GetFollowService().UserFollowList(strconv.FormatInt(userId, 10))
	if err != nil {
		console.Error(err)
		return
	}
	//change map
	m := make(map[int64]struct{}, len(followList))
	list = make([]models.VideoVo, len(followList))
	for _, follow := range followList {
		m[follow.Id] = struct{}{}
	}
	//dto to vo
	for _, favorite := range favoriteList {
		_, ok := m[favorite.Author.Id]
		list = append(list, models.VideoVo{
			Id: favorite.VideoId,
			Author: models.UserMessage{
				Id:            favorite.Author.Id,
				Name:          favorite.Author.Name,
				FollowerCount: favorite.Author.FollowerCount,
				FollowCount:   favorite.Author.FollowCount,
				IsFollow:      ok,
			},
			PlayUrl:       favorite.Video.PlayUrl,
			CoverUrl:      favorite.Video.CoverUrl,
			FavoriteCount: favorite.Video.FavoriteCount,
			CommentCount:  favorite.Video.CommentCount,
			IsFavorite:    true,
		})
	}
	//favorite
	return
}

// GetUserFavoriteVideoListFromDB get favorite video from db
func (f *FavoriteServiceImpl) GetUserFavoriteVideoListFromDB(userId int64) (list []models.FavoriteList, err error) {
	favorites, err := f.favoriteDao.UserFavorites(userId)
	if err != nil {
		console.Error(err)
		return
	}
	return favorites, nil
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
