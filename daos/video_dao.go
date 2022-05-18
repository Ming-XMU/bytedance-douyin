package daos

import (
	"douyin/models"
	"gorm.io/gorm"
	"sync"
)

var (
	videoDao    VideoDao
	videDaoOnce sync.Once
)

type VideoDao interface {
	FindById(id int) (*models.Video, error)
}

type VideoDaoImpl struct {
	db *gorm.DB
}

func GetVideoDao() VideoDao {
	videDaoOnce.Do(func() {
		videoDao = &VideoDaoImpl{
			db: models.GetDB(),
		}
	})
	return videoDao
}

func (v *VideoDaoImpl) FindById(id int) (*models.Video, error) {
	var video models.Video
	err := v.db.Debug().Select("id", "user_id", "play_url", "cover_url", "comment_count", "favorite_count").
		Where("id = ?", id).Take(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}
