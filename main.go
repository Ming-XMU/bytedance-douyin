package main

import (
	"douyin/config"
	"github.com/gin-gonic/gin"
)

func main() {
	//初始化配置类
	//配置文件在config.ini中
	config.Init()
	r := gin.Default()
	initRouter(r)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
