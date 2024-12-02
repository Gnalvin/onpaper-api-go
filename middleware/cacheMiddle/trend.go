package cacheMiddle

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	c "onpaper-api-go/cache"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
	"time"
)

// SetTrendDetail 设置Trend详情缓存
func SetTrendDetail(ctx *gin.Context) {
	ctxData, _ := ctx.Get("isHaveCache")
	isHaveCache, _ := ctxData.(bool)

	ctxData, _ = ctx.Get("trendData")
	trendData := ctxData.(m.TrendShowInfo)

	var needCacheCount []m.TrendShowInfo

	// 主体没有缓存 添加缓存
	if !isHaveCache {
		key := fmt.Sprintf(c.TrendProfile, strconv.FormatInt(trendData.TrendId, 10))
		err := c.SetOneStringValue(key, trendData, time.Hour*2*24)
		if err != nil {
			err = errors.Wrap(err, "SetTagRelevant fail")
			logger.ErrZapLog(err, trendData.TrendId)
		}
		needCacheCount = append(needCacheCount, trendData)
	}
	// 只要有转发都设置 count
	if trendData.Forward != nil {
		var temp m.TrendShowInfo
		// 通过json 拷贝对应的数据
		str, _ := json.Marshal(trendData.Forward)
		_ = json.Unmarshal(str, &temp)
		needCacheCount = append(needCacheCount, temp)
	}

	// 设置统计数
	err := c.SetTrendCount(needCacheCount)
	if err != nil {
		logger.ErrZapLog(err, "SetTrendDetail  SetTrendCount fail")
	}
	// 如果是作品设置 浏览量
	if trendData.Type == "aw" {
		err = c.BatchSetArtViews([]string{strconv.FormatInt(trendData.TrendId, 10)}, ctx.ClientIP())
		if err != nil {
			logger.ErrZapLog(err, "SetTrendDetail  BatchSetArtViews fail")
		}
	}
}

// BatchSetTrend 批量设置trend 缓存
func BatchSetTrend(ctx *gin.Context) {
	ctxData, _ := ctx.Get("trendData")
	trendData := ctxData.(m.TrendList)

	var sets []string
	for i, data := range trendData {
		str := strconv.FormatInt(data.TrendId, 10)
		sets = append(sets, str)
		// 转发的数据不缓存
		trendData[i].Forward = nil
	}
	// 动态主体缓存
	fmtKey := fmt.Sprintf(c.TrendProfile, "%s")
	err := c.BatchSetTypeOfString(fmtKey, sets, trendData, time.Hour*24*2)
	if err != nil {
		logger.ErrZapLog(err, "BatchSetTrend fail")
	}

	// 设置统计数
	err = c.SetTrendCount(trendData)
	if err != nil {
		logger.ErrZapLog(err, "BatchSetTrend  SetTrendCount fail")
	}
}
