package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

func commentRouter(router *gin.Engine) {
	rMustAuth := router.Group("/comment", hm.VerifyAuthMust)
	rNoAuth := router.Group("/comment", hm.VerifyAuth)
	//获取单个作品/动态的 评论
	rNoAuth.GET("/root", hm.HandleCommentQuery, cm.GetRootComment, ctl.GetRootComment, cm.SetRootComment)
	//获取作品/动态里的单条根评论的子评论
	rNoAuth.GET("/reply", hm.HandleRootCommentReply, cm.GetChildrenComment, ctl.GetCommentReply, cm.SetChildrenComment)
	// 获取单条根评论详情
	rNoAuth.GET("/root/one", hm.HandleOneRootComment, ctl.GetOneRootComment)

	//评论点赞接口
	rMustAuth.POST("/like", hm.HandleCommentLike, ctl.SaveCommentLike, ctl.SetLikeCommentNotify, cm.SetUserNotifyConfig)
	//发布作品/动态的评论
	rMustAuth.POST("", hm.HandlePostComment, ctl.SaveComment, cm.AddComment, ctl.SetCommentNotify, cm.SetUserNotifyConfig)
	//删除评论接口
	rMustAuth.DELETE("", hm.HandleCommentDelete, ctl.DelComment)
}
