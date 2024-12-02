package cacheMiddle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	c "onpaper-api-go/cache"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"time"
)

// DelCommissionStatus 删除约稿状态相关缓存
func DelCommissionStatus(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	err := c.DelCommissionStatus(loginUser.Id)
	if err != nil {
		logger.ErrZapLog(err, fmt.Sprintf("DelCommissionStatus cache fail userId:%s ", loginUser.Id))
	}
}

// SetAcceptPlan 设置接稿计划缓存
func SetAcceptPlan(ctx *gin.Context) {
	ctxData, _ := ctx.Get("plan")
	plan := ctxData.(m.UserAcceptPlan)

	key := fmt.Sprintf(c.AcceptPlan, plan.UserId)
	err := c.SetOneStringValue(key, plan, 12*time.Hour)
	if err != nil {
		logger.ErrZapLog(err, fmt.Sprintf("SetAcceptPlan cache fail userId:%s ", plan.UserId))
	}
}

// GetAcceptPlan 获取接稿缓存
func GetAcceptPlan(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userId")
	queryId := ctxData.(string)

	key := fmt.Sprintf(c.AcceptPlan, queryId)
	// 反序列化
	var temp m.UserAcceptPlan

	planCtx, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, queryId)
	}

	ctx.Set("cache", planCtx)
}

// SetInvitePlan 设置邀请计划详情
func SetInvitePlan(ctx *gin.Context) {
	ctxData, _ := ctx.Get("plan")
	plan := ctxData.(m.UserInvitePlan)

	key := fmt.Sprintf(c.InvitePlan, plan.InviteId)
	err := c.SetOneStringValue(key, plan, 12*time.Hour)
	if err != nil {
		logger.ErrZapLog(err, fmt.Sprintf("SetAcceptPlan cache fail userId:%s ", plan.UserId))
	}
}

// GetInvitePlan 获取邀请计划详情缓存
func GetInvitePlan(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("inviteId")
	inviteId := ctxData.(int64)

	key := fmt.Sprintf(c.InvitePlan, inviteId)
	// 反序列化
	var temp m.UserInvitePlan

	planCtx, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, inviteId)
	}

	ctx.Set("cache", planCtx)
}

// DelInviteStatus 删除约稿计划相关缓存
func DelInviteStatus(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("inviteId")
	inviteId := ctxData.(int64)

	err := c.DelInvitePlanStatus(inviteId)
	if err != nil {
		logger.ErrZapLog(err, inviteId)
	}
}
