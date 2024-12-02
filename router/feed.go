package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

// feedRouter feed流相关的路由
func feedRouter(router *gin.Engine) {
	//定义路由组
	r := router.Group("/feed", hm.VerifyAuthMust)

	//获取 aw 类型的 feed
	r.GET("/art", hm.HandleNextIdQuery, ctl.GetArtFeed, cm.BatchSetBasicArt)
	// 获取所有类型的 feed
	r.GET("/all", hm.HandleNextIdQuery, ctl.GetAllFeed, ctl.GetFeedTrend, cm.BatchSetTrend, cm.BatchSetArtViews)
}
