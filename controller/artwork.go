package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/singleFlight"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// GetOneArtworkInfo 返回单个作品信息
func GetOneArtworkInfo(ctx *gin.Context) {
	// 获取 ctx 数据
	dataCtx, _ := ctx.Get("userInfo")
	userInfo, _ := dataCtx.(m.UserTokenPayload)

	data, _ := ctx.Get("artworkId")
	artId := data.(string)

	dataCtx, _ = ctx.Get("artInfo")
	artInfoCtx := dataCtx.(cache.CtxCacheVale)

	dataCtx, _ = ctx.Get("artCount")
	countCtx := dataCtx.(cache.CtxCacheVale)

	dataCtx, _ = ctx.Get("views")
	views := dataCtx.(int)

	var artInfo m.ShowArtworkInfo

	if artInfoCtx.HaveCache {
		// 1. 如果有缓存 缓存获取 并且不再执行下一个
		artInfo = artInfoCtx.Val.(m.ShowArtworkInfo)
		ctx.Abort()
	} else {
		// 2.没有缓存从数据库中获取
		ctxTimeOut, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		// 使用完整路径为 key
		key := (ctx.Request.URL).String()
		res, err := singleFlight.Do(ctxTimeOut, key, func() (interface{}, error) {
			artRes, err := mysql.GetOneArtwork(artId)
			return artRes, err
		})
		if err != nil {
			if errors.Cause(err) == sql.ErrNoRows {
				ResponseError(ctx, CodeArtworkNoExists)
				return
			}
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		artInfo = res.(m.ShowArtworkInfo)
		artInfo.Views = views + artInfo.Views // 数据库数据 + 缓存数据
	}

	// 是否是本人的作品
	isOwner := artInfo.UserId == userInfo.Id
	//如果是私密作品不是本人查看 返回没有作品
	if artInfo.WhoSee == "privacy" && !isOwner {
		ResponseError(ctx, CodeArtworkNoExists)
		return
	}

	//合成作品点赞收藏数据,没有缓存会从数据库查到 就不用合成
	if countCtx.HaveCache {
		artCount := countCtx.Val.(map[string]string)
		artInfo.Likes, _ = strconv.Atoi(artCount["Likes"])
		artInfo.Collects, _ = strconv.Atoi(artCount["Collects"])
		artInfo.Comments, _ = strconv.Atoi(artCount["Comments"])
		artInfo.Forwards, _ = strconv.Atoi(artCount["Forwards"])
	}

	// 查询用户互动数据
	var interact = new(m.UserArtworkInteract)
	// 如果 是登录状态
	if userInfo.Id != "" {
		// 查询缓存
		cmders, mErr := cache.GetUserToArtInteract(userInfo.Id, artInfo.ArtworkId)
		if mErr != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, mErr)
			return
		}
		// 返回的是时间戳，如果没有收藏或点赞 结果是 0
		interact.IsCollect = cmders[0].(*redis.FloatCmd).Val() != 0
		interact.IsLike = cmders[1].(*redis.FloatCmd).Val() != 0
		// 查相互关注
		focus, mErr := cache.CheckUserFollow([]string{artInfo.UserId}, userInfo.Id)
		if mErr != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, mErr)
			return
		}
		interact.IsFocusAuthor = focus[artInfo.UserId]
	}
	artInfo.Interact = *interact
	artInfo.IsOwner = isOwner
	// 返回数据
	ResponseSuccess(ctx, artInfo)
	//传递给设置缓存
	ctx.Set("artInfo", artInfo)
}

// SaveCollect 保存收藏
func SaveCollect(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	userData, _ := ctx.Get("userInfo")
	data, _ := ctx.Get("interact")

	collectData, _ := data.(*m.PostInteractData)
	userInfo, _ := userData.(m.UserTokenPayload)
	collectData.Action = "collect"

	// 设置登陆用户的count 缓存，之后要对收藏数 +-1
	err := cache.CheckUserCount(userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	var key string
	if collectData.Type == "aw" {
		key = fmt.Sprintf(cache.ArtworkCount, collectData.MsgId)
	} else {
		key = fmt.Sprintf(cache.TrendCount, collectData.MsgId)
	}

	// 如果缓存不存在 说明 超过3天在页面未操作 返回错误
	isExists, err := cache.CheckExistsKey(key)
	// 如果缓存不存在 或者 查询是否存在出错
	if isExists == 0 || err != nil {
		// 用户收藏页 有些作品无法显示时 需要强制 取消收藏 就不判断缓存
		if collectData.Force {
			mErr := cache.DeleteOneUserCollect(userInfo.Id, collectData.MsgId)
			if mErr != nil {
				logger.ErrZapLog(mErr, "DeleteOneUserCollect fail ")
			}
			ctx.Abort()
		} else {
			ResponseError(ctx, CodeLongTimeNoOperate)
			return
		}
	}

	//mongodb 数据库保存
	isChange, err := mongo.SetUserCollect(userInfo.Id, *collectData)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 如果没有实际更新 不对缓存操作
	if isChange == false {
		// 返回数据
		ResponseSuccess(ctx, collectData)
		ctx.Abort()
		return
	}

	// 返回数据
	ResponseSuccess(ctx, collectData)
}

// SaveUserLike 点赞作品
func SaveUserLike(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	userData, _ := ctx.Get("userInfo")
	data, _ := ctx.Get("interact")

	likeData, _ := data.(*m.PostInteractData)
	userInfo, _ := userData.(m.UserTokenPayload)
	likeData.Action = "like"

	//检测是否有count 缓存 没有则先添加缓存
	var err error
	if likeData.Type == "aw" {
		err = cache.CheckArtworkCount(likeData.MsgId)
	} else {
		err = cache.CheckTrendCount(likeData.MsgId)
	}
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// MongoDB 保存
	isChange, err := mongo.SetUserLike(userInfo.Id, *likeData)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 如果没有实际更新 不对缓存操作
	if isChange == false {
		ctx.Abort()
	}

	// 返回数据
	ResponseSuccess(ctx, likeData)
}

// GetArtworkRank 获取作品排名
func GetArtworkRank(ctx *gin.Context) {
	// 获取缓存数据
	dataCtx, _ := ctx.Get("rankData")
	rankData, _ := dataCtx.(cache.CtxCacheVale)
	if rankData.HaveCache {
		ResponseSuccess(ctx, rankData.Val.([]m.ArtworkRank))
		ctx.Abort()
		return
	}

	dataCtx, _ = ctx.Get("rankType")
	queryData, _ := dataCtx.(m.QueryArtworkRank)

	ctxTimeOut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 使用完整路径为 key
	key := (ctx.Request.URL).String()
	res, err := singleFlight.Do(ctxTimeOut, key, func() (res interface{}, err error) {
		artInfo, artCount, err := mysql.GetArtworkRank(queryData.RankType)
		if err != nil {
			return
		}
		// 拼接数据
		var artDate []m.ArtworkRank
		// 通过json 拷贝对应的数据
		str, _ := json.Marshal(artInfo)
		_ = json.Unmarshal(str, &artDate)

		for ia, artwork := range artDate {
			for ic, count := range artCount {
				if artwork.ArtworkId == count.ArtworkId {
					artDate[ia].Likes = count.Likes
					artDate[ia].Collects = count.Collects
					// 找到后删除 并跳出循环 减少循环次数
					artCount = append(artCount[:ic], artCount[ic+1:]...)
					break
				}
			}
		}

		return artDate, err
	})
	if err != nil {
		//数据库错误
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 返回数据
	artData := res.([]m.ArtworkRank)
	ResponseSuccess(ctx, artData)
	ctx.Set("artworkRank", artData)
}

// GetHotArtwork 查询热门作品数据
func GetHotArtwork(ctx *gin.Context) {
	userData, _ := ctx.Get("userInfo")
	loginInfo, _ := userData.(m.UserTokenPayload)

	artData, artIds, err := cache.GetHotArtwork()
	if err != nil {
		//数据库错误
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
	}

	// 登陆用户查找是否点赞过
	if loginInfo.Id != "" {
		likeMap, _err := cache.CheckUserLike(artIds, loginInfo.Id)
		if _err != nil {
			logger.ErrZapLog(err, "GetHotArtwork CheckUserLike fail")
		} else {
			for i, art := range artData {
				if like, ok := likeMap[art.ArtworkId]; ok {
					artData[i].IsLike = like
				}
			}
		}
	}

	// 返回数据
	ResponseSuccess(ctx, artData)

	ctx.Set("viewIds", artIds)
}

// GetChannelArtwork 查询最新作品数据
func GetChannelArtwork(ctx *gin.Context) {
	ctxData, _ := ctx.Get("query")
	queryData, _ := ctxData.(m.QueryChanelType)

	dataList, err := mysql.GetChannelArtwork(queryData)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	artIds := make([]string, 0)
	userIds := make([]string, 0)
	for _, data := range dataList {
		artIds = append(artIds, data.ArtworkId)
		userIds = append(userIds, data.AuthorId)
	}

	artData, findData, err := BatchGetBasicArtInfo(artIds, userIds)
	if err != nil {
		err = errors.Wrap(err, "GetArtFeed fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	if len(artData) == 0 {
		artData = make([]m.BasicArtwork, 0)
	}
	// 返回数据
	ResponseSuccess(ctx, artData)

	ctx.Set("artData", findData)
}

// UpdateArtInfo 更新作品信息
func UpdateArtInfo(ctx *gin.Context) {
	ctxData, _ := ctx.Get("artInfo")
	artInfo, _ := ctxData.(m.UpdateArtInfo)

	ctxData, _ = ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	err := mysql.UpdateArtInfo(artInfo)
	if err != nil {
		err = errors.Wrap(err, "UpdateArtInfo 更新错误")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	var keys []string
	//删除 作品封面 缓存
	artBasicsKey := fmt.Sprintf(cache.ArtworkBasic, artInfo.ArtworkId)
	keys = append(keys, artBasicsKey)
	//删除作品详情
	artProfileKey := fmt.Sprintf(cache.ArtworkProfile, artInfo.ArtworkId)
	keys = append(keys, artProfileKey)
	// 删除动态缓存
	trendProfileKey := fmt.Sprintf(cache.TrendProfile, artInfo.ArtworkId)
	keys = append(keys, trendProfileKey)
	// 用户大卡片资料
	userBigCardKey := fmt.Sprintf(cache.UserBigCard, userInfo.Id)
	keys = append(keys, userBigCardKey)

	err = cache.BatchDelCache(keys)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 如果权限下设置成 不是公开
	if artInfo.WhoSee != "public" {
		// 到热门中删除
		err = cache.DeleteOneHotArtwork(artInfo.ArtworkId)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		err = cache.DeleteOneHotTrend(artInfo.ArtworkId + "&aw&" + userInfo.Id)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
	}

	ResponseSuccess(ctx, artInfo)
	return
}

// DeleteArtwork 删除作品
func DeleteArtwork(ctx *gin.Context) {
	ctxData, _ := ctx.Get("artworkId")
	artId := ctxData.(string)

	ctxData, _ = ctx.Get("intArtId")
	intArtId := ctxData.(int64)

	ctxData, _ = ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	isOwner, err := mysql.VerifyArtOwner(userInfo.Id, artId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	if !isOwner {
		err = errors.New(fmt.Sprintf("VerifyArtOwner no auth user:%s,art:%s", userInfo.Id, artId))
		ResponseErrorAndLog(ctx, CodeUnPermission, err)
		return
	}

	err = mysql.DeleteArtwork(artId, userInfo.Id)
	if err != nil {
		err = errors.Wrap(err, "DeleteArtwork mysql 删除错误")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	err = mongo.DeleteOneFeed(intArtId, userInfo.Id, userInfo.Id)
	if err != nil {
		err = errors.Wrap(err, "DeleteArtwork mongo 删除错误")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	err = cache.DeleteAboutArt(userInfo.Id, artId)
	if err != nil {
		err = errors.Wrap(err, "DeleteAboutArt cache 删除错误")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, "删除成功")
	return
}

// GetHomePageZone 获取首页分区作品
func GetHomePageZone(ctx *gin.Context) {
	ctxData, _ := ctx.Get("zone")
	zoneIndex := ctxData.(int)

	artData, err := cache.GetHotZoneArt(strconv.Itoa(zoneIndex))
	if err != nil {
		//数据库错误
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
	}
	// 返回数据
	zoneList := []string{"插画", "同人", "古风", "日系", "场景", "原画", "头像", "Q版", "自设/OC"}
	ResponseSuccess(ctx, gin.H{
		"title":   zoneList[zoneIndex],
		"artwork": artData,
	})

	var artIds []string
	for _, art := range artData {
		artIds = append(artIds, art.ArtworkId)
	}
	ctx.Set("viewIds", artIds)
}
