package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

// 作品信息相关路由
func commissionRouter(router *gin.Engine) {
	rMustAuth := router.Group("/commission", hm.VerifyAuthMust)
	rNoAuth := router.Group("/commission", hm.VerifyAuth)

	// 验证创建接稿方案的权限
	rMustAuth.GET("/check", ctl.CheckAcceptPermission)
	// 创建/编辑接稿方案
	rMustAuth.POST("/accept", hm.VerifyPostContract, ctl.SaveContractPlan, cm.DelCommissionStatus)
	// 发送约稿邀请
	rMustAuth.POST("/invite", hm.VerifyInvitePlan, ctl.SaveInvitePlan, ctl.SetCommissionNotify)
	// 查看用户发出的邀请
	rMustAuth.GET("/send", hm.VerifyQueryPlan, ctl.GetSendPlan)
	// 计划下一步
	rMustAuth.PATCH("/next", hm.VerifyPlanNext, ctl.HandlePlanNext, ctl.SetCommissionNotify, cm.DelInviteStatus)
	// 查看双方联系方式
	rMustAuth.GET("/contact", hm.VerifyPlanQueryId, ctl.GetUserContact)
	// 发布约稿评价
	rMustAuth.POST("/evaluate", hm.VerifyEvaluate, ctl.SaveEvaluate, ctl.SetCommissionNotify, cm.DelInviteStatus)
	// 更新约稿状态
	rMustAuth.PATCH("/status", hm.VerifyCommissionStatus, ctl.UpdateCommissionStatus, cm.DelCommissionStatus)

	// 查看用户接稿方案
	rNoAuth.GET("/plan", hm.VerifyQueryUserId, cm.GetAcceptPlan, ctl.GetAcceptPlan, cm.SetAcceptPlan)
	// 查看用户收到的邀请
	rNoAuth.GET("/receive", hm.VerifyQueryPlan, ctl.GetInvitePlan)
	// 查看约稿详情
	rNoAuth.GET("/detail", hm.VerifyPlanQueryId, cm.GetInvitePlan, ctl.GetPlanDetail, cm.SetInvitePlan)
	// 查看完稿评价
	rNoAuth.GET("/evaluate", hm.VerifyEvaluateQuery, ctl.GetUserReceiveEvaluate)
}
