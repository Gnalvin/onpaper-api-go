package router

import (
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"

	"github.com/gin-gonic/gin"
)

// 处理上传文件的路由
func fileRouter(router *gin.Engine) {
	saveRouter := router.Group("/save", hm.VerifyAuthMust)
	{
		//查询cos文件信息并保存到数据库 接口
		//保存banner信息
		saveRouter.POST("/banner", hm.HandleBannerInfo, ctl.SaveBannerInfo, cm.DelUserProfile)
		//保存avatar信息
		saveRouter.POST("/avatar", hm.HandleAvatarInfo, ctl.SaveAvatarInfo, cm.DelUserProfile)
		//保存artwork信息
		saveRouter.POST("/artwork", hm.HandleArtworkInfo, ctl.SaveArtworkInfo, cm.SetUserAboutArtCache, ctl.SetFeed)
		//保存trend 信息
		saveRouter.POST("/trend", hm.HandleTrendInfo, ctl.SaveTrendInfo, cm.SetAboutTrendCache, ctl.SetFeed)
	}

	// 删除banner 接口
	router.DELETE("/delete/banner", hm.VerifyAuthMust, ctl.BannerDelete, cm.DelUserProfile)
}
