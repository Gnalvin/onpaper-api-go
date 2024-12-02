package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"onpaper-api-go/dao/mysql"
	m "onpaper-api-go/models"
	tools "onpaper-api-go/utils/formatTools"
	"strconv"
	"time"
)

func CreatUserID() (uid int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	uid, err = Rdb.IncrBy(ctx, ServeUserId, 1).Result()
	if err != nil {
		err = errors.Wrap(err, "CreatUserID redis fail")
	}
	return
}

// SetUserHomeProfile 设置对用户资料缓存
func SetUserHomeProfile(profile m.UserProfileTableInfo) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	byteData, err := json.Marshal(&profile)
	if err != nil {
		return err
	}
	key := fmt.Sprintf(UserProfile, profile.UserId)
	str := string(byteData)
	_, err = Rdb.Set(ctx, key, str, time.Hour*2).Result()
	if err != nil {
		err = errors.Wrap(err, "SetUserHomeProfile Cache Fail")
		return err
	}
	return
}

// SetUserFocusUserId 设置用户关注列表
func SetUserFocusUserId(userId string, focusList []interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	key := fmt.Sprintf(UserFollow, userId)
	pipe.SAdd(ctx, key, focusList...)
	// 过期时间比 RefreshToken 时间长 确保登录有效时间都存在
	pipe.Expire(ctx, key, 200*time.Hour*24)
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("SetUserFocusUserId Cache fail %s", userId))
		return
	}

	return
}

// SetUserLikeArtId 设置用户点赞过的作品id缓存
func SetUserLikeArtId(userId string, likeList []m.InitUserData) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	key := fmt.Sprintf(UserLike, userId)

	sortItem := make([]redis.Z, 0)
	for _, data := range likeList {
		sortItem = append(sortItem, redis.Z{
			Score:  float64(data.Time.Unix()),
			Member: data.MsgId,
		})
	}

	pipe.ZAdd(ctx, key, sortItem...)
	// 过期时间比 RefreshToken 时间长 确保登录有效时间都存在
	pipe.Expire(ctx, key, 200*time.Hour*24)
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("SetUserLikeArtId Cache fail %s", userId))
		return
	}

	return
}

// SetUerCollectId 设置用户收藏作品id
func SetUerCollectId(userId string, collectList []m.InitUserData) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	key := fmt.Sprintf(UserCollect, userId)

	sortItem := make([]redis.Z, 0)
	for _, data := range collectList {
		sortItem = append(sortItem, redis.Z{
			Score:  float64(data.Time.Unix()),
			Member: data.MsgId,
		})
	}

	pipe.ZAdd(ctx, key, sortItem...)
	// 过期时间比 RefreshToken 时间长 确保登录有效时间都存在
	pipe.Expire(ctx, key, 200*time.Hour*24)
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("SetUerCollectId Cache fail %s", userId))
		return
	}

	return
}

// SetUserRank 设置用户排名缓存
func SetUserRank(rankType string, userData []m.UserBigCard) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf(RankUser, rankType)
	byteData, err := json.Marshal(&userData)
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

// SetUserCount 设置用户点赞收藏分数数缓存
func SetUserCount(uCount m.UserAllCount, userId string) (err error) {
	// 设置作者统计数据缓存
	authorCount := map[string]interface{}{
		"Likes":        uCount.Likes,
		"Fans":         uCount.Fans,
		"Collects":     uCount.Collects,
		"Follows":      uCount.Following,
		"ArtCount":     uCount.ArtCount,
		"CollectCount": uCount.CollectCount,
		"TrendCount":   uCount.TrendCount,
	}
	key := fmt.Sprintf(UserCount, userId)
	// 如果有缓存时使用缓存数据。没有缓存才会设置 因为 数据库的数据可能比较老
	// 因为不同作品过期时间不一定什么时候，不能依靠作品过期时间来设置 Author 的数据
	err = BatchSetHashValue(key, authorCount, 200*24)
	return

}

// GetUserCount 获取用户统计数
func GetUserCount(userId string) (userCount CtxCacheVale, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf(UserCount, userId)
	res, err := Rdb.HGetAll(ctx, key).Result()
	if err != nil {
		err = errors.Wrap(err, "GetUserCount HGetAll fail")
		return
	}
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
	return
}

// GetHotUserId 获取热门用户的id
func GetHotUserId(loginId string) (userIds []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fKey := fmt.Sprintf(UserFollow, loginId)

	userIds, err = Rdb.SDiff(ctx, HotUserAll, fKey).Result()
	if err != nil {
		return
	}

	return
}

// CheckUserLike 批量检查 用户是否点赞过
func CheckUserLike(ids []string, userId string) (res map[string]bool, err error) {
	if userId == "" {
		return
	}
	ids, _ = tools.RemoveSliceDuplicate(ids)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := fmt.Sprintf(UserLike, userId)
	pipe := Rdb.Pipeline()
	for _, id := range ids {
		pipe.ZScore(ctx, key, id)
	}
	cmder, err := pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
		} else {
			err = errors.Wrap(err, fmt.Sprintf("userId: %s", userId))
			return
		}
	}
	res = make(map[string]bool, 0)
	for i, cmd := range cmder {
		val := cmd.(*redis.FloatCmd).Val() != 0
		res[ids[i]] = val
	}

	return
}

// CheckUserFollow 批量检查 用户是否关注过
func CheckUserFollow(checkIds []string, userId string) (focusStatus map[string]uint8, err error) {
	checkIds, _ = tools.RemoveSliceDuplicate(checkIds)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := fmt.Sprintf(UserFollow, userId)
	pipe := Rdb.Pipeline()
	for _, id := range checkIds {
		pipe.SIsMember(ctx, key, id)
	}
	cmder, err := pipe.Exec(ctx)
	if err != nil {
		return
	}

	focusStatus = make(map[string]uint8, 0) // 没关注 0 关注 1 相互关注 2
	// 如果有关注的 需要再次查找 是否互相关注
	var checkAgain []string
	for i, cmd := range cmder {
		val := cmd.(*redis.BoolCmd).Val()
		isFocus := uint8(0)
		if val {
			isFocus = 1
			checkAgain = append(checkAgain, checkIds[i])
		}
		focusStatus[checkIds[i]] = isFocus
	}

	if len(checkAgain) == 0 {
		return
	}
	for _, id := range checkAgain {
		_key := fmt.Sprintf(UserFollow, id)
		pipe.Exists(ctx, _key)
		pipe.SIsMember(ctx, _key, userId)
	}

	cmder, err = pipe.Exec(ctx)
	if err != nil {
		return
	}

	n := 0
	var checkMysql []string
	for i, cmd := range cmder {
		if i%2 != 0 {
			continue
		}
		uid := checkAgain[n]
		n += 1
		// 是否存在key 没有的到数据中查
		isExists := cmd.(*redis.IntCmd).Val()
		if isExists != 1 {
			checkMysql = append(checkMysql, uid)
			continue
		}
		val := cmder[i+1].(*redis.BoolCmd).Val()
		if val {
			focusStatus[uid] = 2
		}
	}
	// 到数据库中查找
	mysqlRes, err := mysql.CheckUserFocus(checkMysql, userId)
	if err != nil {
		err = errors.Wrap(err, "CheckUserFollow fail")
		return
	}

	// 如果另一个用户也关注了 则是互相关注 2
	for _, r := range mysqlRes {
		isFocus := uint8(1)
		if r.IsFocus == 1 {
			isFocus = 2
		}
		focusStatus[r.UserId] = isFocus
	}

	return
}

// CheckUserCount 检测用户统计数是否有缓存
func CheckUserCount(userId string) (err error) {
	// 作者统计缓存 key
	uKey := fmt.Sprintf(UserCount, userId)
	isExists, err := CheckExistsKey(uKey)
	if err != nil {
		err = errors.Wrap(err, "CheckUserCount CheckExistsKey fail")
		return
	}

	if isExists != 0 {
		return
	}

	//如果不存在需要查找 添加缓存之后再继续
	count, err := mysql.GetUserCount(userId)
	if err != nil {
		err = errors.Wrap(err, "CheckUserCount GetUserCount fail")
		return
	}

	err = SetUserCount(count, userId)
	if err != nil {
		err = errors.Wrap(err, "CheckUserCount SetUserCount fail")
		return
	}

	return
}

// SetFocusCount 对关注做缓存事件处理
func SetFocusCount(userId string, focusData m.VerifyUserFocus) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	setCount := true
	// 给 被关注的用户设置 userCount缓存
	_err := CheckUserCount(focusData.FocusId)
	if _err != nil {
		// 如果出现错误 不对被关注用户 缓存 设置 fans +-1 ，避免没有缓存时+-1 导致缓存数据错误
		setCount = false
	}

	pipe := Rdb.Pipeline()
	// 被关注人统计缓存 key
	focusUserKey := fmt.Sprintf(UserCount, focusData.FocusId)
	// 登陆用户统计缓存 key
	loginUserKey := fmt.Sprintf(UserCount, userId)
	// 给关注者 关注列表添加/删除
	cSortKey := fmt.Sprintf(UserFollow, userId)

	if !focusData.IsCancel {
		if setCount {
			pipe.HIncrBy(ctx, focusUserKey, "Fans", 1)
		}
		pipe.HIncrBy(ctx, loginUserKey, "Follows", 1)
		pipe.SAdd(ctx, cSortKey, focusData.FocusId)
	} else {
		if setCount {
			pipe.HIncrBy(ctx, focusUserKey, "Fans", -1)
		}
		pipe.HIncrBy(ctx, loginUserKey, "Follows", -1)
		pipe.SRem(ctx, cSortKey, focusData.FocusId)
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

// SetUserCountField 设置用户统计某个字段+1
func SetUserCountField(userId, field string, number int64) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf(UserCount, userId)
	err = Rdb.HIncrBy(ctx, key, field, number).Err()
	if err != nil {
		err = errors.Wrap(err, "SetUserCountField fail")
	}
	return
}

func SetUserBigCarCache(userData []m.UserBigCard) (err error) {
	if len(userData) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for _, d := range userData {
		byteData, _err := json.Marshal(&d)
		if _err != nil {
			err = errors.Wrap(_err, "SetUserBigCarCache json Marshal fail")
			return err
		}
		key := fmt.Sprintf(UserBigCard, d.UserId)
		str := string(byteData)
		pipe.Set(ctx, key, str, time.Hour*24)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "SetUserBigCarCache pipe.Exec fail")
	}
	return
}

func GetUserBigCarCache(uIds []string) (userData []m.UserBigCard, needFind []NeedFindData, err error) {
	fmtKey := fmt.Sprintf(UserBigCard, "%s")
	var temp m.UserBigCard

	userData, needFind, err = BatchGetTypeOfString(fmtKey, uIds, temp)
	if err != nil {
		err = errors.Wrap(err, "GetUserBigCarCache BatchGetTypeOfString fail ")
		// 出错把 所有 Ids 写入 needFind
		for i, id := range uIds {
			needFind = append(needFind, NeedFindData{Id: id, Index: i})
		}
	}

	return
}

func SetUserSmallCarCache(userData []m.UserSmallCard) (err error) {
	if len(userData) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for _, d := range userData {
		byteData, _err := json.Marshal(&d)
		if _err != nil {
			err = errors.Wrap(_err, "SetUserSmallCarCache json Marshal fail")
			return err
		}
		key := fmt.Sprintf(UserSmallCard, d.UserId)
		str := string(byteData)
		pipe.Set(ctx, key, str, time.Hour*24)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "SetUserSmallCarCache pipe.Exec fail")
	}
	return
}

func GetUserSmallCarCache(uIds []string) (userData []m.UserSmallCard, needFind []NeedFindData, err error) {
	fmtKey := fmt.Sprintf(UserSmallCard, "%s")
	var temp m.UserSmallCard

	userData, needFind, err = BatchGetTypeOfString(fmtKey, uIds, temp)
	if err != nil {
		err = errors.Wrap(err, "GetUserBigCarCache BatchGetTypeOfString fail ")
		// 出错把 所有 Ids 写入 needFind
		for i, id := range uIds {
			needFind = append(needFind, NeedFindData{Id: id, Index: i})
		}
	}

	return
}
