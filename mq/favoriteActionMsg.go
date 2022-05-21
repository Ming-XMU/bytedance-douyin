package mq

import "douyin/models"

/**
 * @Author: Ember
 * @Date: 2022/5/21 20:45
 * @Description: TODO
 **/

type FavoriteActionMsg struct {
	//favorite model
	Favorite *models.Favorite
	//action type
	//1:favorite 2:cancel favorite
	Action int
}
