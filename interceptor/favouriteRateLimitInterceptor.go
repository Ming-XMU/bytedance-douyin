package interceptor

import (
	"douyin/controller"
	"douyin/tools"
	"github.com/anqiansong/ketty/console"
	"github.com/gin-gonic/gin"
	"net/http"
)

/**
 * @Author: Ember
 * @Date: 2022/5/25 19:53
 * @Description: TODO
 **/

//favorite action rate limit
var (
	favouriteActionPath = "/douyin/favorite/action/"
)

func FavouriteRateLimitInterceptor() gin.HandlerFunc {

	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == favouriteActionPath {
			token := context.Query("token")
			loginInfo, err := tools.VeifyToken(token)
			if err != nil {
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  "token is expire or empty",
				})
				//stop
				context.Abort()
				return
			}
			result, err := tools.FavouriteRateLimit(loginInfo.UserId)
			if err != nil {
				console.Error(err)
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  "occured unknown error",
				})
				//stop
				context.Abort()
				return
			}
			resultToInt := result.(int64)
			if resultToInt == 0 {
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  "点赞太频繁了，休息一下",
				})
				//stop
				context.Abort()
				return
			}
		}
		context.Next()
	}
}
