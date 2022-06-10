package services

import (
	"douyin/daos"
	"douyin/models"
	"douyin/tools"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	commentService     CommentService
	commentServiceOnce sync.Once
)

const commentCacheName = "comment_cache"

func GetCommentService() CommentService {
	commentServiceOnce.Do(func() {
		commentService = &CommentServiceImpl{
			commentDao: daos.GetCommentDao(),
		}
	})
	return commentService
}

type CommentService interface {
	CommentAction(userId, videoId, commentId int64, action int, commentText string) error
	CommentList(userId, videoId int) ([]models.Comment, error)
	GetCommentCount(videoId int64) int64
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

func (f *CommentServiceImpl) CommentAction(userId, videoId, commentId int64, action int, commentText string) error {
	if action == 1 {
		comment := &models.Comment{
			VideoId:  videoId,
			UserId:   userId,
			Message:  commentText,
			CreateAt: time.Now(),
		}
		err := removeRedisCommentCountKey(videoId)
		if err != nil {
			logrus.Errorln("removeRedisCommentCountKey is false by ", videoId)
		}
		return f.commentDao.InsertComment(comment)
	} else if action == 2 {
		err := removeRedisCommentCountKey(videoId)
		if err != nil {
			logrus.Errorln("removeRedisCommentCountKry is false by ", videoId)
		}
		return f.commentDao.DeleteComment(int(commentId))
	}
	return errors.New("action is error")
}

func getCommentKey(videoId int64) string {
	return strings.Join([]string{commentCacheName, strconv.FormatInt(videoId, 10)}, "_")
}

//commentCdRedis 加载redis中comment缓存
func (f *CommentServiceImpl) commentCdRedis(videoId int64) error {
	commentkey := getCommentKey(videoId)
	if tools.RedisKeyExists(commentkey) {
		return tools.RedisKeyFlush(commentkey)
	}
	comments, err := f.commentDao.ListCommentById(int(videoId))
	if err != nil {
		return err
	}
	for _, comment := range comments {
		commentJson, _ := json.Marshal(comment)
		_, _ = tools.RedisDo("ZADD", commentkey, comment.ID, commentJson)
	}
	_, _ = tools.RedisDo("EXPIRE", commentkey, 1800)
	return nil
}

//commentIdCdRedis 获取缓存的下一个自增主键的commentId
//注意：这里redis的key为commentId val为mysql下一个自增主键的值
func (f *CommentServiceImpl) commentIdCdRedis() error {
	if tools.RedisKeyExists("commentId") {
		return tools.RedisKeyFlush("commentId")
	}
	num, err := daos.GetCommentDao().GetCommentIdNext()
	if err != nil {
		logrus.Errorln("getcommentnextid err :", err)
	}
	_, _ = tools.RedisDo("SET", "commentId", num)
	_, _ = tools.RedisDo("EXPIRE", "commentId", 1800)
	return nil
}

func (f *CommentServiceImpl) GetCommentCount(videoId int64) int64 {
	count, err := getRedisCommentCountKey(videoId)
	if err != nil {
		c, e := addRedisCommentCountKey(videoId)
		if e != nil {
			return 0
		}
		return c
	}
	return count
}

func addRedisCommentCountKey(videoId int64) (int64, error) {
	commentKey := getCommentKey(videoId)
	count, err := daos.GetCommentDao().GetCommentCountByVideoId(int(videoId))
	if err != nil {
		return -1, err
	}
	_, _ = tools.RedisDo("SET", commentKey, count)
	return count, nil
}

func removeRedisCommentCountKey(videoId int64) error {
	commentKey := getCommentKey(videoId)
	_, _ = tools.RedisDo("DEL", commentKey)
	return nil
}

func getRedisCommentCountKey(videoId int64) (int64, error) {
	commentKey := getCommentKey(videoId)
	commentNum, err := tools.RedisDo("GET", commentKey)
	if commentNum == nil || err != nil {
		return 0, err
	}
	return commentNum.(int64), nil
}
