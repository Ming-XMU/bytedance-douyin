package controller

import (
	"douyin/models"
	"douyin/services"
	"douyin/tools"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type CommentListResponse struct {
	Response
	CommentList []Comment `json:"comment_list,omitempty"`
}

// CommentAction no practical effect, just check if token is valid
func CommentAction(c *gin.Context) {
	token := c.Query("token")

	// 身份认证
	user, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "请先登录！"},
		})
		return
	}
	// 获取请求参数
	videoId, e1 := strconv.Atoi(c.Query("video_id"))
	actionType, e2 := strconv.Atoi(c.Query("action_type"))

	exist := services.GetVideoService().VideoExist(int64(videoId))

	// 异常处理
	if e1 != nil || e2 != nil || (actionType != 1 && actionType != 2) || !exist {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "参数错误！"})
		return
	}

	var commentText string
	var commentId int
	if actionType == 1 {
		commentText = c.Query("comment_text")
	} else {
		cId, e4 := strconv.Atoi(c.Query("comment_id"))
		if e4 != nil {
			c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "参数错误！"})
			return
		}
		commentId = cId
	}

	// Service调用
	err = services.GetCommentService().CommentAction(user.UserId, int64(videoId), int64(commentId), actionType, commentText)

	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, Response{
			StatusCode: 0,
			StatusMsg:  "操作成功",
		})
	}
}

// CommentList all videos have same demo comment list
func CommentList(c *gin.Context) {
	// 这里请求也携带Token，暂时按照需要认证做
	token := c.Query("token")
	// 身份认证
	loginUser, err := tools.VeifyToken(token)
	if err != nil {
		c.JSON(http.StatusOK, UserResponse{
			Response: Response{StatusCode: 0, StatusMsg: "请先登录！"},
		})
		return
	}
	// 获取请求参数
	videoId, e1 := strconv.Atoi(c.Query("video_id"))

	exist := services.GetVideoService().VideoExist(int64(videoId))
	// 异常处理
	if e1 != nil || !exist {
		c.JSON(http.StatusOK, Response{StatusCode: 1, StatusMsg: "参数错误！"})
		return
	}
	comments, err := services.GetCommentService().CommentList(int(loginUser.UserId), videoId)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
	} else {
		var commentList []Comment
		var userMap map[int64]models.UserMessage // 缓存
		userMap = make(map[int64]models.UserMessage)
		// 这部分比较冗余，后续优化
		// 主要是根据评论的userid查询发布该评论的用户信息
		for _, comment := range comments {
			var tmp Comment
			tmp.Id = comment.ID
			tmp.Content = comment.Message
			tmp.CreateDate = comment.CreateAt.Format("2006-01-02")
			uId := comment.UserId
			user, ok := userMap[uId]
			if !ok {
				usr, e := services.GetUserService().UserInfo(strconv.FormatInt(uId, 10))
				if e != nil {
					continue
				} else {
					user.Id = usr.Id
					user.Name = usr.Name
					user.FollowCount = usr.FollowCount
					user.FollowerCount = usr.FollowerCount
					user.Avatar = usr.Avatar
					user.BackGround = usr.BackGround
					user.Signature = usr.Signature
					info := services.GetFollowService().UserFollowInfo(&models.User{Id: loginUser.UserId}, strconv.FormatInt(usr.Id, 10))
					user.IsFollow = info.IsFollow
					userMap[uId] = user
				}
			}
			fmt.Println(userMap)
			tmp.User = user
			commentList = append(commentList, tmp)
		}
		c.JSON(http.StatusOK, CommentListResponse{
			Response:    Response{StatusCode: 0},
			CommentList: commentList,
		})
	}
}
