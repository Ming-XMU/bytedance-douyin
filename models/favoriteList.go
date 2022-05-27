package models

/**
 * @Author: Ember
 * @Date: 2022/5/27 20:38
 * @Description: TODO
 **/

type FavoriteList struct{
	UserId  int64 `gorm:"column:user_id"`
	VideoId int64 `gorm:"column:video_id"`
	Author User `gorm:"foreignKey:UserId"`
	Video Video `gorm:"foreignKey:VideoId"`
}
func (f *FavoriteList) TableName() string {
	return FavoriteTableName
}