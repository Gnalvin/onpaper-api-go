package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"onpaper-api-go/logger"
	"strconv"
	"strings"
	"time"
)

// GetCmderStringResult 获取cmder 里面的 结果
func GetCmderStringResult(cmder redis.Cmder) (res string) {
	err := cmder.Err()
	if err != nil {
		// 如果返回的错误 是key不存在
		if errors.Is(err, redis.Nil) {
			return ""
		}
		err = errors.Wrap(err, fmt.Sprintf("redis Pipeline Fail %s", cmder.Args()))
		// 记录错误信息
		logger.ErrZapLog(err, cmder.String())
		return
	}
	var temp []string
	for _, arg := range cmder.Args() {
		temp = append(temp, arg.(string))
	}

	sepStr := strings.Join(temp, " ") + ": "
	strList := strings.Split(cmder.String(), sepStr)
	res = strList[1]

	return
}

// GetOneStringValue 获取 一条 Redis 缓存中 类型为 string 的值
func GetOneStringValue[T any](key string, temp T) (ctxVal CtxCacheVale, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		// 如果返回的错误是key不存在
		if errors.Is(err, redis.Nil) {
			err = nil
			return
		}
		// 出其他错了
		err = errors.Wrap(err, fmt.Sprintf("GetOneStringValue Cache fail %s", key))
		return
	}

	err = json.Unmarshal([]byte(val), &temp)
	if err != nil {
		err = errors.Wrap(err, "GetOneStringValue Unmarshal fail ")
		return
	} else {
		ctxVal.Val = temp
		ctxVal.HaveCache = true
	}

	return
}

// SetOneStringValue 把结构体序列化后保存到redis
func SetOneStringValue[T any](key string, info T, timeOut time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	byteData, err := json.Marshal(&info)
	if err != nil {
		return err
	}
	str := string(byteData)
	_, err = Rdb.Set(ctx, key, str, timeOut).Result()
	if err != nil {
		err = errors.Wrap(err, "SetOneStringValue Cache Fail")
	}
	return
}

// BatchGetTypeOfString 批量获取 类型为 string 的值
func BatchGetTypeOfString[T any](fmtKey string, finds []string, format T) (data []T, needFind []NeedFindData, err error) {
	if len(finds) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for _, find := range finds {
		key := fmt.Sprintf(fmtKey, find)
		pipe.Get(ctx, key)
	}

	cmders, err := pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
		} else {
			// 如果错误 不是 key不存在 返回错误
			return
		}
	}
	//储存空值
	temp := format
	for index, c := range cmders {
		val := c.(*redis.StringCmd).Val()
		if val == "" {
			id := strings.Split(c.Args()[1].(string), ":")[2]
			needFind = append(needFind, NeedFindData{
				Id:    id,
				Index: index,
			})
			continue
		}
		err = json.Unmarshal([]byte(val), &format)
		data = append(data, format)
		format = temp
	}

	return
}

// BatchSetTypeOfString 批量设置string 缓存
func BatchSetTypeOfString[T any](fmtKey string, sets []string, info []T, duration time.Duration) (err error) {

	if len(sets) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for i, set := range sets {
		key := fmt.Sprintf(fmtKey, set)
		temp := info[i]
		byteData, mErr := json.Marshal(&temp)
		if mErr != nil {
			return mErr
		}
		str := string(byteData)
		pipe.Set(ctx, key, str, duration)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "BatchSetTypeOfString fail")
	}

	return
}

// DelOneCache 删除 一条缓存
func DelOneCache(key string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = Rdb.Del(ctx, key).Result()
	if err != nil {
		// 如果返回的错误是key不存在
		if errors.Is(err, redis.Nil) {
			return nil
		}
		// 出其他错了
		err = errors.Wrap(err, fmt.Sprintf("DelOneCache Cache fail %s", key))
		return
	}
	return
}

// BatchDelCache 批量删除缓存
func BatchDelCache(keys []string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
		} else {
			// 如果错误 不是 key不存在 返回错误
			return
		}
	}
	return
}

// SetHashKeyValue 设置hash类型的 缓存 当key 存在 则覆盖
func SetHashKeyValue(key string, info map[string]interface{}, timeOut time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	pipe.HSet(ctx, key, info)
	pipe.Expire(ctx, key, timeOut*time.Hour)
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("SetHashKeyValue Cache Fail %s", key))
	}
	return
}

// BatchSetHashValue 批量设置hash的value 当key 不存在
func BatchSetHashValue(key string, info map[string]interface{}, timeOut time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for k, v := range info {
		pipe.HSetNX(ctx, key, k, v)
	}
	// 不管结果如何 都延长过期时间
	pipe.Expire(ctx, key, timeOut*time.Hour)
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("BatchSetHashValue Cache Fail %s", key))
	}

	return err
}

// CheckExistsKey 查看某个Key 是否存在
func CheckExistsKey(key string) (isExists int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isExists, err = Rdb.Exists(ctx, key).Result()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("checkExistsKey get Fail %s", key))
		return
	}

	return
}

func BatchGetTypeOfHash(Keys []string) (data map[int64]map[string]string, needFind []NeedFindData, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	for _, key := range Keys {
		pipe.HGetAll(ctx, key)
	}

	cmders, err := pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
		} else {
			// 如果错误 不是 key不存在 返回错误
			return
		}
	}
	data = make(map[int64]map[string]string)
	//储存空值
	for index, c := range cmders {
		val := c.(*redis.MapStringStringCmd).Val()
		id := strings.Split(c.Args()[1].(string), ":")[2]
		if len(val) == 0 {
			needFind = append(needFind, NeedFindData{
				Id:    id,
				Index: index,
			})
			continue
		}
		intId, _ := strconv.ParseInt(id, 10, 64)
		data[intId] = val
	}
	return
}

// GetAllHashValue 获取hash 键值对
func GetAllHashValue(key string) (res map[string]string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err = Rdb.HGetAll(ctx, key).Result()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("GetAllHashValue fail %s", key))
	}
	return
}

// CheckZSortLen 检查 ZSort 的长度，如果超出则删除部分
func CheckZSortLen(key string, maxSize, removeCount int64) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	size, err := Rdb.ZCard(ctx, key).Result()
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("CheckZSortLen fail key:%s", key))
		return
	}

	if size > maxSize {
		// 删除最后 100 个元素
		_, err = Rdb.ZRemRangeByRank(ctx, key, 0, removeCount-1).Result()
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("CheckZSortLen fail key:%s", key))
		}
	}
	return
}
