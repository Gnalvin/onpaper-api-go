package controller

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	c "onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
)

// BatchGetBasicArtInfo 批量获取基础作品数据
func BatchGetBasicArtInfo(artIds, userIds []string) (artData, findData []m.BasicArtwork, err error) {
	if len(artIds) == 0 {
		return
	}
	// 先从缓存获取
	artData, needFind, err := c.GetBasicArtCache(artIds)
	if err != nil {
		logger.ErrZapLog(err, "BatchGetBasicArtInfo fail")
	}

	var findArtIds []string
	for _, find := range needFind {
		findArtIds = append(findArtIds, find.Id)
	}
	// 缓存没有获取到的部分 从数据库获取
	findData, err = mysql.GetBatchBasicShowArtInfo(findArtIds, userIds)
	if err != nil {
		err = errors.Wrap(err, "BatchGetBasicArtInfo fail")
		return
	}

	//按顺序插回
	for i, data := range needFind {
		temp := []m.BasicArtwork{findData[i]}
		artData = append(artData[:data.Index], append(temp, artData[data.Index:]...)...)
	}

	return
}

// GetTrendInfo 查询动态
func GetTrendInfo(queryId []int64, queryType string) (trendData, needForwardCache m.TrendList, err error) {
	if len(queryId) == 0 || queryType == "" {
		return
	}
	// 1. 查询主题动态
	if queryType == "aw" {
		trendData, err = mysql.BatchGetTrendArtInfo(queryId)
		if err != nil {
			err = errors.Wrap(err, "GetTrendInfo aw fail")
		}
		return
	}

	trendData, err = mongo.GetMoreTrendInfo(queryId)
	if err != nil {
		err = errors.Wrap(err, "GetTrendInfo tr fail")
		return
	}

	// 2. 如果存在转发的动态
	var forwardId []string
	for _, data := range trendData {
		if data.ForwardInfo.Id != 0 {
			forwardId = append(forwardId, strconv.FormatInt(data.ForwardInfo.Id, 10))
		}
	}
	if len(forwardId) == 0 {
		return
	}
	// 2.1 到缓存中查找
	cacheTrends, needFind, err := c.GetBatchTrendCache(forwardId)
	if err != nil {
		logger.ErrZapLog(err, "GetBatchTrendCache fail")
	}
	// 2.2在缓存中没有
	findMap := make(map[int64]struct{})
	for _, data := range needFind {
		id, _ := strconv.ParseInt(data.Id, 10, 64)
		findMap[id] = struct{}{}
	}

	var findArt []int64
	var findTrend []int64
	for _, data := range trendData {
		id := data.ForwardInfo.Id
		_, ok := findMap[id]
		if ok {
			if data.ForwardInfo.Type == "aw" {
				findArt = append(findArt, id)
			} else {
				findTrend = append(findTrend, id)
			}
		}
	}

	var forwardAw m.TrendList
	var forwardTr m.TrendList
	forwardAw, _, err = GetTrendInfo(findArt, "aw")
	if err != nil {
		err = errors.Wrap(err, "GetAllFeed GetTrendInfo fail")
		return
	}

	forwardTr, _, err = GetTrendInfo(findTrend, "tr")
	if err != nil {
		err = errors.Wrap(err, "GetAllFeed GetTrendInfo fail")
		return
	}

	// 3. 合成数据
	ForwardMap := make(map[int64]m.TrendShowInfo)
	needForwardCache = append(forwardAw, forwardTr...)
	for _, f := range needForwardCache {
		ForwardMap[f.TrendId] = f
	}

	for _, f := range cacheTrends {
		ForwardMap[f.TrendId] = f
	}

	for i, data := range trendData {
		if data.ForwardInfo.Id != 0 {
			f := ForwardMap[data.ForwardInfo.Id]
			var temp m.TrendInfo
			// 通过json 拷贝对应的数据
			str, _ := json.Marshal(f)
			_ = json.Unmarshal(str, &temp)
			trendData[i].Forward = &temp
		}
	}
	return
}

// FormatTrendData 拼接 动态信息
func FormatTrendData(trendData m.TrendList, userMap map[string]m.UserSimpleInfo, likeMap map[string]bool, focusMap map[string]uint8, countMap map[int64]map[string]string) {
	for i, data := range trendData {
		// 初始化互动数据 避免缓存数据存在
		interact := m.UserTrendInteract{}
		trendData[i].Interact = interact
		srtId := strconv.FormatInt(data.TrendId, 10)
		if like, ok := likeMap[srtId]; ok {
			trendData[i].Interact.IsLike = like
		}
		if focus, ok := focusMap[data.UserId]; ok {
			trendData[i].Interact.IsFocusAuthor = focus
		}
		if user, ok := userMap[data.UserId]; ok {
			trendData[i].UserName = user.UserName
			trendData[i].Avatar = user.Avatar
			trendData[i].VTag = user.VTag
			trendData[i].VStatus = user.VStatus
		}

		if count, ok := countMap[data.TrendId]; ok {
			trendData[i].Count.Likes, _ = strconv.Atoi(count["Likes"])
			trendData[i].Count.Comments, _ = strconv.Atoi(count["Comments"])
			trendData[i].Count.Forwards, _ = strconv.Atoi(count["Forwards"])
			trendData[i].Count.Collects, _ = strconv.Atoi(count["Collects"])
		}

		if data.ForwardInfo.Id != 0 && data.Forward != nil {
			trendData[i].Forward.Interact = interact
			forwardId := strconv.FormatInt(data.Forward.TrendId, 10)
			if like, ok := likeMap[forwardId]; ok {
				trendData[i].Forward.Interact.IsLike = like
			}
			if user, ok := userMap[data.Forward.UserId]; ok {
				trendData[i].Forward.UserName = user.UserName
				trendData[i].Forward.Avatar = user.Avatar
				trendData[i].Forward.VTag = user.VTag
				trendData[i].Forward.VStatus = user.VStatus
			}
			if count, ok := countMap[data.Forward.TrendId]; ok {
				trendData[i].Forward.Count.Likes, _ = strconv.Atoi(count["Likes"])
				trendData[i].Forward.Count.Comments, _ = strconv.Atoi(count["Comments"])
				trendData[i].Forward.Count.Forwards, _ = strconv.Atoi(count["Forwards"])
				trendData[i].Forward.Count.Collects, _ = strconv.Atoi(count["Collects"])
			}

		}
	}
}

// BatchGetTrendCount 批量查询动态统计
func BatchGetTrendCount(findData []m.MongoFeed) (result map[int64]map[string]string, err error) {
	if len(findData) == 0 {
		return
	}
	countData, needFind, err := c.BatchGetTrendCount(findData)
	if err != nil {
		logger.ErrZapLog(err, "BatchGetTrendCount cache fail")
	}

	var findArtCount []int64
	for _, data := range needFind {
		d := findData[data.Index]
		if d.Type == "aw" {
			findArtCount = append(findArtCount, d.MsgID)
		}
	}

	artCount, err := mysql.GetArtCount(findArtCount)
	if err != nil {
		err = errors.Wrap(err, "BatchGetTrendCount GetArtCount mysql fail")
		return
	}

	result = make(map[int64]map[string]string)
	for _, data := range artCount {
		tempMap := map[string]string{
			"Comments": strconv.Itoa(data.Comments),
			"Forwards": strconv.Itoa(data.Forwards),
			"Likes":    strconv.Itoa(data.Likes),
			"Collects": strconv.Itoa(data.Collects),
		}
		intId, _ := strconv.ParseInt(data.ArtworkId, 10, 64)
		countData[intId] = tempMap
	}

	result = countData

	return
}

// BatchGetArtworkCount 批量获取作品统计
func BatchGetArtworkCount(findIds []string) (result map[string]map[string]string, needCacheCount []m.ArtworkCount, err error) {
	if len(findIds) == 0 {
		return
	}

	var keys []string
	for _, id := range findIds {
		keys = append(keys, fmt.Sprintf(c.ArtworkCount, id))
	}

	countData, needFind, err := c.BatchGetTypeOfHash(keys)
	if err != nil {
		err = errors.Wrap(err, "GetBatchTrendCache BatchGetTypeOfString fail ")
		// 出错把 所有 Ids 写入 needFind
		for i, id := range findIds {
			needFind = append(needFind, c.NeedFindData{Id: id, Index: i})
		}
	}

	var findArtCount []string
	for _, find := range needFind {
		findArtCount = append(findArtCount, find.Id)
	}
	needCacheCount, err = mysql.GetArtCount(findArtCount)
	if err != nil {
		err = errors.Wrap(err, "BatchGetTrendCount GetArtCount mysql fail")
		return
	}

	result = make(map[string]map[string]string)
	for _, data := range needCacheCount {
		tempMap := map[string]string{
			"Comments": strconv.Itoa(data.Comments),
			"Forwards": strconv.Itoa(data.Forwards),
			"Likes":    strconv.Itoa(data.Likes),
			"Collects": strconv.Itoa(data.Collects),
		}
		result[data.ArtworkId] = tempMap
	}

	for k, v := range countData {
		result[strconv.FormatInt(k, 10)] = v
	}

	return
}

// BatchGetNotifyFactorInfo 批量获取提醒需要的要素
func BatchGetNotifyFactorInfo(uIds, findAw []string, findTr []int64) (
	artMap, trendMap map[string]m.NotifyArtOrTrendInfo, findData []m.BasicArtwork, err error) {

	artData, findData, err := BatchGetBasicArtInfo(findAw, uIds)
	if err != nil {
		err = errors.Wrap(err, "BatchGetNotifyFactorInfo BatchGetBasicArtInfo fail")
		return
	}

	trendData, err := mongo.GetNotifyTrendInfo(findTr)
	if err != nil {
		err = errors.Wrap(err, "BatchGetNotifyFactorInfo GetNotifyTrendInfo fail")
		return
	}

	artMap = make(map[string]m.NotifyArtOrTrendInfo)
	for _, art := range artData {
		// 拼接数据
		artDate := m.NotifyArtOrTrendInfo{
			Id:       art.ArtworkId,
			Cover:    art.Cover,
			IsDelete: art.IsDelete,
			Author:   art.UserId,
		}
		artMap[art.ArtworkId] = artDate
	}

	trendMap = make(map[string]m.NotifyArtOrTrendInfo)
	for _, tr := range trendData {
		trendMap[tr.Id] = tr
	}

	return
}

// GetUserNotifyConfig 获取用户通知配置缓存
func GetUserNotifyConfig(userId string) (config m.NotifyConfig, haveCache bool, err error) {
	key := fmt.Sprintf(c.NotifyConfig, userId)
	cache, err := c.GetAllHashValue(key)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetUserNotifyConfig chace fail %s", userId))
		return
	}
	if len(cache) != 0 {
		at, _ := strconv.ParseUint(cache["At"], 10, 64)
		like, _ := strconv.ParseUint(cache["Like"], 10, 64)
		comment, _ := strconv.ParseUint(cache["Comment"], 10, 64)
		follow, _ := strconv.ParseUint(cache["Follow"], 10, 64)
		message, _ := strconv.ParseUint(cache["Message"], 10, 64)
		collect, _ := strconv.ParseUint(cache["Collect"], 10, 64)
		config.At = uint8(at)
		config.Like = uint8(like)
		config.Comment = uint8(comment)
		config.Follow = uint8(follow)
		config.Message = uint8(message)
		config.Collect = uint8(collect)
		haveCache = true
		return
	}
	// 没有缓存到数据库查
	config, err = mysql.GetNotifySetting(userId)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetUserNotifyConfig myqsl fail %s", userId))
		return
	}

	return
}
