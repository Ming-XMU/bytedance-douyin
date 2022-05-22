package models

const FavoriteTableName string = "favorite"

type Favorite struct {
	UserId  int64 `gorm:"column:user_id"`
	VideoId int64 `gorm:"column:video_id"`
}

func (f *Favorite) TableName() string {
	return FavoriteTableName
}
