package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	hm "onpaper-api-go/middleware/handleMiddle"
)

// feedbackRouter 反馈接口
func feedbackRouter(router *gin.Engine) {
	//定义路由组
	r := router.Group("/feedback", hm.VerifyAuthMust)
	// 举报
	r.POST("/report", hm.HandlePostReport, ctl.SaveReport)
	//意见反馈
	r.POST("", hm.HandleUserFeedback, ctl.SaveFeedback)
}
