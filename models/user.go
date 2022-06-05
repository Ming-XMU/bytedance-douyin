package models

const TableNameUser string = "user"

type User struct {
	Id            int64  `gorm:"column:id";primaryKey`
	Name          string `gorm:"column:name"`
	Pwd           string `gorm:"column:pwd"`
	Salt          string `gorm:"column:salt"`
	FollowCount   int64  `gorm:"column:follow_count"`
	FollowerCount int64  `gorm:"column:follower_count"`
	Avatar        string `gorm:"column:avatar"`
	Signature     string `gorm:"column:signature "`
	BackGround    string `gorm:"column:background_image"`
}

func (*User) TableName() string {
	return TableNameUser
}

type UserMessage struct {
	Id            int64  `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	FollowCount   int64  `json:"follow_count"`
	FollowerCount int64  `json:"follower_count"`
	IsFollow      bool   `json:"is_follow"`
	Avatar        string `json:"column:avatar"`
	Signature     string `json:"column:signature "`
	BackGround    string `json:"column:background_image"`
}
