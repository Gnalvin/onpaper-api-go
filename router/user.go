package router

import (
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"

	"github.com/gin-gonic/gin"
)

// userRouter 处理用户相关的路由
func userRouter(router *gin.Engine) {
	// 非验证的路由
	rNoAuth := router.Group("/user", hm.VerifyAuth)
	// 必须验证的路由
	rMustAuth := router.Group("/user", hm.VerifyAuthMust)

	//获取用户主页 profile资料接口
	rNoAuth.GET("/profile", hm.VerifyQueryUserId, cm.GetUserHomeProfile, ctl.GetProfileData, cm.SetUserHomeProfile)
	// 获取用户主页的作品列表
	rNoAuth.GET("/profile/artwork", hm.VerifyUserHomeArtwork, ctl.GetUserHomeArtwork, cm.BatchSetBasicArt, cm.BatchSetArtViews)
	// 获取用户主页收藏作品列表
	rNoAuth.GET("/profile/collect", hm.VerifyUserAndPage, ctl.GetUserHomeCollect, cm.BatchSetBasicArt)

	//推荐关注用户
	rNoAuth.GET("/recommend", ctl.GetRecommendUser)
	//用户排名
	rNoAuth.GET("/rank", hm.VerifyUserRankRequest, cm.GetUserRank, ctl.GetUserRank, cm.SetUserRank)
	//查询用户的粉丝或关注
	rNoAuth.GET("/follow", hm.VerifyUserFollow, ctl.GetUserFollowList, cm.SetUserSmallCarCache)

	//通过用户名搜索用户
	rNoAuth.GET("/search", ctl.SearchUserByName)
	// 用户简单面板资料
	rNoAuth.GET("/panel", hm.VerifyQueryUserId, cm.GetUserPanel, ctl.GetUserPanelInfo, cm.SetUserPanel)

	// 获取导航栏信息
	rMustAuth.GET("/nav", hm.VerifyQueryUserId, ctl.GetNavData)
	//更新用户资料
	rMustAuth.PATCH("/profile", hm.VerifyUpdateProfileData, ctl.UpdateUserProfile, cm.DelUserProfile)
	//添加关注用户
	rMustAuth.POST("/focus", hm.VerifyUserFocus, ctl.SaveUserFocus, cm.SetFocusCount, ctl.SetRecentlyFeed, ctl.SetFocusNotify, cm.SetUserNotifyConfig)
	//精确搜索关注的 用户
	rMustAuth.GET("/focus/search", ctl.SearchOurFocusUser)

	// 获取邀请码
	rMustAuth.GET("/invitation", ctl.GetUserInvitationCode)

	//点赞作品/动态
	rMustAuth.POST("/like", hm.HandlePostInteract, ctl.SaveUserLike, cm.SetLikeCount, ctl.SetLikeOrCollectNotify, cm.SetUserNotifyConfig)
	//收藏
	rMustAuth.POST("/collect", hm.HandlePostInteract, ctl.SaveCollect, cm.SetCollectCount, ctl.SetLikeOrCollectNotify, cm.SetUserNotifyConfig)

	//全站用户展示
	rNoAuth.GET("/show", hm.VerifyAllUserShow, hm.VerifyQuerySign, ctl.GetUserShow, cm.SetUserBigCarCache)
}
