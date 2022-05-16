package models

import "time"

/**
 * @Author: Ember
 * @Date: 2022/5/16 9:54
 * @Description: TODO
 **/
var(
	VideoTableName = "video"
)
type Video struct{
	//primary key
	ID int64 `gorm:"column:id";primaryKey`
	//outer user table
	UserId int64 `gorm:"column:user_id"`
	//play url
	PlayUrl string `gorm:"column:play_url"`
	//cover url
	CoverUrl string `gorm:"column:cover_url"`
	//comment count
	CommentCount int32 `gorm:"column:comment_count"`
	//favorite count
	FavoriteCount int32 `gorm:"column:favorite_count"`
	//create time
	CreateAt time.Time `gorm:"column:create_at;autoCreateTime"`
}
func(v *Video) TableName()string{
	return VideoTableName
}