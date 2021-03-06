package main

import (
	"douyin/controller"
	"douyin/interceptor"
	"github.com/gin-gonic/gin"
)

func initRouter(r *gin.Engine) {
	// public directory is used to serve static resources
	r.Static("/static", "./public")
	//后期修改为minio磁盘路径
	//r.Static("/static", "/mnt/data")

	apiRouter := r.Group("/douyin")
	//register interceptor
	apiRouter.Use(interceptor.TokenVerifyVerifyInterceptor())
	apiRouter.Use(interceptor.FavouriteRateLimitInterceptor())
	apiRouter.Use(interceptor.FollowLimitInterceptor())
	// basic apis
	apiRouter.GET("/feed/", controller.Feed)
	apiRouter.GET("/user/", controller.UserInfo)
	apiRouter.POST("/user/register/", controller.Register)
	apiRouter.POST("/user/login/", controller.Login)
	apiRouter.POST("/publish/action/", controller.Publish)
	apiRouter.GET("/publish/list/", controller.PublishList)

	// extra apis - I
	apiRouter.POST("/favorite/action/", controller.FavoriteAction)
	apiRouter.GET("/favorite/list/", controller.FavoriteList)
	apiRouter.POST("/comment/action/", controller.CommentAction)
	apiRouter.GET("/comment/list/", controller.CommentList)

	// extra apis - II
	apiRouter.POST("/relation/action/", controller.RelationAction)
	apiRouter.GET("/relation/follow/list/", controller.FollowList)
	apiRouter.GET("/relation/follower/list/", controller.FollowerList)

	//extra apis - III
	apiRouter.POST("user/update", controller.UpdateUser)
}
