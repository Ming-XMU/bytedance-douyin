package daos

import (
	"douyin/models"
	"sync"
)

/**
 * @Author: Ember
 * @Date: 2022/5/16 9:43
 * @Description: TODO
 **/

type FeedDao interface{
	//create feed record
	CreateFeed(video models.Video) (rowsAffected int64,err error)
}
type FeedDaoImpl struct{
}
func(f *FeedDaoImpl)CreateFeed(video models.Video) (rowsAffected int64,err error){
	result := models.GetDB().Create(&video)
	return result.RowsAffected,result.Error
}

//single create
var(
	feeDao FeedDao
	feeDaoOnce sync.Once
)
func GetFeeDao() FeedDao{
	feeDaoOnce.Do(func() {
		feeDao = &FeedDaoImpl{
		}
	})
	return feeDao
}
