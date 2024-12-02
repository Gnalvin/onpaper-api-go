package mysql

import (
	"encoding/json"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	m "onpaper-api-go/models"
	tools "onpaper-api-go/utils/formatTools"
	"strconv"
	"strings"
)

// GetBatchBasicShowArtInfo 批量获取作品展示的基本信息
func GetBatchBasicShowArtInfo(artIds, userId []string) (artData []m.BasicArtwork, err error) {
	if len(artIds) == 0 {
		return
	}
	var eg errgroup.Group

	var artworks []m.ArtSimpleInfo
	eg.Go(func() (mErr error) {
		artworks, mErr = GetBatchArtSimpleInfo(artIds)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetBatchBasicShowArtInfo GetBatchArtSimpleInfo fail")
		}
		return
	})

	userMap := make(map[string]m.UserSimpleInfo)
	eg.Go(func() (mErr error) {
		userMap, mErr = GetBatchUserSimpleInfo(userId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "GetBatchBasicShowArtInfo GetBatchUserSimpleInfo fail")
		}
		return
	})

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	// 拼接数据
	for _, artwork := range artworks {
		var temp m.BasicArtwork
		str, _ := json.Marshal(artwork)
		_ = json.Unmarshal(str, &temp)

		temp.UserAvatar = userMap[artwork.UserId].Avatar
		temp.UserName = userMap[artwork.UserId].UserName

		artData = append(artData, temp)
	}

	return
}

// GetBatchArtSimpleInfo 批量获取作品信息
func GetBatchArtSimpleInfo(artIds []string) (artworks []m.ArtSimpleInfo, err error) {
	sqlStr1 := `(SELECT a.artwork_id ,user_id,title,pic_count,cover,adults,whoSee,is_delete,a.createAT,first_pic,ap.width,ap.height
				from artwork a left join artwork_picture ap on a.artwork_id = ap.artwork_id
				WHERE a.artwork_id = ? and ap.sort = 0 limit 1 )`

	artCount := len(artIds)
	// 存放 select语句的 切片
	sqlStrings := make([]string, 0, artCount)
	artworkIds := make([]interface{}, 0, artCount)
	for _, artId := range artIds {
		sqlStrings = append(sqlStrings, sqlStr1)
		artworkIds = append(artworkIds, artId)
	}
	sqlStr1 = strings.Join(sqlStrings, " UNION ALL ")

	// 查询 数据库
	err = db.Select(&artworks, sqlStr1, artworkIds...)
	if err != nil {
		err = errors.Wrap(err, "GetBatchArtSimpleInfo: sql1 get fail")
	}

	return
}

// GetBatchUserSimpleInfo 批量获取用户短资料 头像\姓名\id
func GetBatchUserSimpleInfo(userIdList []string) (userMap map[string]m.UserSimpleInfo, err error) {
	if len(userIdList) == 0 {
		return
	}
	// 对 uidList 去重
	userIdList, _ = tools.RemoveSliceDuplicate(userIdList)
	// 需要查询的 所有用户列表
	sqlStrList := make([]string, 0)
	uIds := make([]interface{}, 0)

	sqlStr := `(SELECT user_id,username,avatar_name,v_tag,v_status,commission FROM user_profile WHERE user_id = ?  limit 1)`
	for _, uid := range userIdList {
		sqlStrList = append(sqlStrList, sqlStr)
		uIds = append(uIds, uid)
	}

	sqlStr = strings.Join(sqlStrList, " UNION ALL ")

	var userInfo []m.UserSimpleInfo
	// 查询 数据库
	err = db.Select(&userInfo, sqlStr, uIds...)
	if err != nil {
		err = errors.Wrap(err, "GetBatchUserSimpleInfo get fail")
		return
	}

	userMap = make(map[string]m.UserSimpleInfo)
	for _, u := range userInfo {
		userMap[u.UserId] = u
	}

	return
}

// GetBatchUserMoreInfo 批量获取用户短资料 + snsLink
func GetBatchUserMoreInfo(userIdList []string) (userInfo []m.UserSimpleInfoAndLink, err error) {
	// 需要查询的 所有用户列表
	sqlStrList := make([]string, 0)
	uIds := make([]interface{}, 0)
	//查询 作品的用户信息
	sqlStr := `(SELECT up.user_id,username,avatar_name,work_email,QQ,Weibo,Twitter,Pixiv,Bilibili,WeChat,v_status,left(introduce,180) as introduce
 				FROM user_profile as up 
 				left join user_intro ui on up.user_id = ui.user_id
 				WHERE up.user_id = ?  
            	limit 1)`
	for _, uid := range userIdList {
		sqlStrList = append(sqlStrList, sqlStr)
		uIds = append(uIds, uid)
	}

	sqlStr = strings.Join(sqlStrList, " UNION ALL ")

	// 查询 数据库
	err = db.Select(&userInfo, sqlStr, uIds...)
	if err != nil {
		err = errors.Wrap(err, "GetBatchUserMoreInfo get fail")
		return
	}

	return
}

// BatchGetUserNewArt 批量获取用户最近作品
func BatchGetUserNewArt(userIds []string, artCount uint8) (artworks []m.ArtSimpleInfo, err error) {

	sqlStr := `(SELECT artwork_id,title,cover,user_id,pic_count,adults,whoSee,createAT from artwork 
                WHERE user_id = ? and is_delete = 0 and whoSee ='public' and state = 0
				ORDER BY artwork_id DESC
				LIMIT ?)`

	userCount := len(userIds)
	// 存放 select语句的 切片
	sqlStrings := make([]string, 0, userCount)
	uIds := make([]interface{}, 0, userCount)

	for _, id := range userIds {
		sqlStrings = append(sqlStrings, sqlStr)
		uIds = append(uIds, id, artCount)
	}

	sqlStr = strings.Join(sqlStrings, " UNION ALL ")
	// 查询 数据库
	err = db.Select(&artworks, sqlStr, uIds...)
	if err != nil {
		err = errors.Wrap(err, "BatchGetUserNewArt: get fail")
		return
	}
	return
}

// BatchGetUserAllInfo 批量获取用户所有资料   头像\姓名\id + snsLink + 简介 + 最近作品
func BatchGetUserAllInfo(userIds []string, artCount uint8) (userData []m.UserBigCard, err error) {
	if len(userIds) == 0 {
		return make([]m.UserBigCard, 0), err
	}
	var eg errgroup.Group

	var userInfo []m.UserSimpleInfoAndLink
	eg.Go(func() (mErr error) {
		userInfo, mErr = GetBatchUserMoreInfo(userIds)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetUserAllInfo: GetBatchUserSimpleInfo get fail")
		}
		return
	})

	var artworks []m.ArtSimpleInfo
	eg.Go(func() (mErr error) {
		artworks, mErr = BatchGetUserNewArt(userIds, artCount)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetUserAllInfo: BatchGetUserNewArt get fail")
		}
		return
	})

	var userCountMap map[string]m.UserAllCountAndId
	eg.Go(func() (mErr error) {
		userCountMap, mErr = GetBatchUserCount(userIds)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetUserAllInfo: GetBatchUserCount get fail")
		}
		return
	})

	err = eg.Wait()
	if err != nil {
		return
	}

	artMap := make(map[string][]m.ArtSimpleInfo)
	for _, artwork := range artworks {
		if tempArr, ok := artMap[artwork.UserId]; ok {
			artMap[artwork.UserId] = append(tempArr, artwork)
		} else {
			artMap[artwork.UserId] = []m.ArtSimpleInfo{artwork}
		}
	}

	// 拼接数据
	for _, user := range userInfo {
		var tempUser m.UserBigCard
		str, _ := json.Marshal(user)
		_ = json.Unmarshal(str, &tempUser)

		if arr, ok := artMap[user.UserId]; ok {
			tempUser.Artworks = arr
		} else {
			tempUser.Artworks = make([]m.ArtSimpleInfo, 0)
		}

		if count, ok := userCountMap[user.UserId]; ok {
			tempUser.Count = count
		}

		userData = append(userData, tempUser)
	}

	return
}

func BatchGetTrendArtInfo(artIds []int64) (trendArt m.TrendList, err error) {
	if len(artIds) == 0 {
		return
	}
	artIds, _ = tools.RemoveSliceDuplicate(artIds)
	sqlStr1 := `(SELECT artwork_id, description from art_intro WHERE artwork_id = ? limit 1 )`
	sqlStr2 := `(SELECT artwork_id, filename,sort,width,height from artwork_picture WHERE artwork_id = ? limit 15 )`
	sqlStr4 := `(SELECT artwork_id,user_id,comment,whoSee,is_delete,createAt from artwork WHERE artwork_id = ? limit 1 )`

	artCount := len(artIds)
	// 存放 select语句的 切片
	sql1Strings := make([]string, 0, artCount)
	sql2Strings := make([]string, 0, artCount)
	sql4Strings := make([]string, 0, artCount)
	artworkIds := make([]interface{}, 0, artCount)
	for _, artId := range artIds {
		sql1Strings = append(sql1Strings, sqlStr1)
		sql2Strings = append(sql2Strings, sqlStr2)
		sql4Strings = append(sql4Strings, sqlStr4)
		artworkIds = append(artworkIds, strconv.FormatInt(artId, 10))
	}
	sqlStr1 = strings.Join(sql1Strings, " UNION ALL ")
	sqlStr2 = strings.Join(sql2Strings, " UNION ALL ")
	sqlStr4 = strings.Join(sql4Strings, " UNION ALL ")

	var eg errgroup.Group
	var intro []m.ArtIntro
	// 查询 数据库
	eg.Go(func() (mErr error) {
		mErr = db.Select(&intro, sqlStr1, artworkIds...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetTrendArtInfo: sql1 get fail")
		}
		return
	})

	var pics []m.ArtPic
	eg.Go(func() (mErr error) {
		mErr = db.Select(&pics, sqlStr2, artworkIds...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetTrendArtInfo: sql2 get fail")
		}
		return
	})

	trendArt = make([]m.TrendShowInfo, artCount, artCount)
	eg.Go(func() (mErr error) {
		mErr = db.Select(&trendArt, sqlStr4, artworkIds...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetTrendArtInfo: sql4 get fail")
		}
		return
	})

	err = eg.Wait()
	if err != nil {
		return
	}

	picMap := make(map[string][]m.PicsType)
	for _, pic := range pics {
		picMap[pic.ArtworkId] = append(picMap[pic.ArtworkId], m.PicsType{
			FileName: pic.Pic,
			Sort:     pic.Sort,
			Width:    pic.Width,
			Height:   pic.Height,
		})
	}

	introMap := make(map[string]string)
	for _, c := range intro {
		introMap[c.ArtworkId] = c.Intro
	}

	for i, trend := range trendArt {
		strId := strconv.FormatInt(trend.TrendId, 10)
		trendArt[i].Intro = introMap[strId]
		trendArt[i].Pics = picMap[strId]
		trendArt[i].Type = "aw"
	}

	return
}

// GetBatchUserCount 批量获取用户统计
func GetBatchUserCount(userIds []string) (userCountMap map[string]m.UserAllCountAndId, err error) {
	sqlStr1 := `(SELECT user_id,fans, likes,collects,art_count,collect_count,following,trend_count from user_count WHERE user_id = ? limit 1 )`

	count := len(userIds)
	// 存放 select语句的 切片
	sql1Strings := make([]string, 0, count)
	uIds := make([]interface{}, 0, count)
	for _, uId := range userIds {
		sql1Strings = append(sql1Strings, sqlStr1)
		uIds = append(uIds, uId)
	}

	var userCount []m.UserAllCountAndId
	userCountMap = make(map[string]m.UserAllCountAndId)

	sqlStr1 = strings.Join(sql1Strings, " UNION ALL ")
	err = db.Select(&userCount, sqlStr1, uIds...)
	if err != nil {
		err = errors.Wrap(err, "GetBatchUserCount fail")
	}

	for _, data := range userCount {
		userCountMap[data.UserId] = data
	}

	return
}

// BatchGetUserBaseInfo 获取用户小卡片数据
func BatchGetUserBaseInfo(userIdList []string) (userInfo []m.UserSmallCard, err error) {
	if len(userIdList) == 0 {
		return
	}
	// 需要查询的 所有用户列表
	sqlStrList := make([]string, 0)
	uIds := make([]interface{}, 0)

	sqlStr := `(SELECT up.user_id,username,avatar_name,v_status,v_tag,left(introduce,30) as introduce
 				FROM user_profile as up 
 				left join user_intro ui on up.user_id = ui.user_id
 				WHERE up.user_id = ?  
            	limit 1)`
	for _, uid := range userIdList {
		sqlStrList = append(sqlStrList, sqlStr)
		uIds = append(uIds, uid)
	}
	sqlStr = strings.Join(sqlStrList, " UNION ALL ")

	var eg errgroup.Group

	eg.Go(func() (mErr error) {
		err = db.Select(&userInfo, sqlStr, uIds...)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetUserBaseInfo get fail")
		}
		return
	})

	var userCountMap map[string]m.UserAllCountAndId
	eg.Go(func() (mErr error) {
		userCountMap, mErr = GetBatchUserCount(userIdList)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "BatchGetUserBaseInfo: GetBatchUserCount get fail")
		}
		return
	})

	err = eg.Wait()
	if err != nil {
		return
	}

	for i, card := range userInfo {
		if count, ok := userCountMap[card.UserId]; ok {
			userInfo[i].Count = m.UserAllCount{
				Fans:         count.Fans,
				Likes:        count.Likes,
				Collects:     count.Collects,
				Following:    count.Following,
				TrendCount:   count.TrendCount,
				ArtCount:     count.ArtCount,
				CollectCount: count.CollectCount,
			}
		}
	}

	return
}
