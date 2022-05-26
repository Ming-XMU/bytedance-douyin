package mq

import "douyin/models"

//Author:wechan

type CommentActionMsg struct {
	Comment *models.Comment
	//1-发布评论 2-删除评论
	Action int
}
