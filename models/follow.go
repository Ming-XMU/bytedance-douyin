package models

//Author: Wechan

const FollowTableName string = "follow"

type Follow struct {
	FollowId   int `gorm:"column:follow_id"`
	FollowerId int `gorm:"column:follower_id"`
}

func (f *Follow) TableName() string {
	return FollowTableName
}
