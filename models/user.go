package models

const TableNameUser string = "user"

type User struct {
	Id   int64
	Name string
	Pwd  string
	Salt string
	//关注和被关注应该是int吧
	FollowCount   int64
	FollowerCount int64
}

func (*User) TableName() string {
	return TableNameUser
}
