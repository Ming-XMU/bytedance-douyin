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
	//comment := &models.Comment{ //对应数据库的comment表
	//	ID:       commentId,
	//	VideoId:  videoId,
	//	UserId:   userId,
	//	Message:  commentText,
	//	CreateAt: time.Now(),
	//}
	//commentAction := &mq.CommentActionMsg{
	//	Comment: comment,
	//	Action:  action,
	//}
	//jsonMsg, err := json.Marshal(commentAction)
	//if err != nil {
	//	log.Println("mq pre marshal failed" + err.Error())
	//	return err
	//}
	//rabbitMqSimple := mq.NewRabbitMQSimple("commentActionQueue", "amqp://admin:123456@120.78.238.68:5672/default_host")
	////存入redis
	//// 先加载redis缓存
	//err = f.commentCdRedis(videoId)
	//err = f.commentIdCdRedis()
	//if err != nil {
	//	log.Println("缓存错误")
	//	return err
	//}
	//conn := models.GetRec()
	//defer conn.Close()
	//
	//commentKey := getCommentKey(videoId)
	//
	//if action == 1 { //发布评论
	//
	//	//数据类型使用zset，这里需要考虑到评论时间的问题
	//	commentJSON, _ := json.Marshal(comment)
	//	//TODO 错误处理
	//	//获取commentID并且维护
	//	getcommentId, _ := conn.Do("get", "commentId")
	//	commentId = getcommentId.(int64)
	//	conn.Do("incr", "commentId") //评论后redis中的commentId加1进行维护
	//	// 将发布的评论信息存入redis
	//	conn.Do("ZADD", commentKey, commentId, commentJSON)
	//} else if action == 2 {
	//	//删除redis中的评论
	//	//根据commentid 删除redis评论信息
	//	conn.Do("ZREMBYSCORE", commentKey, commentId, commentId) //根据commentId删除元素
	//}
	//err = rabbitMqSimple.PublishSimple(string(jsonMsg))
	//if err != nil {
	//	//存入mq失败，要回滚
	//	if action == 1 {
	//		conn.Do("ZREMBYSCORE", commentKey, commentId, commentId) //根据commentId删除元素
	//		conn.Do("decr", "commentId")
	//	} else if action == 2 {
	//		// redis中恢复comment信息
	//		comment, _ := f.commentDao.GetCommentByCommentId(int(commentId))
	//		//数据类型使用zset，这里需要考虑到评论时间的问题
	//		commentJSON, _ := json.Marshal(comment)
	//		conn.Do("ZADD", commentKey, commentId, commentJSON)
	//	}
	//}
	//return err
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
	//// 先加载redis缓存
	//err := f.commentCdRedis(videoId)
	//if err != nil {
	//	log.Println("com" + err.Error())
	//	return 0
	//}
	//conn := models.GetRec()
	//defer conn.Close()
	//
	//commentKey := getCommentKey(videoId)
	//commentNum, err := conn.Do("ZCARD", commentKey)
	//if err != nil { //读取缓存数据失败
	//	log.Println("commentCount ZCARD error", err.Error())
	//	//从数据库中获取
	//	var video *models.Video
	//	video, err = daos.GetVideoDao().FindById(videoId)
	//	if err != nil {
	//		return 0
	//	}
	//	return video.CommentCount
	//}
	//return commentNum.(int64)
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
	if err != nil {
		return 0, err
	}
	return commentNum.(int64), nil
}

//////////////////////////////////////////////////////
//第一版方案，评论操作写入redis和消息队列，后mq消费
//实际中redis kv存储，而评论列表和删除评论所依据的key不同，kv要分为videoId-commentIdZSet和commentId-commentMsg分别存入redis
//但又遇到一个问题:所依据的comment id为数据库主键自增id，要这样写的话需要将id改为雪花算法生成
//代码以下实现redis操作，未包含雪花id
//////////////////////////////////////////////////////
//接口直接写入mysql的代码
//author:ch
//if action == 1 {
//		comment := &models.Comment{
//			VideoId:  int64(videoId),
//			UserId:   int64(userId),
//			Message:  commentText,
//			CreateAt: time.Now(),
//		}
//		return f.commentDao.InsertComment(comment)
//	} else if action == 2 {
//		return f.commentDao.DeleteComment(int(commentId))
//	}
//	return errors.New("action is error")
//////////////////////////////////////////////////////
//CommentAction 发布/删除评论
// action 1-发布评论，2-删除评论
// author wechan
//const commentCacheName = "comment_cache_set"
//func (f *CommentServiceImpl) CommentAction(userId, videoId, commentId int64, action int, commentText string) error {
//	comment := &models.Comment{ //对应数据库的comment表
//		ID:       commentId,
//		VideoId:  videoId,
//		UserId:   userId,
//		Message:  commentText,
//		CreateAt: time.Now(),
//	}
//	commentAction := &mq.CommentActionMsg{
//		Comment: comment,
//		Action:  action,
//	}
//	jsonMsg, err := json.Marshal(commentAction)
//	if err != nil {
//		log.Println("mq pre marshal failed" + err.Error())
//		return err
//	}
//	rabbitMqSimple := mq.NewRabbitMQSimple("commentActionQueue", "amqp://admin:123456@120.78.238.68:5672/default_host")
//	//存入redis
//	// 加载redis缓存
//	if action == 1 { //发布评论
//		conn := models.GetRec()
//		defer conn.Close()
//		//数据类型使用zset，这里需要考虑到评论时间的问题
//		commentJSON, _ := json.Marshal(comment)
//		//错误处理
//		// 将发布的评论信息存入redis
//		//两种redis存储
//		//videoid-commentid
//		//commentid-commentmsg
//		commentcacheVideoidcommentidkey := getCommentVideoIdKey(videoId)
//		_, err = conn.Do("ZADD", commentcacheVideoidcommentidkey, comment.CreateAt, commentId)
//		commentcacheCommentidmsgkey := getCommentIDKey(commentId)
//		_, err = conn.Do("ZADD", commentcacheCommentidmsgkey, comment.CreateAt, commentJSON)
//
//	} else if action == 2 {
//		//删除redis中的评论
//		conn := models.GetRec()
//		defer conn.Close()
//		//根据commentid 删除redis评论信息
//		commentcacheVideoidcommentidkey := getCommentVideoIdKey(videoId)
//		_, err = conn.Do("ZREM", commentcacheVideoidcommentidkey, string(commentId))
//		commentcacheCommentidmsgkey := getCommentIDKey(commentId)
//		_, err = conn.Do("del", commentcacheCommentidmsgkey)
//	}
//	err = rabbitMqSimple.PublishSimple(string(jsonMsg))
//	if err != nil {
//		//存入mq失败，要回滚
//	}
//	return err
//}
//
//func getCommentVideoIdKey(videoId int64) string {
//	return strings.Join([]string{commentCacheName, "videoId", string(videoId)}, "_")
//}
//func getCommentIDKey(commentId int64) string {
//	return strings.Join([]string{commentCacheName, "commentId", string(commentId)}, "_")
//}
