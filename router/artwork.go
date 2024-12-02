package router

import (
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"

	"github.com/gin-gonic/gin"
)

// 作品信息相关路由
func artworkRouter(router *gin.Engine) {
	rMustAuth := router.Group("/artwork", hm.VerifyAuthMust)
	rNoAuth := router.Group("/artwork", hm.VerifyAuth)
	//获取单个作品信息
	rNoAuth.GET("/info", hm.HandleOneArtwork, cm.GetArtworkProfile, ctl.GetOneArtworkInfo, cm.SetArtworkProfile)
	//获取作品排行数据
	rNoAuth.GET("/rank", hm.HandleArtworkRank, cm.GetArtworkRank, ctl.GetArtworkRank, cm.SetArtworkRank)
	//获取首页热门作品数据
	rNoAuth.GET("/hot", ctl.GetHotArtwork, cm.BatchSetArtViews)
	//查询分类作品展示
	rNoAuth.GET("/show", hm.HandleQueryArtworkShow, hm.VerifyQuerySign, ctl.GetChannelArtwork, cm.BatchSetBasicArt)
	//首页下拉加载分区作品
	rNoAuth.GET("/hot/zone", hm.HandleQueryZone, ctl.GetHomePageZone, cm.BatchSetArtViews)

	// 更新作品资料
	rMustAuth.PATCH("/info", hm.HandleUpdateArtInfo, ctl.UpdateArtInfo)
	//删除作品
	rMustAuth.DELETE("/delete", hm.HandleOneArtwork, ctl.DeleteArtwork)
}
