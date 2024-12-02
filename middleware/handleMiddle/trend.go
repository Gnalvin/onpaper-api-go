package handleMiddle

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	"onpaper-api-go/dao/mongo"
	m "onpaper-api-go/models"
)

// HandleTrendQuery 验证单个trend 查询参数
func HandleTrendQuery(ctx *gin.Context) {
	// 验证 url参数
	var query m.TrendQuery
	err := ctx.ShouldBindQuery(&query)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("query", query)
	ctx.Set("trendId", query.TrendId)
}

// HandleUserTrendQuery 验证查询某个用户动态参数
func HandleUserTrendQuery(ctx *gin.Context) {
	// 验证 url参数
	var query m.TrendUserQuery
	err := ctx.ShouldBindQuery(&query)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	ctx.Set("query", query)
}

// HandleTrendPermission 权限表单验证
func HandleTrendPermission(ctx *gin.Context) {
	// 1.获取用户和密码
	var data m.TrendPermission
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	ctx.Set("permission", data)
	ctx.Set("trendId", data.TrendId)
}

func VerifyTrendOwner(ctx *gin.Context) {
	ctxData, _ := ctx.Get("trendId")
	trendId, _ := ctxData.(int64)

	ctxData, _ = ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	isOwner, err := mongo.VerifyTrendOwner(trendId, userInfo.Id)
	if err != nil {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	if !isOwner {
		err = errors.New(fmt.Sprintf("VerifyTrendOwner no auth user:%s,trend:%d", userInfo.Id, trendId))
		ctl.ResponseErrorAndLog(ctx, ctl.CodeUnPermission, err)
		return
	}
}
