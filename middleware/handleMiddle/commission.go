package handleMiddle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	ctl "onpaper-api-go/controller"
	"onpaper-api-go/dao/mongo"
	m "onpaper-api-go/models"
	"onpaper-api-go/settings"
	"onpaper-api-go/utils/oss"
	"onpaper-api-go/utils/verify"
	"strconv"
)

// VerifyPostContract 验证上传的 接稿计划
func VerifyPostContract(ctx *gin.Context) {
	var data m.AcceptPlan
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		fmt.Println(err.Error())
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ok := verify.FileTypeList(data.FileType)
	if !ok {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	// 把它传递到上下文
	ctx.Set("contractPlan", data)
}

// VerifyInvitePlan 验证上传的约稿邀请
func VerifyInvitePlan(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	var data m.InvitePlan
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	// 自己不能给自己约稿
	if userInfo.Id == data.UserId {
		ctl.ResponseError(ctx, ctl.CodeUnPermission)
		return
	}

	ok := verify.FileTypeList(data.FileType)
	if !ok {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ok, _ = verify.DateRule(data.Date)
	if !ok {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	// cos验证文件是否存在
	for i, file := range data.FileList {
		key := "commission/" + userInfo.Id + "/" + file.FileName
		mErr := oss.MoveTempToOriginal(key)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "MoveTempToOriginal fail")
			ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, mErr)
			return
		}
		//到cos中查询文件信息
		fInfo, fErr := oss.SelectOssFileInfo(settings.Conf.OriginalBucket, key)
		if fErr != nil {
			fErr = errors.Wrap(fErr, "cos 查询错误")
			ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, fErr)
			return
		}
		// 文件类型
		contentType := fInfo.Get("Content-Type")
		// 文件大小
		contentLength := fInfo.Get("Content-Length")
		//字符串 -> int64
		size, _ := strconv.ParseInt(contentLength, 10, 64)

		data.FileList[i] = m.PicsType{
			FileName: file.FileName,
			Mimetype: contentType,
			Size:     size,
			Sort:     file.Sort,
			Width:    file.Width,
			Height:   file.Height,
		}

	}

	// 把它传递到上下文
	ctx.Set("invitePlan", data)
}

// VerifyQueryPlan 查询用户计划
func VerifyQueryPlan(ctx *gin.Context) {
	var data m.PlanQuery
	// 验证 url参数
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("query", data)

}

func VerifyPlanQueryId(ctx *gin.Context) {
	var data m.PlanIdQuery
	// 验证 url参数
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("inviteId", data.InviteId)
}

// VerifyEvaluate 验证约稿评价
func VerifyEvaluate(ctx *gin.Context) {
	var data m.Evaluate
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	// 查找约稿方案的两个用户
	userInfo, err := mongo.GetPlanUserInfo(data.InviteId)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			ctl.ResponseError(ctx, ctl.CodeParamsError)
			return
		}
		err = errors.Wrap(err, "GetPlanUserInfo mongodb fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	// 如果 不是计划中的两个用户不能评价
	if userInfo.Sender != loginUser.Id && userInfo.ArtistId != loginUser.Id {
		ctl.ResponseError(ctx, ctl.CodeUnPermission)
		return
	}

	data.InviteOwn = userInfo.Sender
	data.Score = float64(data.Rate1+data.Rate2+data.Rate3) / 3
	ctx.Set("evaluate", data)
	ctx.Set("planUser", userInfo)
}

// VerifyPlanNext 处理计划下一步
func VerifyPlanNext(ctx *gin.Context) {
	var planNext m.PlanNext
	err := ctx.ShouldBindJSON(&planNext)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	// 查找约稿方案的两个用户
	userInfo, err := mongo.GetPlanUserInfo(planNext.InviteId)
	if err != nil {
		if err == mongodb.ErrNoDocuments {
			ctl.ResponseError(ctx, ctl.CodeParamsError)
			return
		}
		err = errors.Wrap(err, "GetPlanUserInfo mongodb fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	// 如果 不是计划中的两个用户不能修改方案
	if userInfo.Sender != loginUser.Id && userInfo.ArtistId != loginUser.Id {
		ctl.ResponseError(ctx, ctl.CodeUnPermission)
		return
	}

	// 0 未接受 1 沟通中  2 创作中 3 已完成  -1 画师/约稿人关闭(待接稿阶段和沟通阶段关闭) -2 退出（创作中散伙）
	// 修改状态大于0 只能画师操作
	if planNext.Status > 0 && loginUser.Id != userInfo.ArtistId {
		ctl.ResponseError(ctx, ctl.CodeUnPermission)
		return
	}

	ctx.Set("planNext", planNext)
	ctx.Set("planUser", userInfo)
}

// VerifyEvaluateQuery 验证查询用户评论
func VerifyEvaluateQuery(ctx *gin.Context) {
	var data m.EvaluateQuery
	// 验证 url参数
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("query", data)
}

// VerifyCommissionStatus 验证上传的约稿状态
func VerifyCommissionStatus(ctx *gin.Context) {
	var status m.CommissionStatus
	err := ctx.ShouldBindJSON(&status)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctx.Set("status", status.Status)
}
