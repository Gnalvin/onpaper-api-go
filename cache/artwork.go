package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"time"
)

// SetArtworkProfile 设置作品页面资料缓存
func SetArtworkProfile(artwork m.ShowArtworkInfo) (err error) {
	key := fmt.Sprintf(ArtworkProfile, artwork.ArtworkId)

	err = SetOneStringValue(key, artwork, time.Hour*2)
	if err != nil {
		err = errors.Wrap(err, "SetArtworkProfile Cache Fail")
	}
	return
}

// GetArtworkProfile 获取作品页资料缓存
func GetArtworkProfile(artId, ip string) (res []redis.Cmder, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	// 获取 详情资料
	pkey := fmt.Sprintf(ArtworkProfile, artId)
	pipe.Get(ctx, pkey)
	//获取 数据资料
	cKey := fmt.Sprintf(ArtworkCount, artId)
	pipe.HGetAll(ctx, cKey)

	//添加浏览记录
	lKey := fmt.Sprintf(ArtworkHLog, artId)
	pipe.PFAdd(ctx, lKey, ip)
	pipe.PFCount(ctx, lKey)

	res, err = pipe.Exec(ctx)
	if err != nil {
		// 如果返回的错误是key不存在
		if errors.Is(err, redis.Nil) {
			return res, nil
		}
		err = errors.Wrap(err, "GetArtworkProfile Cache Fail")
	}
	return
}

// GetUserToArtInteract 查询用户对作品的互动数据 点赞、收藏、关注
func GetUserToArtInteract(userId, artId string) (res []redis.Cmder, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	collectKey := fmt.Sprintf(UserCollect, userId)
	likeKey := fmt.Sprintf(UserLike, userId)

	pipe.ZScore(ctx, collectKey, artId)
	pipe.ZScore(ctx, likeKey, artId)

	res, err = pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return res, nil
		}
		err = errors.Wrap(err, "GetUserToArtInteract Get Cache Fail,"+"userId:"+userId)
		return
	}

	return
}

// SetArtworkRank 设置作品排行缓存
func SetArtworkRank(rankType string, artworks []m.ArtworkRank) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf(RankArtwork, rankType)
	byteData, err := json.Marshal(&artworks)
	if err != nil {
		return err
	}
	str := string(byteData)
	_, err = Rdb.Set(ctx, key, str, time.Hour*24).Result()
	if err != nil {
		err = errors.Wrap(err, "SetArtworkRank Cache Fail")
		return err
	}
	return
}

// SetCollectCount 设置收藏数据
func SetCollectCount(userId string, collectData m.PostInteractData) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	setCount := true
	// 给 被关注的用户设置 userCount缓存
	_err := CheckUserCount(collectData.AuthorId)
	if _err != nil {
		// 如果出现错误 不对被收藏作品的用户 缓存设置 +-1 ，避免没有缓存时+-1 导致缓存数据错误
		setCount = false
		logger.ErrZapLog(_err, "SetCollectCount CheckUserCount fail ")
	}

	pipe := Rdb.Pipeline()
	// 作者统计缓存 key
	uKey := fmt.Sprintf(UserCount, collectData.AuthorId)
	// 作品统计缓存 key
	aKey := fmt.Sprintf(ArtworkCount, collectData.MsgId)
	// 给收藏者 收藏列表添加/删除
	cSortKey := fmt.Sprintf(UserCollect, userId)
	// 给收藏者 收藏作品数 +-1
	cKey := fmt.Sprintf(UserCount, userId)
	if !collectData.IsCancel {
		if setCount {
			pipe.HIncrBy(ctx, uKey, "Collects", 1)
		}
		pipe.HIncrBy(ctx, aKey, "Collects", 1)
		pipe.HIncrBy(ctx, cKey, "CollectCount", 1)
		pipe.ZAdd(ctx, cSortKey, redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: collectData.MsgId,
		})
	} else {
		if setCount {
			pipe.HIncrBy(ctx, uKey, "Collects", -1)
		}
		pipe.HIncrBy(ctx, aKey, "Collects", -1)
		pipe.HIncrBy(ctx, cKey, "CollectCount", -1)
		pipe.ZRem(ctx, cSortKey, collectData.MsgId)
	}
	// 用户首页有收藏数的缓存 也删除
	pKey := fmt.Sprintf(UserProfile, userId)
	pipe.Del(ctx, pKey)

	_, err = pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		err = errors.Wrap(err, "SetCollectCount Cache Fail")
		return err
	}
	return
}

// SetLikeCount 设置点赞数
func SetLikeCount(userId string, likeData m.PostInteractData) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setCount := true
	// 给 被关注的用户设置 userCount缓存
	_err := CheckUserCount(likeData.AuthorId)
	if _err != nil {
		// 如果出现错误 不对被喜欢用户 缓存设置 +-1 ，避免没有缓存时+-1 导致缓存数据错误
		setCount = false
		logger.ErrZapLog(_err, "SetLikeCount CheckUserCount fail ")
	}

	pipe := Rdb.Pipeline()
	// 作品统计缓存 key
	var aKey string
	if likeData.Type == "aw" {
		aKey = fmt.Sprintf(ArtworkCount, likeData.MsgId)
	} else {
		aKey = fmt.Sprintf(TrendCount, likeData.MsgId)
	}

	// 作者统计缓存 key
	uKey := fmt.Sprintf(UserCount, likeData.AuthorId)
	// 给点赞者 点赞列表添加/删除
	cSortKey := fmt.Sprintf(UserLike, userId)
	if !likeData.IsCancel {
		if setCount {
			pipe.HIncrBy(ctx, uKey, "Likes", 1)
		}
		pipe.HIncrBy(ctx, aKey, "Likes", 1)
		pipe.ZAdd(ctx, cSortKey, redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: likeData.MsgId,
		})
	} else {
		if setCount {
			pipe.HIncrBy(ctx, uKey, "Likes", -1)
		}
		pipe.HIncrBy(ctx, aKey, "Likes", -1)
		pipe.ZRem(ctx, cSortKey, likeData.MsgId)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		err = errors.Wrap(err, "SetLikeCount Cache Fail")
		return err
	}
	return
}

// SendCompressQueue 上传作品完成后 发送需要压缩的消息给 图片服务器
func SendCompressQueue(uid string, mid int64, mType string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = Rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: CompressStreamName,
		MaxLen: 5000,
		Approx: true,
		Values: map[string]interface{}{
			"uid":  uid,
			"mid":  mid,
			"type": mType,
		},
	}).Result()

	return
}

// GetHotArtwork 获取热门作品数据
func GetHotArtwork() (artData []m.HotArtworkData, artIds []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	artIds, err = Rdb.SRandMemberN(ctx, HotArtworkAll, 60).Result()
	if err != nil {
		err = errors.Wrap(err, "GetHotArtwork SRandMemberN fail")
		return
	}

	fmtKey := fmt.Sprintf(HotArtwork, "%s")
	var temp m.HotArtworkData
	artData, _, err = BatchGetTypeOfString(fmtKey, artIds, temp)
	if err != nil {
		err = errors.Wrap(err, "GetHotArtwork BatchGetTypeOfString fail")
		return
	}
	return
}

// GetBasicArtCache 批量获取基础作品缓存
func GetBasicArtCache(artIds []string) (artData []m.BasicArtwork, needFind []NeedFindData, err error) {
	fmtKey := fmt.Sprintf(ArtworkBasic, "%s")
	var temp m.BasicArtwork

	artData, needFind, err = BatchGetTypeOfString(fmtKey, artIds, temp)
	if err != nil {
		err = errors.Wrap(err, "GetBasicArtCache BatchGetTypeOfString fail ")
		// 出错把 所有 查找的artIds 写入 needFind
		for i, id := range artIds {
			needFind = append(needFind, NeedFindData{Id: id, Index: i})
		}
	}

	return
}

// SetArtworkCount 设置作品的统计缓存
func SetArtworkCount(countData []m.ArtworkCount) (err error) {
	if len(countData) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for _, data := range countData {
		key := fmt.Sprintf(ArtworkCount, data.ArtworkId)

		pipe.HSetNX(ctx, key, "Likes", data.Likes)
		pipe.HSetNX(ctx, key, "Forwards", data.Forwards)
		pipe.HSetNX(ctx, key, "Comments", data.Comments)
		pipe.HSetNX(ctx, key, "Collects", data.Collects)
		pipe.Expire(ctx, key, time.Hour*3*24)
	}
	_, err = pipe.Exec(ctx)

	return
}

// DeleteOneHotArtwork 删除热门中的一个
func DeleteOneHotArtwork(artId interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = Rdb.SRem(ctx, HotArtworkAll, artId).Err()

	return
}

// DeleteOneUserCollect 删除一个收藏作品
func DeleteOneUserCollect(userId string, artId interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	// 收藏列表删除一个
	cSortKey := fmt.Sprintf(UserCollect, userId)
	pipe.ZRem(ctx, cSortKey, artId)

	// 给收藏者 收藏作品数 +-1
	cKey := fmt.Sprintf(UserCount, userId)
	pipe.HIncrBy(ctx, cKey, "CollectCount", -1)

	_, err = pipe.Exec(ctx)

	return
}

func DeleteAboutArt(userId string, artId string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//删除 作品资料 缓存
	artProfileKey := fmt.Sprintf(ArtworkProfile, artId)
	//删除 作品封面 缓存
	artBasicsKey := fmt.Sprintf(ArtworkBasic, artId)
	//删除 动态缓存
	trendProfileKey := fmt.Sprintf(TrendProfile, artId)
	// 用户大卡片资料
	userBigCardKey := fmt.Sprintf(UserBigCard, userId)

	pipe := Rdb.Pipeline()
	pipe.Del(ctx, artProfileKey)
	pipe.Del(ctx, artBasicsKey)
	pipe.Del(ctx, trendProfileKey)
	pipe.Del(ctx, userBigCardKey)
	// 到热门中删除
	pipe.SRem(ctx, HotArtworkAll, artId)
	pipe.SRem(ctx, HotTrendAll, artId+"&aw&"+userId)
	// 作者作品数 - 1
	key := fmt.Sprintf(UserCount, userId)
	pipe.HIncrBy(ctx, key, "ArtCount", -1)

	_, err = pipe.Exec(ctx)
	if err != nil {
		// 如果返回的错误是key不存在
		if errors.Is(err, redis.Nil) {
			return nil
		}
		return err
	}

	return
}

// BatchSetArtViews 批量设置作品浏览量
func BatchSetArtViews(artIds []string, ip string) (err error) {
	if len(artIds) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	//添加浏览记录
	for _, artId := range artIds {
		lKey := fmt.Sprintf(ArtworkHLog, artId)
		pipe.PFAdd(ctx, lKey, ip)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "BatchSetArtViews fail")
	}
	return
}

func GetHotZoneArt(zoneIndex string) (artData []m.BasicArtwork, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zoneKey := fmt.Sprintf(HotZone, zoneIndex)
	artIds, err := Rdb.SRandMemberN(ctx, zoneKey, 15).Result()
	if err != nil {
		err = errors.Wrap(err, "GetHotZoneArt SRandMemberN fail")
		return
	}

	fmtKey := fmt.Sprintf(ArtworkBasic, "%s")
	var temp m.BasicArtwork
	artData, _, err = BatchGetTypeOfString(fmtKey, artIds, temp)
	if err != nil {
		err = errors.Wrap(err, "GetHotArtwork BatchGetTypeOfString fail")
		return
	}

	return
}

func SetUploadArtAbout(userId string, tagName []string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tagHotkey := fmt.Sprintf(RankTagHot)
	useCountKey := fmt.Sprintf(UserCount, userId)
	userBigCardKey := fmt.Sprintf(UserBigCard, userId)

	pipe := Rdb.Pipeline()
	//作品数 +1
	pipe.HIncrBy(ctx, useCountKey, "ArtCount", 1)
	pipe.Del(ctx, userBigCardKey)

	for _, s := range tagName {
		pipe.ZIncrBy(ctx, tagHotkey, 1, s)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "SetUploadArtAbout Cache Fail")
	}
	return
}

// CheckArtworkCount 检测作品是否有统计缓存 没有则设置
func CheckArtworkCount(artworkId string) (err error) {
	key := fmt.Sprintf(ArtworkCount, artworkId)
	isExists, err := CheckExistsKey(key)
	if err != nil {
		err = errors.Wrap(err, "CheckArtworkCount CheckExistsKey fail")
		return
	}

	// 存在直接返回
	if isExists != 0 {
		return
	}

	//如果不存在需要查找 添加缓存之后再继续
	count, err := mysql.GetArtCount([]string{artworkId})
	if err != nil {
		err = errors.Wrap(err, "CheckArtworkCount GetArtCount fail")
		return
	}

	if len(count) == 0 {
		err = errors.New("GetArtCount no result")
		return
	}

	// 设置作品统计数据缓存
	artworkCount := map[string]interface{}{
		"Likes":    count[0].Likes,
		"Collects": count[0].Collects,
		"Comments": count[0].Comments,
		"Forwards": count[0].Forwards,
	}
	cKey := fmt.Sprintf(ArtworkCount, artworkId)
	// 3天后过期 避免用户打开页面 不操作过了几小时后点赞 此时没有缓存 会导致数据出错
	err = BatchSetHashValue(cKey, artworkCount, 3*24)
	if err != nil {
		err = errors.Wrap(err, "CheckArtworkCount BatchSetHashValue fail")
		return
	}

	return
}
