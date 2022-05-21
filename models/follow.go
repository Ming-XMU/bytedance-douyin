package models

//Author: Wechan

const FollowTableName string = "follow"

type Follow struct {
	FollowId   int `gorm:"column:follow_id";primaryKey`
	FollowerId int `gorm:"column:follower_id";primaryKey`
}

func (f *Follow) TableName() string {
	return FollowTableName
}
