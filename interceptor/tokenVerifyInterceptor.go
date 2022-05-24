package interceptor

import (
	"douyin/controller"
	"douyin/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

/**
 * @Author: Ember
 * @Date: 2022/5/18 14:12
 * @Description: 拦截器
 **/

var (
	InCludeUrl = []string{
		//user
		"/douyin/user/",
		//feed
		"/douyin/public/action",
		"/douyin/public/list",
		//favorite
		"/douyin/favorite/action",
		"/douyin/favorite/list",
		//comment
		"/douyin/comment/action",
		"/douyin/comment/list",
		//relation
		"/douyin/relation/action",
		"/douyin/relation/follow/list",
		"/douyin/relation/follower/list",
	}

	IncludeUrlMap = map[string]struct{}{
		//user
		"/douyin/user/": {},
		//feed
		"/douyin/publish/action/": {},
		"/douyin/publish/list/":   {},
		//favorite
		"/douyin/favorite/action/": {},
		"/douyin/favorite/list/":   {},
		//comment
		"/douyin/comment/action/": {},
		"/douyin/comment/list/":   {},
		//relation
		//"/douyin/relation/action/":        {},
		"/douyin/relation/follow/list/":   {},
		"/douyin/relation/follower/list/": {},
	}
)

func TokenVerifyVerifyInterceptor() gin.HandlerFunc {
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		//path is in include url
		if _, ok := IncludeUrlMap[path]; ok {
			token := GetToken(context)
			if tools.VerifyParamsEmpty(token) {
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  "token is expire or empty",
				})
				//stop
				context.Abort()
			}
			_, err := tools.VeifyToken(token)
			if err != nil {
				//TODO log format
				fmt.Println("token is expire:", err.Error())
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  err.Error(),
				})
				//stop
				context.Abort()
			}
		}
		context.Next()
	}
}

//different method get token
func GetToken(context *gin.Context) string {
	token := context.Query("token")
	if tools.VerifyParamsEmpty(token) {
		return context.PostForm("token")
	}
	return token
}
