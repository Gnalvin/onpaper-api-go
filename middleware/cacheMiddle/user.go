package cacheMiddle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	c "onpaper-api-go/cache"
	ctl "onpaper-api-go/controller"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
	"time"
)

// InitUserData 登录成功后把 如关注列表 收藏列表 这些数据保存到cache
func InitUserData(ctx *gin.Context) {
	//获取到 用户信息
	dataCtx, _ := ctx.Get("userInfo")
	userData := dataCtx.(m.UserTableInfo)
	isAllOk := true

	var eg errgroup.Group

	// 1. 设置 关注缓存
	eg.Go(func() error {
		err := initFocusIdList(userData.SnowId)
		return err
	})

	// 2. 设置收藏作品缓存
	eg.Go(func() error {
		err := initCollectList(userData.SnowId)
		return err
	})

	//3.设置点赞作品缓存
	eg.Go(func() error {
		err := initLikeArtId(userData.SnowId)
		return err
	})

	//4。初始化用户统计
	eg.Go(func() error {
		err := c.CheckUserCount(userData.SnowId)
		return err
	})

	err := eg.Wait()
	if err != nil {
		isAllOk = false
		logger.ErrZapLog(err, userData.SnowId)
	}

	ctx.Set("isOk", isAllOk)
}

func initFocusIdList(userId string) (err error) {
	focusList, err := mysql.GetUserFocusUserId(userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserFocusUserId mongo")
		return
	}
	// 如果数组为空的 0 占位
	if len(focusList) == 0 {
		focusList = []string{"0"}
	}
	//转换数组格式
	listTemp := make([]interface{}, len(focusList))
	for i := range focusList {
		listTemp[i] = focusList[i]
	}

	err = c.SetUserFocusUserId(userId, listTemp)
	return
}

func initCollectList(userId string) (err error) {
	collectData, err := mongo.GetUserALlCollect(userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserALlCollect mongo")
		return
	}

	if len(collectData) == 0 {
		collectData = append(collectData, m.InitUserData{
			MsgId: "0",
			Time:  time.Now(),
		})
	}

	err = c.SetUerCollectId(userId, collectData)

	return
}

func initLikeArtId(userId string) (err error) {

	likeData, err := mongo.GetUserAllLike(userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserAllLike mongo")
		return
	}
	if len(likeData) == 0 {
		likeData = append(likeData, m.InitUserData{
			MsgId: "0",
			Time:  time.Now(),
		})
	}

	err = c.SetUserLikeArtId(userId, likeData)

	return
}

// GetUserHomeProfile 获取用户首页资料
func GetUserHomeProfile(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userId")
	urlUserId := ctxData.(string)
	ctxData, _ = ctx.Get("userInfo")
	loginUserInfo := ctxData.(m.UserTokenPayload)

	err := c.CheckUserCount(urlUserId)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			ctl.ResponseError(ctx, ctl.CodeUserDoseNotExists)
			return
		}
		logger.ErrZapLog(err, "GetUserHomeProfile CheckUserCount fail")
	}

	ctxTimeOut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 1. 管道查询 profile
	pipe := c.Rdb.Pipeline()
	pKey := fmt.Sprintf(c.UserProfile, urlUserId)
	pipe.Get(ctxTimeOut, pKey)

	// 查询用户点赞收藏数
	cKey := fmt.Sprintf(c.UserCount, urlUserId)
	pipe.HGetAll(ctx, cKey)

	cmders, err := pipe.Exec(ctxTimeOut)
	if err != nil {
		// 如果返回的错误 不是 key不存在
		if errors.Is(err, redis.Nil) == false {
			logger.ErrZapLog(err, "GetUserHomeProfile cache fail")
		}
	}

	// 2. 转换str 数据 并传递
	var temp m.UserProfileTableInfo
	var profile c.CtxCacheVale

	pVal := c.GetCmderStringResult(cmders[0])
	if pVal != "" {
		err = json.Unmarshal([]byte(pVal), &temp)
		if err != nil {
			err = errors.Wrap(err, "GetUserHomeProfile Unmarshal fail ")
			logger.ErrZapLog(err, pVal)
		} else {
			profile.Val = temp
			profile.HaveCache = true
		}
	}

	var isFocus c.CtxCacheVale
	//如果是游客 或者是自己查看自己 设置成有缓存 避免到数据库中查询
	if loginUserInfo.Id == "" || loginUserInfo.Id == urlUserId {
		isFocus.Val = uint8(0)
		isFocus.HaveCache = true
	} else {
		focusMap, _err := c.CheckUserFollow([]string{urlUserId}, loginUserInfo.Id)
		if _err != nil {
			_err = errors.Wrap(_err, "GetUserHomeProfile CheckUserFollow fail ")
			logger.ErrZapLog(_err, urlUserId)
		} else {
			isFocus.Val = focusMap[urlUserId]
			isFocus.HaveCache = true
		}
	}

	var userCount c.CtxCacheVale
	res := cmders[1].(*redis.MapStringStringCmd).Val()
	if len(res) != 0 {
		userCount.HaveCache = true
		fans, _ := strconv.Atoi(res["Fans"])
		likes, _ := strconv.Atoi(res["Likes"])
		collect, _ := strconv.Atoi(res["Collects"])
		artCount, _ := strconv.Atoi(res["ArtCount"])
		collectCount, _ := strconv.Atoi(res["CollectCount"])
		follows, _ := strconv.Atoi(res["Follows"])
		uCount := m.UserAllCount{
			Fans:         fans,
			Likes:        likes,
			Collects:     collect,
			ArtCount:     artCount,
			CollectCount: collectCount,
			Following:    follows,
		}
		userCount.Val = uCount
	}

	ctx.Set("profile", profile)
	ctx.Set("isFocus", isFocus)
	ctx.Set("userCount", userCount)
}

// SetUserHomeProfile  设置用户首页缓存
func SetUserHomeProfile(ctx *gin.Context) {
	// 1. 获取传递的数据
	dataCtx, _ := ctx.Get("profile")
	profileCtx := dataCtx.(c.CtxCacheVale)

	profileData, pExist := ctx.Get("profileRes")

	// 2. 如果没有缓存则设置缓存
	if profileCtx.HaveCache == false && pExist {
		profile := profileData.(m.UserProfileTableInfo)
		err := c.SetUserHomeProfile(profile)
		if err != nil {
			logger.ErrZapLog(err, profile.UserId)
		}
	}

}

// DelUserProfile 删除用户资料缓存
func DelUserProfile(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)
	key := fmt.Sprintf(c.UserProfile, userInfo.Id)
	err := c.DelOneCache(key)
	if err != nil {
		logger.ErrZapLog(err, userInfo.Id)
	}
}

// SetUserAboutArtCache 设置用户作品相关缓存
func SetUserAboutArtCache(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	// 作品相关信息
	infoData, _ := ctx.Get("artworkInfo")
	artworkInfo := infoData.(*m.SaveArtworkInfo)

	err := c.SetUploadArtAbout(userInfo.Id, artworkInfo.Tags)
	if err != nil {
		logger.ErrZapLog(err, userInfo.Id)
	}
}

// SetUserRank 设置用户排行缓存
func SetUserRank(ctx *gin.Context) {
	dataCtx, _ := ctx.Get("rankType")
	queryData := dataCtx.(m.QueryUserRank)
	dataCtx, _ = ctx.Get("userRank")
	userData := dataCtx.([]m.UserBigCard)

	err := c.SetUserRank(queryData.RankType, userData)
	if err != nil {
		logger.ErrZapLog(err, queryData.RankType)
	}
}

func GetUserRank(ctx *gin.Context) {
	dataCtx, _ := ctx.Get("rankType")
	queryData := dataCtx.(m.QueryUserRank)

	key := fmt.Sprintf(c.RankUser, queryData.RankType)

	var temp []m.UserBigCard

	userRank, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, queryData.RankType)
	}

	ctx.Set("rankData", userRank)

}

// SetFocusCount 设置关注缓存
func SetFocusCount(ctx *gin.Context) {
	// 获取传递的参数
	dataCtx, _ := ctx.Get("Focus")
	focusInfo := dataCtx.(m.VerifyUserFocus)

	dataCtx, _ = ctx.Get("userInfo")
	loginInfo := dataCtx.(m.UserTokenPayload)

	err := c.SetFocusCount(loginInfo.Id, focusInfo)
	if err != nil {
		logger.ErrZapLog(err, focusInfo)
	}
}

// SetAboutTrendCache 设置动态相关缓存
func SetAboutTrendCache(ctx *gin.Context) {
	//1.获取传递的 作品信息
	infoData, _ := ctx.Get("trendInfo")
	trendInfo := infoData.(m.SaveTrendInfo)

	// 作者动态数 + 1
	err := c.SetUserCountField(trendInfo.UserId, "TrendCount", 1)
	if err != nil {
		logger.ErrZapLog(err, trendInfo.UserId)
	}

	if trendInfo.ForwardInfo.Id != 0 {
		err = c.SetTrendForwards(trendInfo.ForwardInfo)
		if err != nil {
			logger.ErrZapLog(err, trendInfo.TrendId)
		}
	}

	if trendInfo.Topic.Text != "" {
		err = c.SetHotTopicIncr(trendInfo.Topic.Text)
		if err != nil {
			logger.ErrZapLog(err, trendInfo.TrendId)
		}
	}
}

// GetUserPanel 获取用户面板资料
func GetUserPanel(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userId")
	queryId := ctxData.(string)

	// 反序列化
	var temp m.UserPanel
	key := fmt.Sprintf(c.UserPanel, queryId)
	userCtx, err := c.GetOneStringValue(key, temp)
	if err != nil {
		logger.ErrZapLog(err, queryId)
	}

	countCtx, err := c.GetUserCount(queryId)
	if err != nil {
		logger.ErrZapLog(err, queryId)
	}

	ctx.Set("userCount", countCtx)
	ctx.Set("userProfile", userCtx)
}

func SetUserPanel(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userPanel")
	userPanel := ctxData.(m.UserPanel)

	// 重置是否本人
	userPanel.IsOwner = false

	key := fmt.Sprintf(c.UserPanel, userPanel.UserId)
	err := c.SetOneStringValue(key, userPanel, time.Hour*24)
	if err != nil {
		logger.ErrZapLog(err, "SetUserPanel fail")
	}

	err = c.CheckUserCount(userPanel.UserId)
	if err != nil {
		logger.ErrZapLog(err, "SetUserPanel  CheckUserCount fail")
	}
}

// SetUserNotifyConfig 设置用户通知配置缓存
func SetUserNotifyConfig(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userId")
	userId := ctxData.(string)

	ctxData, _ = ctx.Get("config")
	config, _ := ctxData.(m.NotifyConfig)

	key := fmt.Sprintf(c.NotifyConfig, userId)
	mapConfig := map[string]interface{}{
		"Like":    config.Like,
		"Collect": config.Collect,
		"Comment": config.Comment,
		"Message": config.Message,
		"At":      config.At,
		"Follow":  config.Follow,
	}
	err := c.SetHashKeyValue(key, mapConfig, 3*24)
	if err != nil {
		logger.ErrZapLog(err, "SetUserNotifyConfig fail")
	}
}

// SetUserBigCarCache 设置用户大卡片资料缓存
func SetUserBigCarCache(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userData")
	userData := ctxData.([]m.UserBigCard)

	err := c.SetUserBigCarCache(userData)
	if err != nil {
		logger.ErrZapLog(err, "SetUserBigCarCache fail")
	}
}

// SetUserSmallCarCache 设置用户小卡片资料缓存
func SetUserSmallCarCache(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userData")
	userData := ctxData.([]m.UserSmallCard)

	err := c.SetUserSmallCarCache(userData)
	if err != nil {
		logger.ErrZapLog(err, "SetUserSmallCarCache fail")
	}
}
