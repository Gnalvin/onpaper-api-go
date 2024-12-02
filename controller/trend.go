package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	c "onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
	"strings"
)

// GetNewTrend 获取新动态
func GetNewTrend(ctx *gin.Context) {
	ctxData, _ := ctx.Get("nextId")
	nextId := ctxData.(*int64)

	feedData, err := mongo.GetNewTrend(*nextId)
	if err != nil {
		err = errors.Wrap(err, "GetNewTrend mongo fail")
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

// GetTrendDetail 获取动态详情
func GetTrendDetail(ctx *gin.Context) {
	// 获取 ctx 数据
	dataCtx, _ := ctx.Get("userInfo")
	loginUserInfo, _ := dataCtx.(m.UserTokenPayload)

	dataCtx, _ = ctx.Get("query")
	query, _ := dataCtx.(m.TrendQuery)

	var err error
	var trendData m.TrendList
	// 查找动态主体缓存
	trendData, _, err = c.GetBatchTrendCache([]string{strconv.FormatInt(query.TrendId, 10)})
	if err != nil {
		logger.ErrZapLog(err, "GetTrendDetail GetBatchTrendCache fail")
	}
	// 是否有缓存
	isHaveCache := len(trendData) != 0

	//没有缓存到数据库中查找
	if !isHaveCache {
		trendData, _, err = GetTrendInfo([]int64{query.TrendId}, query.Type)
		if err != nil {
			err = errors.Wrap(err, "GetTrendDetail GetTrendDetail fail")
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
	}

	// 没有查询到返回空数组
	if len(trendData) == 0 {
		ResponseError(ctx, CodeTrendNoExists)
		ctx.Abort()
		return
	}
	// 动态已经删除
	if trendData[0].IsDelete {
		ResponseError(ctx, CodeTrendNoExists)
		ctx.Abort()
		return
	}

	trendData[0].IsOwner = trendData[0].UserId == loginUserInfo.Id
	// 不是自己的动态
	if trendData[0].WhoSee == "privacy" && !trendData[0].IsOwner {
		ResponseError(ctx, CodeTrendNoExists)
		ctx.Abort()
		return
	}

	// 查询转发
	if trendData[0].ForwardInfo.Id != 0 {
		var mErr error
		// 缓存查询
		forwardData, _, mErr := c.GetBatchTrendCache([]string{strconv.FormatInt(trendData[0].ForwardInfo.Id, 10)})
		if err != nil {
			logger.ErrZapLog(err, "GetTrendDetail forwardCache fail")
		}
		// 数据库查询
		if len(forwardData) == 0 {
			forwardData, _, mErr = GetTrendInfo([]int64{trendData[0].ForwardInfo.Id}, trendData[0].ForwardInfo.Type)
			if mErr != nil {
				mErr = errors.Wrap(mErr, "GetTrendDetail Forward fail")
				ResponseErrorAndLog(ctx, CodeServerBusy, mErr)
				return
			}
		}
		if len(forwardData) != 0 {
			trendData = append(trendData, forwardData[0])
		}
	}

	var trendIds []string
	var findCount []m.MongoFeed
	var uIds []string

	for _, data := range trendData {
		findCount = append(findCount, m.MongoFeed{MsgID: data.TrendId, Type: data.Type})
		trendIds = append(trendIds, strconv.FormatInt(data.TrendId, 10))
		uIds = append(uIds, data.UserId)
	}

	userMap := make(map[string]m.UserSimpleInfo)
	// 查找用户数据
	userMap, err = mysql.GetBatchUserSimpleInfo(uIds)
	if err != nil {
		err = errors.Wrap(err, "GetTrendDetail UserSimpleInfo fail")
	}

	// 统计数据
	countMap, err := BatchGetTrendCount(findCount)
	if err != nil {
		err = errors.Wrap(err, "GetTrendDetail GetTrendCount fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 查找是否点赞过 关注过
	isLike := map[string]bool{}
	isFocus := map[string]uint8{}
	if loginUserInfo.Id != "" {
		isLike, err = c.CheckUserLike(trendIds, loginUserInfo.Id)
		if err != nil {
			logger.ErrZapLog(err, "GetTrendDetail CheckUserLike fail ")
		}

		isFocus, err = c.CheckUserFollow([]string{trendData[0].UserId}, loginUserInfo.Id)
		if err != nil {
			logger.ErrZapLog(err, "GetTrendDetail CheckUserFollow fail ")
		}
	}
	FormatTrendData(trendData, userMap, isLike, isFocus, countMap)

	// 合成转发数据
	if trendData[0].ForwardInfo.Id != 0 {
		// 处理转发
		var temp m.TrendInfo
		temp.TrendId = trendData[1].TrendId
		if trendData[1].IsDelete == true {
			temp.IsDelete = true
		} else if trendData[1].WhoSee == "privacy" {
			temp.WhoSee = "privacy"
		} else {
			// 通过json 拷贝对应的数据
			str, _ := json.Marshal(trendData[1])
			_ = json.Unmarshal(str, &temp)
		}
		trendData[0].Forward = &temp
	}

	ResponseSuccess(ctx, trendData[0])
	ctx.Set("trendData", trendData[0])
	ctx.Set("isHaveCache", isHaveCache)
}

// GetHotTrend 获取热门动态
func GetHotTrend(ctx *gin.Context) {

	trendIds, err := c.GetHotTrendId()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	var feedData []m.MongoFeed
	for _, data := range trendIds {
		strList := strings.Split(data, "&")
		msgId, _err := strconv.ParseInt(strList[0], 10, 64)
		if _err != nil {
			//数据库错误
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}
		feedData = append(feedData, m.MongoFeed{
			AcceptId: "",
			MsgID:    msgId,
			SendId:   strList[2],
			Type:     strList[1],
		})
	}
	// 如果没有 直接返回
	dataLen := len(feedData)
	if dataLen == 0 {
		// 返回数据
		ResponseSuccess(ctx, make([]struct{}, 0))
		ctx.Abort()
		return
	}
	ctx.Set("feedData", feedData)
}

// GetUserTrend 获取某个用户的动态
func GetUserTrend(ctx *gin.Context) {
	ctxData, _ := ctx.Get("query")
	query := ctxData.(m.TrendUserQuery)

	feedData, err := mongo.GetOneUserTrend(query.UserId, *query.NextId)
	if err != nil {
		err = errors.Wrap(err, "GetUserTrend mongo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 如果没有feed 直接返回
	dataLen := len(feedData)
	if dataLen == 0 {
		// 返回数据
		ResponseError(ctx, CodeUserNoHaveTrends)
		ctx.Abort()
		return
	}
	ctx.Set("feedData", feedData)
}

// DeleteTrend 删除一条动态
func DeleteTrend(ctx *gin.Context) {
	ctxData, _ := ctx.Get("query")
	query, _ := ctxData.(m.TrendQuery)

	ctxData, _ = ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	// 删除自己的feed
	err := mongo.DeleteOneFeed(query.TrendId, userInfo.Id, userInfo.Id)
	if err != nil {
		err = errors.Wrap(err, "DeleteTrend DeleteOneFeed fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 动态标记删除
	err = mongo.DeleteTrend(query.TrendId)
	if err != nil {
		err = errors.Wrap(err, "DeleteTrend DeleteTrend fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//删除 动态缓存
	key := fmt.Sprintf(c.TrendProfile, strconv.FormatInt(query.TrendId, 10))
	err = c.DelOneCache(key)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 到热门中删除
	err = c.DeleteOneHotTrend(strconv.FormatInt(query.TrendId, 10) + "&tr&" + userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 作者动态数 - 1
	err = c.SetUserCountField(userInfo.Id, "TrendCount", -1)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, query)
}

// UpdateTrendPermission 更新动态权限
func UpdateTrendPermission(ctx *gin.Context) {
	ctxData, _ := ctx.Get("permission")
	permission, _ := ctxData.(m.TrendPermission)

	ctxData, _ = ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	err := mongo.UpdateTrendPermission(permission)
	if err != nil {
		err = errors.Wrap(err, "UpdateTrendPermission mongo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//删除 动态缓存
	key := fmt.Sprintf(c.TrendProfile, strconv.FormatInt(permission.TrendId, 10))
	err = c.DelOneCache(key)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 如果权限下设置成 不是公开
	if permission.WhoSee != "public" {
		// 到热门中删除
		err = c.DeleteOneHotTrend(strconv.FormatInt(permission.TrendId, 10) + "&tr&" + userInfo.Id)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
	}

	ResponseSuccess(ctx, permission)
}
