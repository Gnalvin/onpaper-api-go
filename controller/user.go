package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	c "onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/formatTools"
	"onpaper-api-go/utils/singleFlight"
	"strconv"
	"time"
)

// GetProfileData  返回 获取的主页 profile
func GetProfileData(ctx *gin.Context) {
	// 1. 获取传递的参数
	dataCtx, _ := ctx.Get("userId")
	urlUserId := dataCtx.(string)

	dataCtx, _ = ctx.Get("userInfo")
	loginUserInfo := dataCtx.(m.UserTokenPayload)

	dataCtx, _ = ctx.Get("profile")
	profileCtx := dataCtx.(c.CtxCacheVale)

	dataCtx, _ = ctx.Get("isFocus")
	isFocusCtx := dataCtx.(c.CtxCacheVale)

	dataCtx, _ = ctx.Get("userCount")
	userCountCtx := dataCtx.(c.CtxCacheVale)

	userProfile := m.UserHomeProfile{}
	userProfile.IsFocus = isFocusCtx.Val.(uint8)
	userProfile.Owner = loginUserInfo.Id == urlUserId

	// 2. 查看缓存是否存在 存在读取 否则数据库读取
	if profileCtx.HaveCache {
		userProfile.Profile = profileCtx.Val.(m.UserProfileTableInfo)
		if userCountCtx.HaveCache {
			userProfile.Profile.Count = userCountCtx.Val.(m.UserAllCount)
		}
		// 如果不是自己登陆 不返回登陆邮箱
		if !userProfile.Owner {
			userProfile.Profile.Email = nil
		}
		ResponseSuccess(ctx, gin.H{"profile": userProfile.Profile, "owner": userProfile.Owner, "isFocus": userProfile.IsFocus})
		ctx.Abort()
		return
	}

	ctxTimeOut, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 使用完整路径为 key
	key := (ctx.Request.URL).String()
	profileRes, err := singleFlight.Do(ctxTimeOut, key, func() (res interface{}, err error) {
		return mysql.GetUserProfileById(urlUserId)
	})
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			ResponseError(ctx, CodeUserDoseNotExists)
			return
		}
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	userProfile.Profile = profileRes.(m.UserProfileTableInfo)
	if userCountCtx.HaveCache {
		userProfile.Profile.Count = userCountCtx.Val.(m.UserAllCount)
	}
	if !userProfile.Owner {
		userProfile.Profile.Email = nil
	}
	ResponseSuccess(ctx, gin.H{"profile": userProfile.Profile, "owner": userProfile.Owner, "isFocus": userProfile.IsFocus})

	ctx.Set("profileRes", profileRes)
}

// GetUserHomeCollect 返回用户主页收藏的作品
func GetUserHomeCollect(ctx *gin.Context) {
	dataCtx, _ := ctx.Get("query")
	query := dataCtx.(m.VerifyUserAndPage)

	// 查询用户收藏数据
	artAndUid, err := mongo.GetArtworkCollect(query.UId, query.Page-1)
	if err != nil {
		err = errors.Wrap(err, "GetUserHomeCollect: mongodb get fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	if len(artAndUid) == 0 && query.Page == 1 {
		ResponseError(ctx, CodeUserNoHaveCollects)
		return
	}

	var artIds []string
	var uIds []string
	for _, data := range artAndUid {
		artIds = append(artIds, data.MsgId)
		uIds = append(uIds, data.AuthorId)
	}
	artData, findData, err := BatchGetBasicArtInfo(artIds, uIds)
	if err != nil {
		err = errors.Wrap(err, "GetUserHomeArtwork BatchGetBasicArtInfo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 返回数据
	ResponseSuccess(ctx, artData)

	ctx.Set("artData", findData)
}

// GetUserHomeArtwork 返回用户主页作品
func GetUserHomeArtwork(ctx *gin.Context) {
	// 1. 获取缓存的数据
	dataCtx, _ := ctx.Get("userInfo")
	loginUser := dataCtx.(m.UserTokenPayload)

	dataCtx, _ = ctx.Get("query")
	query := dataCtx.(m.VerifyUserHomeArtwork)
	// 是否自己查看作品
	isOwner := loginUser.Id == query.UId

	//data.Page -1 第一页 从第0条开始
	artCounts, err := mysql.GetUserHomeArtwork(query.UId, query.Page-1, query.Sort)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 查询获得的作品数
	artLen := len(artCounts)
	// 如果没有作品 直接返回
	if artLen == 0 {
		ResponseError(ctx, CodeUserNoHaveArtworks)
		return
	}

	var artIds []string
	var uIds []string
	for _, d := range artCounts {
		artIds = append(artIds, d.ArtworkId)
		uIds = append(uIds, d.UserId)
	}

	var eg errgroup.Group

	var artData []m.BasicArtwork
	var needCacheArt []m.BasicArtwork
	eg.Go(func() (mErr error) {
		artData, needCacheArt, mErr = BatchGetBasicArtInfo(artIds, uIds)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetUserHomeArtwork BatchGetBasicArtInfo fail")
		}
		return
	})

	var cacheCountMap map[string]map[string]string
	var needCacheCount []m.ArtworkCount
	eg.Go(func() (mErr error) {
		cacheCountMap, needCacheCount, mErr = BatchGetArtworkCount(artIds)
		if mErr != nil {
			logger.ErrZapLog(mErr, "GetUserHomeArtwork BatchGetArtworkCount fail")
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
			logger.ErrZapLog(mErr, "GetUserHomeArtwork CheckUserLike fail ")
		}
		// 出错就算了
		return nil
	})

	err = eg.Wait()
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
	}

	// 设置作品流量量
	var viewIds []string
	var resData []m.BasicArtworkAndLike
	// 拼接数据
	for _, artwork := range artData {
		//如果不是自己主页查看 私密作品不返回
		if artwork.WhoSee == "privacy" && !isOwner {
			continue
		}
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
		temp.IsOwner = isOwner
		resData = append(resData, temp)
		viewIds = append(viewIds, temp.ArtworkId)
	}

	// 返回数据
	ResponseSuccess(ctx, gin.H{
		"artworks": resData,
	})

	err = c.SetArtworkCount(needCacheCount)
	if err != nil {
		logger.ErrZapLog(err, "SetArtworkCount fail")
	}

	ctx.Set("viewIds", viewIds)
	ctx.Set("artData", needCacheArt)
}

// UpdateUserProfile 更新用户资料
func UpdateUserProfile(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, exi := ctx.Get("userInfo")
	ctxProfile, exi := ctx.Get("profile")
	// 如果取不到数据
	if !exi {
		err := errors.New("UpdateUserProfile get ctx fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 类型断言 得到 payload 里面的用户id 和名字
	userInfo, ok := ctxData.(m.UserTokenPayload)
	profile, ok := ctxProfile.(m.UpdateProfileData)
	// 如果断言错误
	if !ok {
		err := errors.New("UpdateUserProfile assert fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	var err error
	switch profile.ProfileType {
	case "userName":
		//到数据库中更新数据 如错误 返回
		err = mysql.UpdateUserName(profile.Profile, userInfo.Id)
	case "sex":
		err = mysql.UpdateUserSex(profile.Profile, userInfo.Id)
	case "birthday":
		err = mysql.UpdateUserBirthday(profile.Profile, userInfo.Id)
	case "workEmail":
		err = mysql.UpdateUserWorkEmail(profile.Profile, userInfo.Id)
	case "region":
		err = mysql.UpdateUserAddress(profile.Profile, userInfo.Id)
	case "createStyle":
		err = mysql.UpdateUserCreateStyle(profile.Profile, userInfo.Id)
	case "software":
		err = mysql.UpdateUserSoftware(profile.Profile, userInfo.Id)
	case "exceptWork":
		err = mysql.UpdateUserExpectWork(profile.Profile, userInfo.Id)
	case "introduce":
		err = mysql.UpdateUserIntroduce(profile.Profile, userInfo.Id)
	case "snsLink":
		err = mysql.UpdateUserSns(profile.SnsData, userInfo.Id)
	}
	if err != nil {
		ResponseError(ctx, CodeServerBusy)
		return
	}

	// 返回数据
	ResponseSuccess(ctx, profile)
}

// GetNavData 获取导航栏信息
func GetNavData(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userId")
	urlUserId := ctxData.(string)

	navUserData, err := mysql.GetUserNavDataById(urlUserId)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			ResponseError(ctx, CodeUserDoseNotExists)
			return
		}
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	msgUnread, err := mongo.GetUserUnreadCount(urlUserId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	navUserData.MessageUnread = msgUnread.TotalUnread
	// 获取缓存中的统计
	countCtx, err := c.GetUserCount(urlUserId)
	if err != nil {
		logger.ErrZapLog(err, fmt.Sprintf("GetNavData GetUserCount fail %s", urlUserId))
	} else {
		count := countCtx.Val.(m.UserAllCount)
		navUserData.Fans = count.Fans
		navUserData.Following = count.Following
	}
	// 返回数据
	ResponseSuccess(ctx, navUserData)
}

// SaveUserFocus 保存关注的用户
func SaveUserFocus(ctx *gin.Context) {
	// 获取传递的参数
	FocusData, _ := ctx.Get("Focus")
	userInfo, _ := ctx.Get("userInfo")

	focusInfo := FocusData.(m.VerifyUserFocus)
	loginUser := userInfo.(m.UserTokenPayload)

	// 不允许关注自己
	if focusInfo.FocusId == loginUser.Id {
		ResponseError(ctx, CodeParamsError)
		return
	}

	// 设置登陆用户的count 缓存，之后要对关注数+1
	err := c.CheckUserCount(loginUser.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 数据库保存关注数据
	isChange, err := mysql.SaveUserFocus(focusInfo, loginUser.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	//没有实际更新 不继续缓存操作
	if isChange == false {
		ctx.Abort()
	}

	// 返回数据
	ResponseSuccess(ctx, FocusData)
}

// GetRecommendUser 获取推荐热门用户的数据
func GetRecommendUser(ctx *gin.Context) {
	ctxData, exi := ctx.Get("userInfo")
	if !exi {
		err := errors.New("GetChannelArtwork get ctxData fail")
		zap.L().Error("GetChannelArtwork get ctxData fail", zap.Error(err))
		return
	}

	loginData, _ := ctxData.(m.UserTokenPayload)

	hotUserId, err := c.GetHotUserId(loginData.Id)
	if err != nil {
		err = errors.Wrap(err, "GetRecommendUser GetHotUserId fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 随机取出30个
	hotUserId = formatTools.RandGetSlice(hotUserId, 12)
	// 到缓存中取数据
	fmtKey := fmt.Sprintf(c.HotUser, "%s")
	var temp m.HotUser
	resData, _, err := c.BatchGetTypeOfString(fmtKey, hotUserId, temp)

	// 返回数据
	ResponseSuccess(ctx, resData)
}

// GetUserRank 获取用户排名
func GetUserRank(ctx *gin.Context) {
	var userRank []m.UserBigCard

	dataCtx, _ := ctx.Get("rankData")
	rankData, _ := dataCtx.(c.CtxCacheVale)

	dataCtx, _ = ctx.Get("rankType")
	queryData, _ := dataCtx.(m.QueryUserRank)

	dataCtx, _ = ctx.Get("userInfo")
	tokenInfo, _ := dataCtx.(m.UserTokenPayload)

	// 获取缓存数据
	if rankData.HaveCache {
		userRank = rankData.Val.([]m.UserBigCard)
		ctx.Abort()
	} else {
		ctxTimeOut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// 使用完整路径为 key
		key := (ctx.Request.URL).String()
		res, err := singleFlight.Do(ctxTimeOut, key, func() (interface{}, error) {
			userIntro, err := mysql.GetUserRankData(queryData.RankType)
			if err != nil {
				return nil, err
			}
			return userIntro, err
		})

		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		userRank = res.([]m.UserBigCard)
		ctx.Set("userRank", userRank)
		ctx.Next()
	}

	// 如果是登录用户 到缓存查找关注列表 查看是否在排名里有关注过的
	if tokenInfo.Id != "" {
		var checkId []string
		for _, rank := range userRank {
			checkId = append(checkId, rank.UserId)
		}
		focusIdMap, err := c.CheckUserFollow(checkId, tokenInfo.Id)
		if err != nil {
			logger.ErrZapLog(err, "CheckUserFollow fail")
		}

		// 查询是否存在关注的
		for i, ui := range userRank {
			isFocus, ok := focusIdMap[ui.UserId]
			if ok {
				userRank[i].IsFocus = isFocus
			}
		}
	}
	// 返回数据
	ResponseSuccess(ctx, userRank)
}

// GetUserFollowList 获取用户关注列表
func GetUserFollowList(ctx *gin.Context) {
	// 1. 获取传递的参数
	dataCtx, _ := ctx.Get("query")
	query := dataCtx.(m.VerifyUserFollow)

	dataCtx, _ = ctx.Get("userInfo")
	loginInfo, _ := dataCtx.(m.UserTokenPayload)

	var userIds []string
	var err error
	if query.Type == "follower" {
		userIds, err = mysql.GetUserFansList(query.UId, query.Page-1)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		if len(userIds) == 0 && query.Page == 1 {
			ResponseError(ctx, CodeUserNoHaveFans)
			return
		}
	} else {
		userIds, err = mysql.GetUserFocusIdList(query.UId, query.Page-1)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		if len(userIds) == 0 && query.Page == 1 {
			ResponseError(ctx, CodeUserNoHaveFocus)
			return
		}
	}

	cacheData, needFind, err := c.GetUserSmallCarCache(userIds)
	if err != nil {
		logger.ErrZapLog(err, "GetUserFollowList GetUserSmallCarCache fail")
	}

	var queryUser []string
	for _, data := range needFind {
		queryUser = append(queryUser, data.Id)
	}

	mysqlData, err := mysql.BatchGetUserBaseInfo(queryUser)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 创建一个 map 来保存所有的数据
	dataMap := make(map[string]m.UserSmallCard)

	// 将 mysqlData 和 cacheData 的数据添加到 map 中
	for _, d := range mysqlData {
		dataMap[d.UserId] = d
	}
	for _, d := range cacheData {
		dataMap[d.UserId] = d
	}

	result := make([]m.UserSmallCard, len(userIds))
	// 按照 focusIds 的顺序从 map 中获取数据并添加到结果中
	for i, id := range userIds {
		if data, ok := dataMap[id]; ok {
			result[i] = data
		}
	}

	// 如果是登录用户 到缓存查找关注列表 查看是否在排名里有关注过的
	if loginInfo.Id != "" {
		focusIdMap, mErr := c.CheckUserFollow(userIds, loginInfo.Id)
		if mErr != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, mErr)
			return
		}

		// 查询是否存在关注的
		for i, ui := range result {
			result[i].IsFocus = focusIdMap[ui.UserId]
		}
	}

	ResponseSuccess(ctx, result)

	ctx.Set("userData", mysqlData)
}

// SearchUserByName 通过名字查找用户
func SearchUserByName(ctx *gin.Context) {
	name, ok := ctx.GetQuery("name")
	if !ok || name == "" {
		ResponseError(ctx, CodeParamsError)
		return
	}

	searchData, likeData, err := mysql.SearchUserByName(name)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"likeData":   likeData,
		"searchData": searchData,
	})
}

// SearchOurFocusUser 精确查找自己关注的用户
func SearchOurFocusUser(ctx *gin.Context) {
	name, ok := ctx.GetQuery("name")
	if !ok || name == "" {
		ResponseError(ctx, CodeParamsError)
		return
	}

	dataCtx, _ := ctx.Get("userInfo")
	tokenInfo, _ := dataCtx.(m.UserTokenPayload)

	res, err := mysql.SearchOurFocus(name, tokenInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, res)
}

// GetUserPanelInfo 获取用户面板信息
func GetUserPanelInfo(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userId")
	queryId := ctxData.(string)

	ctxData, _ = ctx.Get("userCount")
	countCtx := ctxData.(c.CtxCacheVale)

	ctxData, _ = ctx.Get("userProfile")
	infoCtx := ctxData.(c.CtxCacheVale)

	ctxData, _ = ctx.Get("userInfo")
	loginUserInfo := ctxData.(m.UserTokenPayload)

	// 如果都有缓存直接返回
	if countCtx.HaveCache && infoCtx.HaveCache {
		resData := infoCtx.Val.(m.UserPanel)
		count := countCtx.Val.(m.UserAllCount)
		resData.Collects = count.Collects
		resData.Fans = count.Fans
		resData.Likes = count.Likes
		resData.IsOwner = resData.UserId == loginUserInfo.Id
		ResponseSuccess(ctx, resData)
		ctx.Abort()
		return
	}

	userPanel, err := mysql.GetUserPanel(queryId)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			ResponseError(ctx, CodeUserDoseNotExists)
			return
		}
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	userPanel.IsOwner = userPanel.UserId == loginUserInfo.Id
	ResponseSuccess(ctx, userPanel)
	ctx.Set("userPanel", userPanel)
}

// GetUserInvitationCode 查找用户邀请码
func GetUserInvitationCode(ctx *gin.Context) {
	dataCtx, _ := ctx.Get("userInfo")
	loginUserInfo := dataCtx.(m.UserTokenPayload)

	codeData, err := mysql.GetUserInvitationCode(loginUserInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, codeData)
}

// GetUserShow 获取展示用户
func GetUserShow(ctx *gin.Context) {
	ctxData, _ := ctx.Get("query")
	query := ctxData.(m.AllUserShowQuery)

	ctxData, _ = ctx.Get("needCacheData")
	tokenInfo, _ := ctxData.(m.UserTokenPayload)

	bigCardUser, err := mysql.GetAllUserShowId(query)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	var queryUser []string
	tempMap := make(map[string]m.BigCardUserId, 0)
	for _, u := range bigCardUser {
		queryUser = append(queryUser, u.UserId)
		tempMap[u.UserId] = u
	}

	cacheData, needFind, err := c.GetUserBigCarCache(queryUser)
	if err != nil {
		logger.ErrZapLog(err, "GetAllFeed GetBatchTrendCache fail")
	}

	queryUser = []string{}
	for _, data := range needFind {
		queryUser = append(queryUser, data.Id)
	}

	needCacheData, err := mysql.BatchGetUserAllInfo(queryUser, 5)
	if err != nil {
		err = errors.Wrap(err, "GetUserShow BatchGetUserAllInfo fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	userData := m.UserBigCardList(append(cacheData, needCacheData...))

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

	for i, d := range userData {
		data := tempMap[d.UserId]
		userData[i].Score = data.Score
		userData[i].Active = data.Active
	}

	if query.Type == "hot" {
		userData.SortByScoreDesc()
	} else if query.Type == "new" {
		userData.SortByUserIdDesc()
	} else {
		userData.SortByActiveDesc()
	}

	ResponseSuccess(ctx, gin.H{
		"userData": userData,
		"isEnd":    len(userData) < 20,
	})

	ctx.Set("userData", needCacheData)
}
