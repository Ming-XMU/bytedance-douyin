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
	CreateFeed(video models.Video) (rowsAffected int64, err error)
}
type FeedDaoImpl struct {
	db *gorm.DB
}

func (f *FeedDaoImpl) CreateFeed(video models.Video) (rowsAffected int64, err error) {
	result := f.db.Create(&video)
	return result.RowsAffected, result.Error
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
