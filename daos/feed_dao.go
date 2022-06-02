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
	UpdateVideoFavoriteCount(videoId int64, count int) error
	//check multiply video
	FindVideoByPlayUrl(videoUrl string) (rowsAffected int64, err error)
	//get videos by userid
	GetUserVideos(userId int64) (list []*models.Video, err error)
	// GetVideosByCreateAt get videos by create_at
	GetVideosByCreateAt(time int64) (LastVideo []*models.Video, err error)
}
type FeedDaoImpl struct {
	db *gorm.DB
}

func (f *FeedDaoImpl) CreateFeed(video *models.Video) (rowsAffected int64, err error) {
	result := f.db.Create(video)
	return result.RowsAffected, result.Error
}
func (f *FeedDaoImpl) UpdateVideoFavoriteCount(videoId int64, count int) error {
	return f.db.Model(&models.Video{}).Where("id = ?", videoId).Update("favourite_count", count).Error
}

func (f *FeedDaoImpl) FindVideoByPlayUrl(videoUrl string) (rowsAffected int64, err error) {
	var video models.Video
	result := f.db.Where("play_url", videoUrl).First(&video)
	return result.RowsAffected, result.Error
}

func (f *FeedDaoImpl) GetUserVideos(userId int64) (list []*models.Video, err error) {
	var userVideos []*models.Video
	err = f.db.Debug().Where("user_id", userId).Find(&userVideos).Error
	list = userVideos
	return
}
func (f *FeedDaoImpl) GetVideosByCreateAt(time int64) (LastVideo []*models.Video, err error) {
	err = f.db.Debug().Limit(30).Where("create_at<=？", time).Find(&LastVideo).Error
	return LastVideo, err
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
