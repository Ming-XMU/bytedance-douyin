package daos

import (
	"douyin/models"
	"gorm.io/gorm"
	"sync"
)

var (
	commentDao     CommentDao
	commentDaoOnce sync.Once
)

type CommentDao interface {
	InsertComment(comment *models.Comment) error
	ListCommentById(videoId int) (comments []models.Comment, err error)
	DeleteComment(commentId int) error
}

type CommentDaoImpl struct {
	db *gorm.DB
}

func (c CommentDaoImpl) InsertComment(comment *models.Comment) error {
	if err := c.db.Debug().Create(comment).Error; err != nil {
		return err
	}
	return nil
}

func (c CommentDaoImpl) ListCommentById(videoId int) ([]models.Comment, error) {
	var comments []models.Comment
	err := c.db.Debug().Select("id", "user_id", "message", "create_at").
		Where("video_id = ?", videoId).Order("create_at desc").Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (c CommentDaoImpl) DeleteComment(commentId int) error {
	var comment models.Comment
	err := c.db.Debug().Where("id = ?", commentId).Delete(&comment).Error
	if err != nil {
		return err
	}
	return nil
}

func GetCommentDao() CommentDao {
	commentDaoOnce.Do(func() {
		commentDao = &CommentDaoImpl{
			db: models.GetDB(),
		}
	})
	return commentDao
}
