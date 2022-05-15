package models

const TableNameUser string = "user"

type User struct {
	Id            int64
	Name          string
	Pwd           string
	Salt          string
	FollowCount   string
	FollowerCount string
}

func (*User) TableName() string {
	return TableNameUser
}
