package cacheMiddle

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	c "onpaper-api-go/cache"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"time"
)

// SetArtworkProfile 设置单个作品页的展示信息缓存
func SetArtworkProfile(ctx *gin.Context) {
	// 1. 获取传递的数据
	dataCtx, _ := ctx.Get("artInfo")
	artInfoCtx := dataCtx.(m.ShowArtworkInfo)

	// 设置作品资料缓存
	err := c.SetArtworkProfile(artInfoCtx)
	if err != nil {
		logger.ErrZapLog(err, artInfoCtx.ArtworkId)
	}

	// 设置作品统计数据缓存
	artworkCount := map[string]interface{}{
		"Likes":    artInfoCtx.Likes,
		"Collects": artInfoCtx.Collects,
		"Comments": artInfoCtx.Comments,
		"Forwards": artInfoCtx.Forwards,
	}
	key := fmt.Sprintf(c.ArtworkCount, artInfoCtx.ArtworkId)
	// 3天后过期 避免用户打开页面 不操作过了几小时后点赞 此时没有缓存 会导致数据出错
	err = c.BatchSetHashValue(key, artworkCount, 3*24)
	if err != nil {
		logger.ErrZapLog(err, artInfoCtx.ArtworkId)
	}
	// 单独设置 Views 的值 因为数据库 Views 比 hash缓存新
	err = c.SetHashKeyValue(key, map[string]interface{}{"Views": artInfoCtx.Views}, 3*24)
	if err != nil {
		logger.ErrZapLog(err, artInfoCtx.ArtworkId)
	}

	// 设置作者点赞收藏数缓存
	err = c.CheckUserCount(artInfoCtx.UserId)
	if err != nil {
		logger.ErrZapLog(err, fmt.Sprintf("CheckUserCount %s fail", artInfoCtx.UserId))
	}
}

// GetArtworkProfile 获取作品页资料缓存
func GetArtworkProfile(ctx *gin.Context) {
	// 1. 获取参数
	data, _ := ctx.Get("artworkId")
	artId := data.(string)

	// 2. 通过参数到缓存查询
	res, err := c.GetArtworkProfile(artId, ctx.ClientIP())
	if err != nil {
		logger.ErrZapLog(err, artId)
	}

	// 作品统计数据
	artCount := res[1].(*redis.MapStringStringCmd).Val()
	// 作品资料数据
	val := res[0].(*redis.StringCmd).Val()
	view := res[3].(*redis.IntCmd).Val()
	// 3.  反序列化
	var temp m.ShowArtworkInfo
	var artInfo c.CtxCacheVale
	var countInfo c.CtxCacheVale

	if val != "" {
		err = json.Unmarshal([]byte(val), &temp)
		if err != nil {
			err = errors.Wrap(err, "GetArtworkProfile Unmarshal fail ")
			logger.ErrZapLog(err, val)
		} else {
			temp.Views = int(view) + temp.Views

			artInfo.Val = temp
			artInfo.HaveCache = true
		}
	}
	// 如果作品统计没过期
	if len(artCount) != 0 {
		countInfo.Val = artCount
		countInfo.HaveCache = true
	}

	ctx.Set("artInfo", artInfo)
	ctx.Set("artCount", countInfo)
	ctx.Set("views", int(view))

}

// SetArtworkRank 设置作品排行缓存
func SetArtworkRank(ctx *gin.Context) {
	data, _ := ctx.Get("rankType")
	queryData, _ := data.(m.QueryArtworkRank)

	data, _ = ctx.Get("artworkRank")
	artworks, _ := data.([]m.ArtworkRank)

	err := c.SetArtworkRank(queryData.RankType, artworks)
	if err != nil {
		logger.ErrZapLog(err, queryData.RankType)
	}
}

// GetArtworkRank 获取作品排行缓存
func GetArtworkRank(ctx *gin.Context) {
	data, _ := ctx.Get("rankType")
	queryData, _ := data.(m.QueryArtworkRank)
	key := fmt.Sprintf(c.RankArtwork, queryData.RankType)

	// 反序列化
	var temp []m.ArtworkRank
	artworks, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, queryData.RankType)
	}

	ctx.Set("rankData", artworks)
}

// SetCollectCount 设置作品收藏数据
func SetCollectCount(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	userData, _ := ctx.Get("userInfo")
	data, _ := ctx.Get("interact")

	collectData, _ := data.(*m.PostInteractData)
	userInfo, _ := userData.(m.UserTokenPayload)

	err := c.SetCollectCount(userInfo.Id, *collectData)
	if err != nil {
		logger.ErrZapLog(err, collectData)
	}
	if !collectData.IsCancel {
		cSortKey := fmt.Sprintf(c.UserCollect, userInfo.Id)
		err = c.CheckZSortLen(cSortKey, 1000, 100)
		if err != nil {
			logger.ErrZapLog(err, "")
		}
	}
}

// SetLikeCount 设置作品点赞数据
func SetLikeCount(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	userData, _ := ctx.Get("userInfo")
	data, _ := ctx.Get("interact")

	likeData, _ := data.(*m.PostInteractData)
	userInfo, _ := userData.(m.UserTokenPayload)

	err := c.SetLikeCount(userInfo.Id, *likeData)
	if err != nil {
		logger.ErrZapLog(err, likeData)
	}

	// 如果是添加点赞 维护ZSort
	if !likeData.IsCancel {
		cSortKey := fmt.Sprintf(c.UserLike, userInfo.Id)
		err = c.CheckZSortLen(cSortKey, 1000, 100)
		if err != nil {
			logger.ErrZapLog(err, "")
		}
	}

}

// BatchSetBasicArt 批量设置单个基本的作品数据（经常用于首次展示）
func BatchSetBasicArt(ctx *gin.Context) {
	ctxData, _ := ctx.Get("artData")
	artData := ctxData.([]m.BasicArtwork)

	var artIds []string
	for _, art := range artData {
		artIds = append(artIds, art.ArtworkId)
	}

	fmtKey := fmt.Sprintf(c.ArtworkBasic, "%s")
	err := c.BatchSetTypeOfString(fmtKey, artIds, artData, time.Hour*24)
	if err != nil {
		logger.ErrZapLog(err, "BatchSetBasicArt fail")
	}
}

// BatchSetArtViews 设置作品浏览量
func BatchSetArtViews(ctx *gin.Context) {
	ctxData, _ := ctx.Get("viewIds")
	artIds := ctxData.([]string)

	err := c.BatchSetArtViews(artIds, ctx.ClientIP())
	if err != nil {
		logger.ErrZapLog(err, "BatchSetArtViews fail")
	}
}
