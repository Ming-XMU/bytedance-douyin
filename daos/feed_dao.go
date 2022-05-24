package daos

import (
	"douyin/models"
	"gorm.io/gorm"
	"sync"
)

/**
 * @Author: Ember
 * @Date: 2022/5/16 9:43
 * @Description: TODO
 **/

type FeedDao interface {
	//create feed record
	CreateFeed(video *models.Video) (rowsAffected int64, err error)
	//update video favourite count
	UpdateVideoFavoriteCount(videoId int64,count int)error
}
type FeedDaoImpl struct {
	db *gorm.DB
}

func (f *FeedDaoImpl) CreateFeed(video *models.Video) (rowsAffected int64, err error) {
	result := f.db.Create(video)
	return result.RowsAffected, result.Error
}
func (f *FeedDaoImpl)UpdateVideoFavoriteCount(videoId int64,count int)error{
	return f.db.Model(&models.Video{}).Where("id = ?", videoId).Update("favourite_count", count).Error
}
//single create
var (
	feeDao     FeedDao
	feeDaoOnce sync.Once
)

func GetFeeDao() FeedDao {
	feeDaoOnce.Do(func() {
		feeDao = &FeedDaoImpl{
			db: models.GetDB(),
		}
	})
	return feeDao
}
