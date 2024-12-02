package controller

import (
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	c "onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
)

// GetTagArtwork 获取tag 对应的作品
func GetTagArtwork(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryArtParam)

	ctxData, _ = ctx.Get("ArtIdAndUid")
	cacheData := ctxData.(c.CtxCacheVale)

	var dataList []m.ArtIdAndUid
	if cacheData.HaveCache {
		dataList = cacheData.Val.([]m.ArtIdAndUid)
	} else {
		res, err := mysql.GetTagArtworkId(queryData.TagId, queryData.Sort, queryData.Page-1)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		dataList = res
	}

	dataLen := len(dataList)
	//如果没有找到相关作品 直接返回
	if dataLen == 0 {
		ResponseError(ctx, CodeArtworkNoExists)
		ctx.Abort()
		return
	}

	artIds := make([]string, 0, dataLen)
	userIds := make([]string, 0, dataLen)
	for _, data := range dataList {
		artIds = append(artIds, data.ArtworkId)
		userIds = append(userIds, data.AuthorId)
	}

	var eg errgroup.Group

	var artData []m.BasicArtwork
	var findData []m.BasicArtwork
	eg.Go(func() (mErr error) {
		artData, findData, mErr = BatchGetBasicArtInfo(artIds, userIds)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetTagArtwork BatchGetBasicArtInfo fail")
		}
		return
	})

	var cacheCountMap map[string]map[string]string
	var needCacheCount []m.ArtworkCount
	eg.Go(func() (mErr error) {
		cacheCountMap, needCacheCount, mErr = BatchGetArtworkCount(artIds)
		if mErr != nil {
			logger.ErrZapLog(mErr, "GetTagArtwork BatchGetArtworkCount fail")
		}
		return
	})

	// 查找是否点赞过
	isLikeMap := map[string]bool{}
	eg.Go(func() (mErr error) {
		// 没有登录直接返回 不需要查询
		if loginUser.Id == "" {
			return
		}
		isLikeMap, mErr = c.CheckUserLike(artIds, loginUser.Id)
		if mErr != nil {
			logger.ErrZapLog(mErr, "GetTagArtwork CheckUserLike fail ")
		}
		// 出错就算了
		return nil
	})

	err := eg.Wait()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
	}

	// 设置作品流量量
	var viewIds []string
	var resData []m.BasicArtworkAndLike
	// 拼接数据
	for _, artwork := range artData {
		var temp m.BasicArtworkAndLike
		str, _ := json.Marshal(artwork)
		_ = json.Unmarshal(str, &temp)

		if count, ok := cacheCountMap[artwork.ArtworkId]; ok {
			likes, _ := strconv.Atoi(count["Likes"])
			temp.Likes = likes
		}

		if like, ok := isLikeMap[artwork.ArtworkId]; ok {
			temp.IsLike = like
		}

		resData = append(resData, temp)
		viewIds = append(viewIds, temp.ArtworkId)
	}

	ResponseSuccess(ctx, resData)

	err = c.SetArtworkCount(needCacheCount)
	if err != nil {
		logger.ErrZapLog(err, "SetArtworkCount fail")
	}

	ctx.Set("viewIds", viewIds)
	ctx.Set("artData", findData)
	ctx.Set("ArtIdAndUid", dataList)
}

// GetRelevantTags 获取和查询tag 相关的tags
func GetRelevantTags(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryParam)

	ctxData, _ = ctx.Get("tagCtx")
	cacheData := ctxData.(c.CtxCacheVale)

	if cacheData.HaveCache {
		ResponseSuccess(ctx, cacheData.Val)
		ctx.Abort()
		return
	}

	var eg errgroup.Group
	var res m.TagRelevant
	var tags []m.ArtworkTag
	eg.Go(func() (err error) {
		tags, err = mysql.GetRelevantTags(queryData.TagId)
		if err != nil {
			err = errors.Wrap(err, "GetRelevantTags fail")
		}
		return
	})

	eg.Go(func() (err error) {
		res, err = mysql.GetTagArtCount(queryData.TagId)
		if err != nil {
			err = errors.Wrap(err, "GetRelevantTags fail")
		}
		return
	})

	err := eg.Wait()
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			res.Tags = make([]m.ArtworkTag, 0, 0)
			ctx.Abort()
		} else {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
	}
	res.Tags = tags
	ResponseSuccess(ctx, res)
	ctx.Set("tagData", res)
	ctx.Set("nowTag", queryData)
}

// GetRelevantUsers 获取和查询tag 相关的users
func GetRelevantUsers(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryParam)

	ctxData, _ = ctx.Get("userInfo")
	tokenInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("userCtx")
	cacheUser := ctxData.(c.CtxCacheVale)

	var userData []m.UserBigCard

	if cacheUser.HaveCache {
		userData = cacheUser.Val.([]m.UserBigCard)
		ctx.Abort()
	} else {
		res, err := mysql.GetRelevantUser(queryData.TagId)
		if err != nil {
			err = errors.Wrap(err, "GetRelevantTags fail")
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		userData = res
	}

	if len(userData) == 0 {
		ResponseSuccess(ctx, userData)
		ctx.Abort()
		return
	}

	// 如果是登录用户 到缓存查找关注列表 查看是否在排名里有关注过的
	if tokenInfo.Id != "" {
		var uIds []string
		for _, u := range userData {
			uIds = append(uIds, u.UserId)
		}
		focusIdMap, mErr := c.CheckUserFollow(uIds, tokenInfo.Id)
		if mErr != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, mErr)
			return
		}

		// 查询是否存在关注的
		for i, ui := range userData {
			userData[i].IsFocus = focusIdMap[ui.UserId]
		}
	}

	ResponseSuccess(ctx, userData)

	ctx.Set("tagUser", userData)
}

// GetHotTagRank 获取热门tag
func GetHotTagRank(ctx *gin.Context) {
	ctxData, _ := ctx.Get("hotCache")
	hotTagCache := ctxData.(c.CtxCacheVale)

	if hotTagCache.HaveCache {
		ResponseSuccess(ctx, hotTagCache.Val)
		ctx.Abort()
		return
	}

	res, err := mysql.GetTagHotRank()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, res)
	ctx.Set("hotTag", res)
}

// GetTopUseTag 获取使用最多的tag
func GetTopUseTag(ctx *gin.Context) {
	ctxData, _ := ctx.Get("topUseCache")
	topUseTagCache := ctxData.(c.CtxCacheVale)

	if topUseTagCache.HaveCache {
		ResponseSuccess(ctx, topUseTagCache.Val)
		ctx.Abort()
		return
	}

	res, err := mysql.GetTopUseTag()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, res)
	ctx.Set("tagCache", res)
}

// GetLikeNameTag 按tagName 查找 like%
func GetLikeNameTag(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TagQueryParam)
	if queryData.TagId != "1001" {
		ResponseError(ctx, CodeParamsError)
		return
	}

	likeData, searchData, err := mysql.SearchTagName(queryData.TagName)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"likeData":   likeData,
		"searchData": searchData,
	})
}

// GetRelevantTopic 模糊查询相关topic
func GetRelevantTopic(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TopicQueryParam)

	topics, err := mysql.SearchRelevantTopic(queryData.TopicName)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	if len(topics) == 0 {
		topics = make([]m.SearchTopicType, 0)
	}
	ResponseSuccess(ctx, topics)
}

func GetTopicTrend(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	query := ctxData.(m.TopicQueryTrendParam)

	ctxData, _ = ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	idInfo, err := mongo.GetTopicTrend(query.TopicId, query.Sort, query.Page-1)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	if len(idInfo) == 0 {
		ResponseError(ctx, CodeTrendNoExists)
		ctx.Abort()
		return
	}

	var trendId []string
	var userId []string
	var findCount []m.MongoFeed
	for _, d := range idInfo {
		trendId = append(trendId, strconv.FormatInt(d.TrendId, 10))
		userId = append(userId, d.UserId)
		findCount = append(findCount, m.MongoFeed{MsgID: d.TrendId, Type: "tr"})
	}

	cacheTrends, needFind, err := c.GetBatchTrendCache(trendId)
	if err != nil {
		logger.ErrZapLog(err, "GetTopicTrend GetBatchTrendCache fail")
	}

	var needFindId []int64
	for _, d := range needFind {
		intId, _ := strconv.ParseInt(d.Id, 10, 64)
		needFindId = append(needFindId, intId)
	}

	findData, _, err := GetTrendInfo(needFindId, "tr")
	if err != nil {
		err = errors.Wrap(err, "GetTopicTrend GetTrendInfo fail")
	}

	// 统计数据
	countMap, err := BatchGetTrendCount(findCount)
	if err != nil {
		err = errors.Wrap(err, "GetTopicTrend GetTrendCount fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 用户数据
	userMap, err := mysql.GetBatchUserSimpleInfo(userId)
	if err != nil {
		err = errors.Wrap(err, "GetTopicTrend GetBatchUserSimpleInfo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	likeMap := map[string]bool{}
	// 如果是登录用户
	if userInfo.Id != "" {
		likeMap, err = c.CheckUserLike(trendId, userInfo.Id)
		if err != nil {
			logger.ErrZapLog(err, "GetTopicTrend CheckUserLike fail ")
		}
	}
	focusMap := map[string]uint8{}
	// 拼接数据
	FormatTrendData(cacheTrends, userMap, likeMap, focusMap, countMap)
	FormatTrendData(findData, userMap, likeMap, focusMap, countMap)

	res := append(cacheTrends, findData...)

	ResponseSuccess(ctx, res)
	ctx.Set("trendData", findData)
}

// GetTopicDetail 查询话题详情
func GetTopicDetail(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.TopicQueryParam)

	ctxData, _ = ctx.Get("detailCache")
	cacheData := ctxData.(c.CtxCacheVale)

	if cacheData.HaveCache {
		ResponseSuccess(ctx, cacheData.Val)
		ctx.Abort()
		return
	}

	detail, err := mysql.GetTopicDetail(queryData.TopicId)
	if err != nil {
		if errors.Cause(err) != sql.ErrNoRows {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		// 如果是 没有找到 返回null
		ResponseSuccess(ctx, nil)
		ctx.Abort()
		return
	}

	ResponseSuccess(ctx, detail)
	ctx.Set("detail", detail)
}

// GetHotTopicRank 获取热门tag
func GetHotTopicRank(ctx *gin.Context) {
	ctxData, _ := ctx.Get("hotCache")
	hotTagCache := ctxData.(c.CtxCacheVale)

	if hotTagCache.HaveCache {
		ResponseSuccess(ctx, hotTagCache.Val)
		ctx.Abort()
		return
	}

	res, err := mysql.GetTopicHotRank()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, res)
	ctx.Set("hotTag", res)
}
