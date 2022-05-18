package models

const FavoriteTableName string = "favorite"

type Favorite struct {
	UserId  int `gorm:"column:user_id"`
	VideoId int `gorm:"column:video_id"`
}

func (f *Favorite) TableName() string {
	return FavoriteTableName
}
