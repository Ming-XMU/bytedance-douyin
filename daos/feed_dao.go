package daos

import (
	"douyin/models"
	"gorm.io/gorm"
	"sync"
	"time"
)

/**
 * @Author: Ember
 * @Date: 2022/5/16 9:43
 * @Description: TODO
 **/

type FeedDao interface {
	// CreateFeed create feed record
	CreateFeed(video *models.Video) (rowsAffected int64, err error)
	// UpdateVideoFavoriteCount update video favourite count
	UpdateVideoFavoriteCount(videoId int64, count int) error
	// FindVideoByPlayUrl check multiply video
	FindVideoByPlayUrl(videoUrl string) (rowsAffected int64, err error)
	// GetUserVideos get videos by userid
	GetUserVideos(userId int64) (list []*models.Video, err error)
	// GetVideosByCreateAt get videos by create_at
	GetVideosByCreateAt(timestamp int64) (LastVideo []models.Video, err error)
}
type FeedDaoImpl struct {
	db *gorm.DB
}

func (f *FeedDaoImpl) CreateFeed(video *models.Video) (rowsAffected int64, err error) {
	result := f.db.Create(video)
	return result.RowsAffected, result.Error
}
func (f *FeedDaoImpl) UpdateVideoFavoriteCount(videoId int64, count int) error {
	return f.db.Model(&models.Video{}).Where("id = ?", videoId).Update("favorite_count", count).Error
}

func (f *FeedDaoImpl) FindVideoByPlayUrl(videoUrl string) (rowsAffected int64, err error) {
	var video models.Video
	result := f.db.Where("play_url", videoUrl).First(&video)
	return result.RowsAffected, result.Error
}

func (f *FeedDaoImpl) GetUserVideos(userId int64) (list []*models.Video, err error) {
	var userVideos []*models.Video
	err = f.db.Where("user_id", userId).Find(&userVideos).Error
	list = userVideos
	return
}
func (f *FeedDaoImpl) GetVideosByCreateAt(timestamp int64) (LastVideo []models.Video, err error) {
	// go语言固定日期模版
	timeLayout := "2006-01-02 15:04:05"
	// 格式化时间
	// 测试发现接口传过来的时间戳是毫秒
	cur := time.UnixMilli(timestamp).Format(timeLayout)
	err = f.db.Limit(30).Where("create_at<=?", cur).Find(&LastVideo).Error
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
