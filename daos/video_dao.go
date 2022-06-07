package daos

import (
	"douyin/models"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

var (
	videoDao    VideoDao
	videDaoOnce sync.Once
)

type VideoDao interface {
	FindById(id int64) (*models.Video, error)
	UpdateVideCommentCount(videoId int64, add int64) error
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

func (v *VideoDaoImpl) FindById(id int64) (*models.Video, error) {
	var video models.Video
	err := v.db.Debug().Select("id", "user_id", "play_url", "cover_url", "comment_count", "favorite_count", "title").
		Where("id = ?", id).Take(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// UpdateVideCommentCount
// @Description: 更新视频的评论数量
// @receiver v
// @param id
// @param add +1 | -1
// @return error
func (v *VideoDaoImpl) UpdateVideCommentCount(videoId int64, add int64) error {
	var video models.Video
	tx := v.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	//锁住指定id的Video记录
	err := tx.Set("gorm:query_option", "FOR UPDATE").Select("id", "comment_count").Where("id = ?", videoId).Take(&video).Error
	if err != nil {
		return err
	}
	fmt.Println(video)
	// 更新CommentCount
	video.CommentCount += add
	err = tx.Model(&models.Video{}).Where("id = ?", video.ID).UpdateColumn("comment_count", video.CommentCount).Error
	if err != nil {
		return err
	}
	// 提交事务，释放锁
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
