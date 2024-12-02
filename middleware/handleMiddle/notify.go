package handleMiddle

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	"onpaper-api-go/models"
)

// NotifyQuery 验证提醒参数
func NotifyQuery(ctx *gin.Context) {
	var data models.NotifyQuery
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeParamsError, err)
		return
	}
	ctx.Set("query", data)
}

// UpdateNotifyConfig 修改通知设置
func UpdateNotifyConfig(ctx *gin.Context) {
	var data models.NotifyConfig
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeParamsError, err)
		return
	}
	ctx.Set("config", data)
}
