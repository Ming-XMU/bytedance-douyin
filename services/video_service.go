package services

import (
	"douyin/daos"
	"sync"
)

var (
	videoService VideoService
	once         sync.Once
)

type VideoService interface {
	VideoExist(id int64) bool
}

type VideoServiceImpl struct {
	videoDao daos.VideoDao
}

func GetVideoService() VideoService {
	once.Do(func() {
		videoService = &VideoServiceImpl{
			videoDao: daos.GetVideoDao(),
		}
	})
	return videoService
}

func (v *VideoServiceImpl) VideoExist(id int64) bool {
	_, err := v.videoDao.FindById(id)
	if err != nil {
		return false
	}
	return true
}
