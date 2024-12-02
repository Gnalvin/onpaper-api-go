package controller

import (
	"github.com/gin-gonic/gin"
	"onpaper-api-go/dao/mongo"
	m "onpaper-api-go/models"
	"time"
)

// SaveReport 保存举报
func SaveReport(ctx *gin.Context) {
	ctxData, _ := ctx.Get("report")
	report := ctxData.(m.PostReport)

	ctxData, _ = ctx.Get("userInfo")
	loginInfo := ctxData.(m.UserTokenPayload)

	report.PostUser = loginInfo.Id
	err := mongo.SaveReport(report)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"report": "ok",
	})
}

// SaveFeedback 保存意见反馈
func SaveFeedback(ctx *gin.Context) {
	dataCtx, _ := ctx.Get("userInfo")
	loginUserInfo := dataCtx.(m.UserTokenPayload)

	dataCtx, _ = ctx.Get("feedback")
	feedback := dataCtx.(m.PostFeedback)

	feedback.UserId = loginUserInfo.Id
	feedback.CreateAt = time.Now()

	err := mongo.SaveFeedBack(feedback)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"feedback": "ok",
	})
}
