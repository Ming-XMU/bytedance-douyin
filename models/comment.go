package models

import "time"

const CommentTableName string = "comment"

type Comment struct {
	ID       int64     `gorm:"column:id"`
	VideoId  int64     `gorm:"column:video_id"`
	UserId   int64     `gorm:"column:user_id"`
	Message  string    `gorm:"column:message"`
	CreateAt time.Time `gorm:"autoCreateTime;column:create_at"`
}

func (v *Comment) TableName() string {
	return CommentTableName
}
