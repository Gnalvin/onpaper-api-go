package cache

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"onpaper-api-go/dao/mongo"
	m "onpaper-api-go/models"
	"strconv"
	"time"
)

func GetHotTrendId() (tIds []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tIds, err = Rdb.SRandMemberN(ctx, HotTrendAll, 30).Result()
	if err != nil {
		err = errors.Wrap(err, "GetHotTrend SRandMemberN fail")
		return
	}

	return
}

// GetBatchTrendCache 批量获取trend缓存
func GetBatchTrendCache(tIds []string) (trendData m.TrendList, needFind []NeedFindData, err error) {
	fmtKey := fmt.Sprintf(TrendProfile, "%s")
	var temp m.TrendShowInfo

	trendData, needFind, err = BatchGetTypeOfString(fmtKey, tIds, temp)
	if err != nil {
		err = errors.Wrap(err, "GetBatchTrendCache BatchGetTypeOfString fail ")
		// 出错把 所有 tIds 写入 needFind
		for i, id := range tIds {
			needFind = append(needFind, NeedFindData{Id: id, Index: i})
		}
	}

	return
}

// SetTrendCount 设置动态的统计缓存
func SetTrendCount(trendData []m.TrendShowInfo) (err error) {
	if len(trendData) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for _, data := range trendData {
		var key string
		if data.Type == "tr" {
			key = fmt.Sprintf(TrendCount, strconv.FormatInt(data.TrendId, 10))
		} else {
			key = fmt.Sprintf(ArtworkCount, strconv.FormatInt(data.TrendId, 10))
		}
		pipe.HSetNX(ctx, key, "Likes", data.Count.Likes)
		pipe.HSetNX(ctx, key, "Forwards", data.Count.Forwards)
		pipe.HSetNX(ctx, key, "Comments", data.Count.Comments)
		pipe.HSetNX(ctx, key, "Collects", data.Count.Collects)
		pipe.Expire(ctx, key, time.Hour*3*24)
	}
	_, err = pipe.Exec(ctx)
	return
}

// SetTrendForwards 设置转发数
func SetTrendForwards(info m.ForwardInfo) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var key string
	strId := strconv.FormatInt(info.Id, 10)
	if info.Type == "aw" {
		key = fmt.Sprintf(ArtworkCount, strId)
	} else {
		key = fmt.Sprintf(TrendCount, strId)
	}
	isExists, err := Rdb.Exists(ctx, key).Result()
	// 存在则 + 1
	if isExists == 1 {
		_, err = Rdb.HIncrBy(ctx, key, "Forwards", 1).Result()
	}

	return
}

func BatchGetTrendCount(findData []m.MongoFeed) (data map[int64]map[string]string, needFind []NeedFindData, err error) {
	var findKeys []string
	for _, d := range findData {
		strId := strconv.FormatInt(d.MsgID, 10)
		if d.Type == "aw" {
			findKeys = append(findKeys, fmt.Sprintf(ArtworkCount, strId))
		} else {
			findKeys = append(findKeys, fmt.Sprintf(TrendCount, strId))
		}
	}

	data, needFind, err = BatchGetTypeOfHash(findKeys)
	if err != nil {
		err = errors.Wrap(err, "GetBatchTrendCache BatchGetTypeOfString fail ")
		// 出错把 所有 tIds 写入 needFind
		for i, d := range findData {
			needFind = append(needFind, NeedFindData{Id: strconv.FormatInt(d.MsgID, 10), Index: i})
		}
	}
	return
}

// DeleteOneHotTrend 删除热门中的一个
func DeleteOneHotTrend(trendId interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = Rdb.SRem(ctx, HotTrendAll, trendId).Err()

	return
}

// CheckTrendCount 检测作品是否有统计缓存 没有则设置
func CheckTrendCount(trendId string) (err error) {
	key := fmt.Sprintf(TrendCount, trendId)
	isExists, err := CheckExistsKey(key)
	if err != nil {
		err = errors.Wrap(err, "CheckTrendCount CheckExistsKey fail")
		return
	}

	// 存在直接返回
	if isExists != 0 {
		return
	}
	tId, _ := strconv.ParseInt(trendId, 10, 64)
	//如果不存在需要查找 添加缓存之后再继续
	trendInfo, err := mongo.GetMoreTrendInfo([]int64{tId})
	if err != nil {
		err = errors.Wrap(err, "CheckTrendCount GetMoreTrendInfo fail")
		return
	}

	if len(trendInfo) == 0 {
		err = errors.New("CheckTrendCount no result")
		return
	}
	err = SetTrendCount(trendInfo)
	if err != nil {
		err = errors.Wrap(err, "CheckTrendCount SetTrendCount fail")
		return
	}

	return
}
