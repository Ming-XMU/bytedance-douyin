package models

const TableNameFollow string = "follow"

type Follow struct {
	FollowId   int64 `gorm:"primaryKey"`
	FollowerId int64 `gorm:"primaryKey"`
}

func (*Follow) TableName() string {
	return TableNameFollow
}
