package daos

import (
	"douyin/models"
	"gorm.io/gorm"
	"sync"
)

var (
	favoriteDao     FavoriteDao
	favoriteDaoOnce sync.Once
)

type FavoriteDao interface {
	InsertFavorite(favorite *models.Favorite) error
	DeleteFavorite(userId, videoId int64) error
	JudgeIsFavorite(userId, videoId int) (bool, error)
	UserFavorites(userId int64) (lists []models.FavoriteList, err error)
}

type FavoriteDaoImpl struct {
	db *gorm.DB
}

func GetFavoriteDao() FavoriteDao {
	favoriteDaoOnce.Do(func() {
		favoriteDao = &FavoriteDaoImpl{
			db: models.GetDB(),
		}
	})
	return favoriteDao
}

func (f *FavoriteDaoImpl) UserFavorites(userId int64) (lists []models.FavoriteList, err error) {
	var favoriteLists []models.FavoriteList
	err = f.db.Preload("Author").Preload("Video").Where("user_id", userId).Find(&favoriteLists).Error
	lists = favoriteLists
	return
}

func (f *FavoriteDaoImpl) InsertFavorite(favorite *models.Favorite) error {
	return f.db.Create(favorite).Error
}

func (f *FavoriteDaoImpl) DeleteFavorite(userId, videoId int64) error {
	return f.db.Where("user_id = ? && video_id = ?", userId, videoId).Delete(&models.Favorite{}).Error
}

// JudgeIsFavorite 判断是否已经点赞
// Author:wechan
func (f *FavoriteDaoImpl) JudgeIsFavorite(userId, videoId int) (bool, error) {
	var exist models.Favorite
	err := f.db.Where("user_id=?&&video_id=?", userId, videoId).Take(&exist).Error
	if exist.VideoId == 0 {
		return false, err
	}
	return true, err
}
