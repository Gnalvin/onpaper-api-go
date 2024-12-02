package handleMiddle

import (
	"errors"
	"fmt"
	ctl "onpaper-api-go/controller"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/models"
	"onpaper-api-go/utils/verify"
	"strconv"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// HandleOneArtwork 获取请求的作品id
func HandleOneArtwork(ctx *gin.Context) {
	var data models.VerifyOneArtworkId
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	//将上传的 string 格式 转成int64
	num, err := strconv.ParseInt(data.ArtId, 10, 64)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	// 把它传递到上下文
	ctx.Set("artworkId", data.ArtId)
	ctx.Set("intArtId", num)
}

// HandlePostComment 处理上传的作品评论
func HandlePostComment(ctx *gin.Context) {
	var data models.PostCommentData
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	//如果文本长度大于140 则返回错误
	if utf8.RuneCountInString(data.Text) > 140 {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	// 把它传递到上下文
	ctx.Set("comment", data)
}

// HandleArtworkRank 处理作品排名请求
func HandleArtworkRank(ctx *gin.Context) {
	var data models.QueryArtworkRank
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	// 把它传递到上下文
	ctx.Set("rankType", data)
}

// HandleQueryArtworkShow 查询分类作品展示
func HandleQueryArtworkShow(ctx *gin.Context) {
	var data models.QueryChanelType
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	// 把它传递到上下文
	ctx.Set("query", data)
}

// HandleUpdateArtInfo 更新作品信息
func HandleUpdateArtInfo(ctx *gin.Context) {
	var data models.UpdateArtInfo
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(models.UserTokenPayload)

	//标体不超过15个字 描述不超过350个字
	isPass := verify.ArtTextInfo(data.Title, data.Description, data.Tags)
	if !isPass {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	//验证区域是否符合
	zoneVerify := verify.ArtZoneText(data.Zone)
	// 如果区域不是上面几个
	if !zoneVerify {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	//验证whoSee参数
	whoSeeVerify := verify.WhoSee(data.WhoSee)
	// 如果不是上面几个
	if !whoSeeVerify {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	isOwner, err := mysql.VerifyArtOwner(userInfo.Id, data.ArtworkId)
	if err != nil {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}
	if !isOwner {
		err = errors.New(fmt.Sprintf("VerifyArtOwner no auth user:%s,art:%s", userInfo.Id, data.ArtworkId))
		ctl.ResponseErrorAndLog(ctx, ctl.CodeUnPermission, err)
		return
	}

	ctx.Set("artInfo", data)
}

func HandleQueryZone(ctx *gin.Context) {
	var query models.ZoneIndex
	err := ctx.ShouldBindQuery(&query)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("zone", query.Zone)
}
