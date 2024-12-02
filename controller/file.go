package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	"onpaper-api-go/models"
	"onpaper-api-go/settings"
	"onpaper-api-go/utils/oss"
	"strconv"
	"strings"
)

// SaveBannerInfo 保存上传的背景图片信息
func SaveBannerInfo(ctx *gin.Context) {
	//1.获取所有的图像信息
	ctxData, _ := ctx.Get("fileInfo")
	info := ctxData.(models.CallBackFileInfo)
	ctxData, _ = ctx.Get("userInfo")
	userInfo := ctxData.(models.UserTokenPayload)

	//替换之前的背景图片数据
	_, err := mysql.UpdateBannerInfo(info, userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//发送 压缩消息到服务器
	intMid, _ := strconv.ParseInt(userInfo.Id, 10, 64)
	err = cache.SendCompressQueue(userInfo.Id, intMid, "br")
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{"fileName": info.FileName})
}

// BannerDelete 删除banner
func BannerDelete(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, exi := ctx.Get("userInfo")
	// 如果取不到数据
	if !exi {
		err := errors.New("BannerDelete get ctxData fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 类型断言
	userInfo := ctxData.(models.UserTokenPayload)

	var bannerInfo models.DeleteBannerInfo
	// 1.绑定body json 信息到 bannerInfo
	err := ctx.ShouldBindJSON(&bannerInfo)
	if err != nil {
		//如果获取参数错误
		ResponseError(ctx, CodeParamsError)
		return
	}

	// 查询登陆用户对应的 banner 信息
	fileInfo, err := mysql.GetBannerInfo(userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 如果上传的数据 和数据库的 数据对不上
	if fileInfo.FileName.String != bannerInfo.FileName {
		ResponseError(ctx, CodeParamsError)
		return
	}

	// 到数据库中删除
	err = mysql.DeleteBanner(userInfo.Id)
	if err != nil {
		//如果删除出错错误
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, "ok")

	//到云服务器删除
	err = oss.BatchDeleteOssObject(settings.Conf.PreviewBucket, "banners/"+userInfo.Id+"/", "")
	if err != nil {
		logger.ErrZapLog(err, "BannerDelete BatchDeleteCosObject fail")
	}
}

// SaveAvatarInfo 保存上传的头像信息
func SaveAvatarInfo(ctx *gin.Context) {
	//1.获取所有的图像信息
	ctxData, _ := ctx.Get("fileInfo")
	info := ctxData.(models.CallBackFileInfo)

	ctxData, _ = ctx.Get("userInfo")
	userInfo := ctxData.(models.UserTokenPayload)

	err := oss.MoveTempToPreView("avatars/" + userInfo.Id + "/" + info.FileName)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	sAvatar := strings.Join(strings.Split(info.FileName, "."), "_s.")
	err = oss.MoveTempToPreView("avatars/" + userInfo.Id + "/" + sAvatar)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//替换之前的头像图片数据
	err = mysql.UpdateAvatarInfo(info, userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{"fileName": info.FileName})

	err = oss.BatchDeleteOssObject(settings.Conf.TempBucket, "avatars/"+userInfo.Id+"/", "")
	if err != nil {
		logger.ErrZapLog(err, "SaveAvatarInfo BatchDeleteCosObject fail")
	}
}

// SaveArtworkInfo 保存上传的作品信息
func SaveArtworkInfo(ctx *gin.Context) {
	//1.获取传递的 作品信息
	ctxData, _ := ctx.Get("artworkInfo")
	artworkInfo := ctxData.(*models.SaveArtworkInfo)

	//到数据库中保存
	err := mysql.CreateArtworkInfo(artworkInfo)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 给自己的feed 流添加一条
	err = mongo.SetTheUserFeed([]int64{artworkInfo.ArtworkId}, "aw", artworkInfo.UserId, artworkInfo.UserId)
	if err != nil {
		err = errors.Wrap(err, "SetTheUserFeed fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	//发送 压缩消息到服务器
	err = cache.SendCompressQueue(artworkInfo.UserId, artworkInfo.ArtworkId, "aw")
	if err != nil {
		err = errors.Wrap(err, "SendCompressQueue fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"artworkId": strconv.FormatInt(artworkInfo.ArtworkId, 10),
	})

	ctx.Set("feed", models.UploadArtOrTrend{
		MsgID:  artworkInfo.ArtworkId,
		SendId: artworkInfo.UserId,
		Type:   "aw",
	})
}

// SaveTrendInfo 保存trend 信息
func SaveTrendInfo(ctx *gin.Context) {
	//1.获取传递的 作品信息
	ctxData, _ := ctx.Get("trendInfo")
	trendInfo := ctxData.(models.SaveTrendInfo)

	// mysql 保存动态信息
	err := mysql.SaveTrendInfo(&trendInfo)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 保存动态信息
	err = mongo.SaveTrendInfo(trendInfo)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 给自己的feed 流添加一条
	err = mongo.SetTheUserFeed([]int64{trendInfo.TrendId}, "tr", trendInfo.UserId, trendInfo.UserId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 如果存在图片
	if trendInfo.PicCount != 0 {
		//发送 压缩消息到服务器
		err = cache.SendCompressQueue(trendInfo.UserId, trendInfo.TrendId, "tr")
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
	}

	formatData := models.TrendInfo{
		TrendId:     trendInfo.TrendId,
		Comment:     trendInfo.Comment,
		WhoSee:      trendInfo.WhoSee,
		UserId:      trendInfo.UserId,
		Pics:        trendInfo.Pics,
		Intro:       trendInfo.Text,
		Avatar:      "",
		Type:        "tr",
		UserName:    "",
		ForwardInfo: trendInfo.ForwardInfo,
		Topic:       trendInfo.Topic,
		Count:       models.TrendCount{},
		Interact:    models.UserTrendInteract{},
		CreateAT:    trendInfo.CreateAt,
		IsDelete:    false,
	}

	ResponseSuccess(ctx, formatData)

	ctx.Set("feed", models.UploadArtOrTrend{
		MsgID:  trendInfo.TrendId,
		SendId: trendInfo.UserId,
		Type:   "tr",
	})
}
