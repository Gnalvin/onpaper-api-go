package handleMiddle

import (
	"fmt"
	"github.com/pkg/errors"
	ctl "onpaper-api-go/controller"
	"onpaper-api-go/models"
	"onpaper-api-go/settings"
	"onpaper-api-go/utils/formatTools"
	"onpaper-api-go/utils/oss"
	"onpaper-api-go/utils/snowflake"
	"onpaper-api-go/utils/verify"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// HandleAvatarInfo 验证头像信息
func HandleAvatarInfo(ctx *gin.Context) {
	var data models.CallBackFileInfo
	// 取出上传的 json data 信息到 key
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		//绑定失败 参数错误
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}

	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(models.UserTokenPayload)

	path := "avatars/" + userInfo.Id + "/" + data.FileName
	// 查询文件是否存在
	fileInfo, err := oss.SelectOssFileInfo(settings.Conf.TempBucket, path)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("头像不存在,%s", path))
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	hSize := fileInfo.Get("Content-Length")
	size, _ := strconv.ParseInt(hSize, 10, 64)
	// 如果图片大于 5M 报错
	if size > 5*1024*1024 {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}
	data.Size = size
	data.Type = fileInfo.Get("Content-Type")

	//移动图片到原始库
	err = oss.MoveTempToOriginal(path)
	if err != nil {
		err = errors.Wrap(err, "MoveTempToOriginal fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}
	// 把它传递到上下文
	ctx.Set("fileInfo", data)
}

// HandleBannerInfo 验证头像信息
func HandleBannerInfo(ctx *gin.Context) {
	var data models.CallBackFileInfo
	// 取出上传的 json data 信息到 key
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		//绑定失败 参数错误
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}

	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(models.UserTokenPayload)

	path := "banners/" + userInfo.Id + "/" + data.FileName
	// 查询文件是否存在
	fileInfo, err := oss.SelectOssFileInfo(settings.Conf.TempBucket, path)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("背景不存在,%s", path))
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	hSize := fileInfo.Get("Content-Length")
	size, _ := strconv.ParseInt(hSize, 10, 64)
	// 如果图片大于 15M 报错
	if size > 15*1024*1024 {
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}
	data.Size = size
	data.Type = fileInfo.Get("Content-Type")

	//移动图片到原始库
	err = oss.MoveTempToOriginal(path)
	if err != nil {
		err = errors.Wrap(err, "MoveTempToOriginal fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}
	// 把它传递到上下文
	ctx.Set("fileInfo", data)
}

// HandleArtworkInfo 查询cos文件信息,保存作品信息
func HandleArtworkInfo(ctx *gin.Context) {
	// -----------------1.取数 ---------------
	// 取出 ctx 传递的数据
	ctxData, exi := ctx.Get("userInfo")
	// 如果取不到数据
	if !exi {
		err := errors.New("SaveProfileData get ctxData fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}
	// 类型断言 得到 payload 里面的用户id
	userInfo, ok := ctxData.(models.UserTokenPayload)
	// 如果断言错误
	if !ok {
		err := errors.New("SaveProfileData assert fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	var data models.CallBackArtworkInfo
	// 取出上传的 json  信息到 data
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		//绑定失败 参数错误
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}
	// -----------------2.验证文本 ---------------
	//标体不超过15个字 描述不超过350个字
	isPass := verify.ArtTextInfo(data.Title, data.Description, data.Tags)
	if !isPass {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	fileListLen := len(data.FileList)
	// 文件个数大于0不超过15
	if fileListLen > 15 || fileListLen == 0 {
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

	//验证封面信息是否正确
	key := "artworks/" + userInfo.Id + "/" + data.Cover
	err = oss.MoveTempToOriginal(key)
	if err != nil {
		err = errors.Wrap(err, "MoveTempToOriginal fail")
		ctl.ResponseErrorAndLog(ctx, ctl.CodeServerBusy, err)
		return
	}

	// -----------------3.cos验证文件是否存在 ---------------
	// 存放循环查询后的 文件信息
	var fileList []*models.PicsType
	var firstPic string
	for _, file := range data.FileList {
		key = "artworks/" + userInfo.Id + "/" + file.FileName
		mErr := oss.MoveTempToOriginal(key)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "MoveTempToOriginal fail: "+key)
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

		var fileInfo = &models.PicsType{
			FileName: file.FileName,
			Mimetype: contentType,
			Size:     size,
			Sort:     file.Sort,
			Width:    file.Width,
			Height:   file.Height,
		}
		fileList = append(fileList, fileInfo)
		//  保持首张图片名
		if file.Sort == 0 {
			firstPic = file.FileName
		}
	}

	// 去重
	tags, _ := formatTools.RemoveSliceDuplicate(data.Tags)

	// 构建需要保存的作品信息
	var artworkInfo = &models.SaveArtworkInfo{
		ArtworkId:   snowflake.CreateID(),
		UserId:      userInfo.Id,
		FileList:    fileList,
		FirstPic:    firstPic,
		Title:       data.Title,
		Description: data.Description,
		Tags:        tags,
		Zone:        data.Zone,
		WhoSee:      data.WhoSee,
		Adults:      data.Adult,
		Cover:       data.Cover,
		Comment:     data.Comment,
		CopyRight:   data.CopyRight,
		Device:      data.Device,
	}
	// 把它传递到上下文
	ctx.Set("artworkInfo", artworkInfo)
}

// HandleTrendInfo 验证trend 信息
func HandleTrendInfo(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	// 类型断言 得到 payload 里面的用户id
	userInfo, _ := ctxData.(models.UserTokenPayload)

	var data models.CallBackTrendInfo
	// 取出上传的 json  信息到 data
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		//绑定失败 参数错误
		ctl.ResponseErrorAndLog(ctx, ctl.CodeJsonFormatError, err)
		return
	}

	//空数据
	if data.Text == "" && len(data.FileList) == 0 && data.ForwardInfo.Id == "0" {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 不能超过9张图片
	if len(data.FileList) > 9 {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	//------------------2.验证相关参数 ---------------------
	//描述不超过350个字
	descriptionLen := utf8.RuneCountInString(data.Text)
	isPass := descriptionLen <= 350
	if !isPass {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ForwardId, err := strconv.ParseInt(data.ForwardInfo.Id, 10, 64)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	// topic 长度不超过 25
	topicLen := utf8.RuneCountInString(data.Topic.Text)
	isPass = topicLen <= 25
	if !isPass {
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

	// -----------------3.cos验证文件是否存在 ---------------
	// 存放循环查询后的 文件信息
	var fileList []models.PicsType
	for _, file := range data.FileList {
		key := "trends/" + userInfo.Id + "/" + file.FileName
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

		var fileInfo = models.PicsType{
			FileName: file.FileName,
			Mimetype: contentType,
			Size:     size,
			Sort:     file.Sort,
			Width:    file.Width,
			Height:   file.Height,
		}
		fileList = append(fileList, fileInfo)
	}

	//生成用户id
	trendId := snowflake.CreateID()
	nowTime := time.Now()
	// 构建需要保存的作品信息
	var trendInfo = models.SaveTrendInfo{
		TrendId:  trendId,
		UserId:   userInfo.Id,
		Text:     data.Text,
		Topic:    data.Topic,
		Pics:     fileList,
		Comment:  data.Comment,
		WhoSee:   data.WhoSee,
		PicCount: uint8(len(data.FileList)),
		UpdateAt: nowTime,
		CreateAt: nowTime,
		State:    0,
		IsDelete: false,
		ForwardInfo: models.ForwardInfo{
			Id:   ForwardId,
			Type: data.ForwardInfo.Type,
		},
	}

	// 把它传递到上下文
	ctx.Set("trendInfo", trendInfo)
}
