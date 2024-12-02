package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

// messageRouter 私信相关的路由
func messageRouter(router *gin.Engine) {
	//定义路由组
	r := router.Group("/message", hm.VerifyAuthMust)

	// 发送消息
	r.POST("/send", hm.HandleSendMsg, ctl.SaveMsg)
	// 获取聊天记录
	r.GET("/record", hm.HandleGetChatRecord, ctl.GetChatRecord, cm.SetUserNotifyConfig)
	// 获取会话列表
	r.GET("/chat", hm.HandleGetChatList, ctl.GetChatList)
}
