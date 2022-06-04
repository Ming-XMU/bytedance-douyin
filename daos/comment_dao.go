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
	GetcCommentIdNext() (num int64, err error)
	GetCommentByCommentId(commentId int) (*models.Comment, error)
	GetCommentCountByVideoId(videoId int) (int64, error)
}

type CommentDaoImpl struct {
	db *gorm.DB
}

func (c CommentDaoImpl) GetCommentCountByVideoId(videoId int) (int64, error) {
	var count int64
	if err := c.db.Debug().Where("video_id = ?", videoId).Count(&count).Error; err != nil {
		return -1, err
	}
	return count, nil
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

// GetcCommentIdNext 获取下一条自增主键Id
// author:wechan
func (c CommentDaoImpl) GetcCommentIdNext() (num int64, err error) {
	comment := &models.Comment{
		VideoId: 0,
		UserId:  0,
		Message: "",
	}
	c.db.Debug().Create(comment)
	num = comment.ID
	c.DeleteComment(int(num))
	//select table_name, AUTO_INCREMENT from information_schema.tables where table_name="get_max_id";
	//num = c.db.Debug().Select("auto_increment", "information_schema.'TABLES'").Where("table_name=?", "Comment")
	return num + 1, nil
}

func (c *CommentDaoImpl) GetCommentByCommentId(commentId int) (*models.Comment, error) {
	var comment models.Comment
	err := c.db.Where("id=?", commentId).Find(&comment)
	if err != nil {
		//错误处理 TODO
	}
	return &comment, nil
}
