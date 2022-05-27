package interceptor

import (
	"douyin/controller"
	"douyin/tools"
	"github.com/gin-gonic/gin"
	"net/http"
)

const FollowActionPath = "/douyin/relation/action/"

// FollowLimitInterceptor
// @author zia
// @Description: 关注限流器
// @return gin.HandlerFunc
func FollowLimitInterceptor() gin.HandlerFunc {
	return func(context *gin.Context) {
		path := context.Request.URL.Path
		if path == FollowActionPath {
			token := context.Query("token")
			loginUser, err := tools.VeifyToken(token)
			if err != nil {
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  "token is expire or empty",
				})
				context.Abort()
			}
			result, err := tools.IsActionLimit(loginUser.UserId)
			if err != nil {
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  err.Error(),
				})
				context.Abort()
			}
			if result {
				context.JSON(http.StatusOK, &controller.Response{
					StatusCode: -1,
					StatusMsg:  "关注太频繁了，休息一下",
				})
				context.Abort()
			}
		}
		context.Next()
	}
}
