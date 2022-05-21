package models

const TableNameUser string = "user"

type User struct {
	Id            int64  `gorm:"column:id";primaryKey`
	Name          string `gorm:"column:name"`
	Pwd           string `gorm:"column:pwd"`
	Salt          string `gorm:"column:salt"`
	FollowCount   int64  `gorm:"column:follow_count"`
	FollowerCount int64  `gorm:"column:follower_count"`
}

func (*User) TableName() string {
	return TableNameUser
}
