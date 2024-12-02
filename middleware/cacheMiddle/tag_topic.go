package cacheMiddle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	c "onpaper-api-go/cache"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
	"time"
)

// SetTagRelevant 设置tag 相关tags数据
func SetTagRelevant(ctx *gin.Context) {
	ctxData, _ := ctx.Get("tagData")
	tagValue := ctxData.(m.TagRelevant)

	ctxData, _ = ctx.Get("nowTag")
	queryData := ctxData.(m.TagQueryParam)

	err := c.SetTagRelevant(tagValue, queryData.TagId)
	if err != nil {
		logger.ErrZapLog(err, queryData.TagName)
	}
}

// GetRelevantTags 获取tag 相关tags数据
func GetRelevantTags(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryParam)

	key := fmt.Sprintf(c.TagRelevant, queryData.TagId)

	// 反序列化
	var temp m.TagRelevant

	tagCtx, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, queryData)
	}

	ctx.Set("tagCtx", tagCtx)

}

// SetTagUser 设置tag 相关users数据
func SetTagUser(ctx *gin.Context) {
	ctxData, _ := ctx.Get("tagUser")
	tagUser := ctxData.([]m.UserBigCard)

	ctxData, _ = ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryParam)

	err := c.SetTagUser(tagUser, queryData.TagId)
	if err != nil {
		logger.ErrZapLog(err, queryData.TagName)
	}

}

// GetTagUser 获取tag 相关users数据
func GetTagUser(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryParam)

	key := fmt.Sprintf(c.TagUser, queryData.TagId)
	var temp []m.UserBigCard

	userCtx, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, queryData)
	}

	ctx.Set("userCtx", userCtx)
}

// SetTagArtwork 设置tag 对应的作品缓存
func SetTagArtwork(ctx *gin.Context) {
	ctxData, _ := ctx.Get("ArtIdAndUid")
	dataList := ctxData.([]m.ArtIdAndUid)

	ctxData, _ = ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryArtParam)

	// 设置缓存
	err := c.SetTagArtId(queryData.TagId, strconv.Itoa(int(queryData.Page)), queryData.Sort, dataList)
	if err != nil {
		logger.ErrZapLog(err, queryData.TagName)
	}
}

// GetTagArtwork 获取tag 对应的作品缓存
func GetTagArtwork(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryArtParam)

	end := fmt.Sprintf("%d&%s", queryData.Page, queryData.Sort)
	key := fmt.Sprintf(c.TagArtworkAndPage, queryData.TagId, end)

	var temp []m.ArtIdAndUid

	dataCtx, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, queryData)
	}

	ctx.Set("ArtIdAndUid", dataCtx)
}

// SetTagHotRank 设置热门tag 缓存
func SetTagHotRank(ctx *gin.Context) {
	ctxData, _ := ctx.Get("hotTag")

	err := c.SetHotTag(ctxData.([]m.HotTagRank))
	if err != nil {
		logger.ErrZapLog(err, "")
	}

}

// SetTopUseTagRank 设置使用最多的tag缓存
func SetTopUseTagRank(ctx *gin.Context) {
	ctxData, _ := ctx.Get("tagCache")

	key := fmt.Sprintf(c.RankTopUseTag)
	err := c.SetOneStringValue(key, ctxData, 3*time.Hour)
	if err != nil {
		err = errors.Wrap(err, "SetTopUseTagRank fail")
	}
}

// GetHotTagRank 获取热门tag缓存
func GetHotTagRank(ctx *gin.Context) {
	key := fmt.Sprintf(c.RankTag, "hours")
	var temp []m.HotTagRank
	ctxVal, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, "GetHotTagRank cache fail")
	}

	ctx.Set("hotCache", ctxVal)
}

// GetTopUseTagRank 获取使用最多的Tag缓存
func GetTopUseTagRank(ctx *gin.Context) {
	var temp []m.SearchTagResult
	ctxVal, err := c.GetOneStringValue(c.RankTopUseTag, temp)
	if err != nil {
		logger.ErrZapLog(err, "GetTopUseTagRank cache fail")
	}

	ctx.Set("topUseCache", ctxVal)
}

// GetHotTopicRank 获取热门Topic缓存
func GetHotTopicRank(ctx *gin.Context) {
	key := fmt.Sprintf(c.RankTopic, "hours")
	var temp []m.HotTopicRank
	ctxVal, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, "GetHotTopicRank cache fail")
	}

	ctx.Set("hotCache", ctxVal)
}

// SetHotTopicRank 设置热门topic 缓存
func SetHotTopicRank(ctx *gin.Context) {
	ctxData, _ := ctx.Get("hotTag")

	err := c.SetHotTopic(ctxData.([]m.HotTopicRank))
	if err != nil {
		logger.ErrZapLog(err, "")
	}

}

// SetTopicDetail 设置话题详情缓存
func SetTopicDetail(ctx *gin.Context) {
	ctxData, _ := ctx.Get("detail")

	err := c.SetTopicDetail(ctxData.(m.TopicDetail))
	if err != nil {
		logger.ErrZapLog(err, "")
	}
}

// GetTopicDetail 获取话题详情缓存
func GetTopicDetail(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TopicQueryParam)

	key := fmt.Sprintf(c.TopicProfile, queryData.TopicId)
	var temp m.TopicDetail
	ctxVal, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, "GetTopicDetail cache fail")
	}

	ctx.Set("detailCache", ctxVal)
}
