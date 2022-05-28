package main

import (
	"douyin/config"
	"douyin/controller"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func main() {
	//初始化配置类
	//配置文件在config.ini中
	config.Init()
	TaskStart()
	r := gin.Default()
	initRouter(r)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func TaskStart() {
	go func() {
		c := cron.New()
		//每天凌晨1点进行刷新
		err := c.AddFunc("0 1 1 * *  ?", controller.FeeSerivce.FlushRedisFavouriteCount)
		if err != nil {
			return
		}
		//每隔一个小时update关注数
		err = c.AddFunc("0 0 */1 * * ?", controller.FollowSerivce.FollowUpdate)
		if err != nil {
			return
		}
		c.Start()
		select {}
	}()
}
