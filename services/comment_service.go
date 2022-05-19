package services

import (
	"douyin/daos"
	"douyin/models"
	"errors"
	"sync"
	"time"
)

var (
	commentService     CommentService
	commentServiceOnce sync.Once
)

func GetCommentService() CommentService {
	commentServiceOnce.Do(func() {
		commentService = &CommentServiceImpl{
			commentDao: daos.GetCommentDao(),
		}
	})
	return commentService
}

type CommentService interface {
	CommentAction(userId, videoId, commentId, action int, commentText string) error
	CommentList(userId, videoId int) ([]models.Comment, error)
}

type CommentServiceImpl struct {
	commentDao daos.CommentDao
}

// CommentList 评论列表
func (f *CommentServiceImpl) CommentList(userId, videoId int) ([]models.Comment, error) {
	comments, err := f.commentDao.ListCommentById(videoId)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// CommentAction 发布/删除评论
// action 1-发布评论，2-删除评论
func (f *CommentServiceImpl) CommentAction(userId, videoId, commentId, action int, commentText string) error {
	if action == 1 {
		comment := &models.Comment{
			VideoId:  int64(videoId),
			UserId:   int64(userId),
			Message:  commentText,
			CreateAt: time.Now(),
		}
		return f.commentDao.InsertComment(comment)
	} else if action == 2 {
		return f.commentDao.DeleteComment(commentId)
	}
	return errors.New("action is error")
}
