package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

// notifyRouter 通知相关的路由
func notifyRouter(router *gin.Engine) {
	//定义路由组
	rMustAuth := router.Group("/notify", hm.VerifyAuthMust)
	// 获取通知未读数
	rMustAuth.GET("/unread", ctl.GetNotifyUnread)
	// 获取点赞收藏通知
	rMustAuth.GET("/like_collect", hm.NotifyQuery, ctl.GetLikeAndCollectNotify, cm.BatchSetBasicArt)
	// 获取关注提醒
	rMustAuth.GET("/focus", hm.NotifyQuery, ctl.GetFocusNotify)
	// 获取评论提醒
	rMustAuth.GET("/comment", hm.NotifyQuery, ctl.GetCommentNotify, cm.BatchSetBasicArt)
	// 获取约稿提醒
	rMustAuth.GET("/commission", hm.NotifyQuery, ctl.GetCommission)
	// 获取消息通知设置
	rMustAuth.GET("/setting", ctl.GetNotifySetting, cm.SetUserNotifyConfig)
	// 更新通知设置
	rMustAuth.PATCH("/setting", hm.UpdateNotifyConfig, ctl.UpdateNotifySetting, cm.SetUserNotifyConfig)
}
