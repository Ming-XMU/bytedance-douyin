package main

import (
	"douyin/config"
	"douyin/controller"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

func main() {
	//初始化配置类
	//配置文件在config.ini中
	config.Init()
	TaskStart()
	r := gin.Default()
	initRouter(r)
	err := r.Run()
	if err != nil {
		logrus.Fatal(err.Error())
	} // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func TaskStart() {
	go func() {
		c := cron.New()
		// 点赞 | 为了方便调试每10秒写回
		err := c.AddFunc("*/10 * * * * ?", controller.FeeSerivce.FlushRedisFavouriteCount)
		if err != nil {
			return
		}
		// 关注 | 为了方便调试每10秒写回
		err = c.AddFunc("*/10 * * * * ?", controller.FollowSerivce.FollowUpdate)
		if err != nil {
			return
		}
		c.Start()
		select {}
	}()
}
