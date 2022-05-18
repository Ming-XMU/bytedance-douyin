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
	DeleteFavorite(userId, videoId int) error
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

func (f *FavoriteDaoImpl) InsertFavorite(favorite *models.Favorite) error {
	return f.db.Debug().Create(favorite).Error
}

func (f *FavoriteDaoImpl) DeleteFavorite(userId, videoId int) error {
	return f.db.Debug().Where("user_id = ? && video_id = ?", userId, videoId).Delete(&models.Favorite{}).Error
}
