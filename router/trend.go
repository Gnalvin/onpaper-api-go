package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

// trendRouter trend 动态相关的路由
func trendRouter(router *gin.Engine) {
	//定义路由组
	rNoAuth := router.Group("/trend", hm.VerifyAuth)
	rMustAuth := router.Group("/trend", hm.VerifyAuthMust)

	//获取 最新动态
	rNoAuth.GET("/new", hm.HandleNextIdQuery, ctl.GetNewTrend, ctl.GetFeedTrend, cm.BatchSetTrend, cm.BatchSetArtViews)
	//获取单个动态信息
	rNoAuth.GET("/one", hm.HandleTrendQuery, ctl.GetTrendDetail, cm.SetTrendDetail)
	//获取热门动态
	rNoAuth.GET("/hot", ctl.GetHotTrend, ctl.GetFeedTrend, cm.BatchSetTrend, cm.BatchSetArtViews)
	// 获取某个用户的动态
	rNoAuth.GET("/user", hm.HandleUserTrendQuery, ctl.GetUserTrend, ctl.GetFeedTrend, cm.BatchSetTrend)
	// 删除一条动态
	rMustAuth.DELETE("/delete", hm.HandleTrendQuery, hm.VerifyTrendOwner, ctl.DeleteTrend)
	// 权限设置
	rMustAuth.PATCH("/permission", hm.HandleTrendPermission, hm.VerifyTrendOwner, ctl.UpdateTrendPermission)
}
