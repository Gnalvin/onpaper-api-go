package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	c "onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	tools "onpaper-api-go/utils/formatTools"
	"sort"
	"strconv"
)

// GetArtFeed 获取关注的用户最近作品
func GetArtFeed(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("nextId")
	MsgId := ctxData.(*int64)

	dataList, err := mongo.GetFeed(*MsgId, userInfo.Id, "aw")
	if err != nil {
		err = errors.Wrap(err, "GetArtFeed mongo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	dataLen := len(dataList)
	if dataLen == 0 {
		// 返回数据
		ResponseSuccess(ctx, make([]struct{}, 0))
		ctx.Abort()
		return
	}

	artIds := make([]string, 0, dataLen)
	uIds := make([]string, 0, dataLen)
	for _, data := range dataList {
		artIds = append(artIds, strconv.FormatInt(data.MsgID, 10))
		uIds = append(uIds, data.SendId)
	}

	artData, findData, err := BatchGetBasicArtInfo(artIds, uIds)
	if err != nil {
		err = errors.Wrap(err, "GetArtFeed fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 清空不属于自己的私密作品 和已经删除的作品
	for i, data := range artData {
		if data.WhoSee == "privacy" && userInfo.Id != data.UserId || data.IsDelete {
			empty := m.BasicArtwork{}
			empty.WhoSee = data.WhoSee
			empty.IsDelete = data.IsDelete
			empty.IsOwner = userInfo.Id == data.UserId
			artData[i] = empty
		}
	}

	// 返回数据
	ResponseSuccess(ctx, artData)

	ctx.Set("artData", findData)
}

// SetRecentlyFeed 关注用户时 设置最近的动态到 关注人feed
func SetRecentlyFeed(ctx *gin.Context) {
	// 获取传递的参数
	ctxData, _ := ctx.Get("Focus")
	focusInfo := ctxData.(m.VerifyUserFocus)

	ctxData, _ = ctx.Get("userInfo")
	loginInfo := ctxData.(m.UserTokenPayload)

	if !focusInfo.IsCancel {
		// 最近作品
		artIds, err := mysql.GetUserRecentlyArtworkId(focusInfo.FocusId)
		if err != nil {
			logger.ErrZapLog(err, "SetTheUserFeed artIds fail")
		}
		// 最近动态
		trendIds, err := mongo.GetUserRecentlyTrendId(focusInfo.FocusId)
		if err != nil {
			logger.ErrZapLog(err, "SetTheUserFeed trendIds fail")
		}

		// 设置 作品feed
		err = mongo.SetTheUserFeed(artIds, "aw", focusInfo.FocusId, loginInfo.Id)
		if err != nil {
			logger.ErrZapLog(err, "SetTheUserFeed mongodb fail")
		}

		// 设置动态 feed
		err = mongo.SetTheUserFeed(trendIds, "tr", focusInfo.FocusId, loginInfo.Id)
		if err != nil {
			logger.ErrZapLog(err, "SetTheUserFeed mongodb fail")
		}

	} else {
		err := mongo.DelTheUserFeed(loginInfo.Id, focusInfo.FocusId)
		if err != nil {
			logger.ErrZapLog(err, "DelTheUserFeed mongodb fail")
		}
	}

	return
}

// GetAllFeed 获取所有类型的 feed
func GetAllFeed(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("nextId")
	MsgId := ctxData.(*int64)

	// 获取 feed
	feedData, err := mongo.GetFeed(*MsgId, userInfo.Id, "all")
	if err != nil {
		err = errors.Wrap(err, "GetAllFeed mongo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 如果没有feed 直接返回
	dataLen := len(feedData)
	if dataLen == 0 {
		// 返回数据
		ResponseSuccess(ctx, make([]struct{}, 0))
		ctx.Abort()
		return
	}
	ctx.Set("feedData", feedData)
}

func GetFeedTrend(ctx *gin.Context) {
	ctxData, _ := ctx.Get("feedData")
	feedData := ctxData.([]m.MongoFeed)

	ctxData, _ = ctx.Get("userInfo")
	loginInfo := ctxData.(m.UserTokenPayload)

	// 1. 到缓存中查找feed
	feedMap := make(map[string]m.MongoFeed)
	findId := make([]string, 0, len(feedData))
	findIsFocusId := make([]string, 0, len(feedData))
	for _, feed := range feedData {
		findIsFocusId = append(findIsFocusId, feed.SendId)
		feedMap[strconv.FormatInt(feed.MsgID, 10)] = feed
		findId = append(findId, strconv.FormatInt(feed.MsgID, 10))
	}

	cacheTrends, needFind, err := c.GetBatchTrendCache(findId)
	if err != nil {
		logger.ErrZapLog(err, "GetAllFeed GetBatchTrendCache fail")
	}

	var findArt []int64
	var findTrend []int64
	// 没有缓存的feed 到数据库查找
	for _, data := range needFind {
		feed := feedMap[data.Id]
		if feed.Type == "aw" {
			findArt = append(findArt, feed.MsgID)
		} else {
			findTrend = append(findTrend, feed.MsgID)
		}
	}

	// 2. 如果缓存有转发的 把转发的内容也查询一遍  redis缓存中只有动态主体 没有转发的数据
	var forwardIds []string
	forwardMap := make(map[string]m.ForwardInfo)
	for _, trend := range cacheTrends {
		if trend.ForwardInfo.Id == 0 {
			continue
		}
		forwardId := strconv.FormatInt(trend.ForwardInfo.Id, 10)
		//为后面查找是否点赞过准备 findId
		findId = append(findId, forwardId)
		//转发也查找count
		feedData = append(feedData, m.MongoFeed{MsgID: trend.ForwardInfo.Id, Type: trend.ForwardInfo.Type})
		// 转发缓存查找
		forwardIds = append(forwardIds, forwardId)
		forwardMap[forwardId] = trend.ForwardInfo
	}
	// 转发缓存查找
	forwardCache, needFindForward, err := c.GetBatchTrendCache(forwardIds)
	for _, f := range needFindForward {
		data := forwardMap[f.Id]
		if data.Type == "aw" {
			findArt = append(findArt, data.Id)
		} else {
			findTrend = append(findTrend, data.Id)
		}
	}

	var eg errgroup.Group

	findArtData := m.TrendList{}
	eg.Go(func() (mErr error) {
		findArtData, _, mErr = GetTrendInfo(findArt, "aw")
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetAllFeed GetTrendInfo fail")
		}
		return
	})

	findTrendData := m.TrendList{}
	noCacheForward := m.TrendList{}
	eg.Go(func() (mErr error) {
		findTrendData, noCacheForward, mErr = GetTrendInfo(findTrend, "tr")
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetAllFeed GetTrendInfo fail")
		}
		return
	})

	err = eg.Wait()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
	}

	// findData 包含 主题没有缓存的 和主题转发的动态数据
	findData := append(findArtData, findTrendData...)
	findData = append(findData, forwardCache...)

	// 查询用户信息
	var findUserInfo []string
	for _, data := range findData {
		findUserInfo = append(findUserInfo, data.UserId)
		// 转发里面的也查询
		if data.ForwardInfo.Id != 0 {
			findUserInfo = append(findUserInfo, data.Forward.UserId)
			//为后面查找是否点赞过准备 findId
			findId = append(findId, strconv.FormatInt(data.ForwardInfo.Id, 10))
			//转发也查找count
			feedData = append(feedData, m.MongoFeed{MsgID: data.Forward.TrendId, Type: data.Forward.Type})
		}
	}

	userMap := make(map[string]m.UserSimpleInfo)
	eg.Go(func() (mErr error) {
		// 查找用户数据
		userMap, mErr = mysql.GetBatchUserSimpleInfo(findUserInfo)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetAllFeed -> GetBatchUserSimpleInfo fail")
		}
		return
	})

	// 查找是否点赞过
	isLike := map[string]bool{}
	eg.Go(func() (mErr error) {
		// 没有登录直接返回 不需要查询
		if loginInfo.Id == "" {
			return
		}
		isLike, mErr = c.CheckUserLike(findId, loginInfo.Id)
		if mErr != nil {
			logger.ErrZapLog(mErr, "GetFeedTrend CheckUserLike fail ")
		}
		// 出错就算了
		return nil
	})

	// 查找是否关注过发布者
	isFocus := map[string]uint8{}
	eg.Go(func() (mErr error) {
		// 没有登录直接返回 不需要查询
		if loginInfo.Id == "" {
			return
		}
		isFocus, mErr = c.CheckUserFollow(findIsFocusId, loginInfo.Id)
		if mErr != nil {
			logger.ErrZapLog(mErr, "GetFeedTrend CheckUserFollow fail ")
		}
		// 出错就算了
		return nil
	})

	// 查找动态的统计数
	countMap := map[int64]map[string]string{}
	eg.Go(func() (mErr error) {
		countMap, mErr = BatchGetTrendCount(feedData)
		return
	})

	err = eg.Wait()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
	}

	// 拼接数据 点赞等数据
	FormatTrendData(cacheTrends, userMap, isLike, isFocus, countMap)
	FormatTrendData(findData, userMap, isLike, isFocus, countMap)
	FormatTrendData(noCacheForward, userMap, isLike, isFocus, countMap)

	// 生成最终的结果切片
	var res m.TrendList
	tempMap := make(map[int64]m.TrendShowInfo)
	for i, data := range findData {
		tempMap[data.TrendId] = data
		// 在feedMap 出现的数据是需要返回的主体动态 而且没有标记删除的
		_, ok := feedMap[strconv.FormatInt(data.TrendId, 10)]
		if !ok {
			continue
		}
		empty := m.TrendShowInfo{}
		empty.TrendId = data.TrendId
		// 如果已经删除 或权限私密 清空数据
		if data.IsDelete {
			empty.IsDelete = true
			findData[i] = empty
			res = append(res, empty)
			continue
		}
		if data.WhoSee == "privacy" && data.UserId != loginInfo.Id {
			empty.WhoSee = "privacy"
			res = append(res, empty)
			continue
		}
		// 处理转发
		if data.Forward != nil {
			if data.Forward.IsDelete {
				data.Forward = &m.TrendInfo{
					TrendId:  data.Forward.TrendId,
					IsDelete: true,
				}
			}
			if data.Forward.WhoSee == "privacy" {
				data.Forward = &m.TrendInfo{
					TrendId: data.Forward.TrendId,
					WhoSee:  "privacy",
				}
			}
		}

		// 是否是本人的动态
		data.IsOwner = loginInfo.Id == data.UserId

		res = append(res, data)
	}
	// 拼接缓存的转发, findData 是从数据库中查找的 已经有转发的数据里 所以不用拼接
	for i := 0; i < len(cacheTrends); i++ {
		// 是否是本人的动态
		cacheTrends[i].IsOwner = loginInfo.Id == cacheTrends[i].UserId

		trend := cacheTrends[i]
		// 如果已经删除 或权限私密 清空数据
		empty := m.TrendShowInfo{}
		empty.TrendId = trend.TrendId
		if trend.WhoSee == "privacy" && !trend.IsOwner {
			empty.WhoSee = "privacy"
			cacheTrends[i] = empty
			continue
		}
		if trend.ForwardInfo.Id == 0 {
			continue
		}
		// 处理转发
		var temp m.TrendInfo
		f := tempMap[trend.ForwardInfo.Id]
		temp.TrendId = f.TrendId
		if f.IsDelete == true {
			temp.IsDelete = true
			cacheTrends[i].Forward = &temp
			continue
		}
		if f.WhoSee == "privacy" {
			temp.WhoSee = "privacy"
			cacheTrends[i].Forward = &temp
			continue
		}
		// 通过json 拷贝对应的数据
		str, _ := json.Marshal(f)
		_ = json.Unmarshal(str, &temp)
		cacheTrends[i].Forward = &temp
	}

	//合成结果
	res = append(res, cacheTrends...)

	// 去重必须！
	res, _ = tools.RemoveSliceDuplicate(res)
	sort.Sort(res)
	ResponseSuccess(ctx, res)

	// 设置缓存
	needCache := append(noCacheForward, findData...)
	ctx.Set("trendData", needCache)

	// 设置作品流量量
	var viewIds []string
	for _, feed := range feedData {
		if feed.Type == "aw" {
			viewIds = append(viewIds, strconv.FormatInt(feed.MsgID, 10))
		}
	}
	ctx.Set("viewIds", viewIds)
}

// SetFeed 发布作品或动态后 设置到用户feed
func SetFeed(ctx *gin.Context) {
	ctxData, _ := ctx.Get("feed")
	feed := ctxData.(m.UploadArtOrTrend)

	lastUserID := "0"
	limit := 500

	for {
		fans, err := mysql.GetUserFans(feed.SendId, lastUserID, limit)
		if err != nil {
			logger.ErrZapLog(err, fmt.Sprintf("SetFeed GetUserFans fail %+v", feed))
			return
		}
		if len(fans) == 0 {
			break
		}

		err = mongo.SetFansFeed(fans, feed.MsgID, feed.SendId, feed.Type)
		if err != nil {
			logger.ErrZapLog(err, fmt.Sprintf("SetFeed SetFansFeed fail %+v", feed))
			return
		}
		lastUserID = fans[len(fans)-1]
	}
}
