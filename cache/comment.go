package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
	"time"
)

// SetRootComment 设置根评论缓存
func SetRootComment(comments []m.ReturnComment, artId, cType string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zSortKey := fmt.Sprintf(CommentRootList, artId)
	count, err := Rdb.ZCount(ctx, zSortKey, "-inf", "+inf").Result()
	if err != nil {
		logger.ErrZapLog(err, "SetRootComment ZCount fail ")
	}
	//一个sort 只保存 300条 超过不再添加缓存
	if count >= 300 {
		return
	}

	pipe := Rdb.Pipeline()

	//1.构建 sort 需要的数据
	var zSetList []redis.Z
	for i, comment := range comments {
		// 如果有子评论要设置 计数
		for ci, childComment := range comment.ChildComments {
			comments[i].ChildComments[ci].IsLike = false // 重置是否点赞
			childId := strconv.FormatInt(childComment.CId, 10)
			cKey := fmt.Sprintf(CommentCount, childId)
			pipe.HSetNX(ctx, cKey, "Likes", comment.Likes)
			pipe.Expire(ctx, cKey, 3*24*time.Hour)
		}

		sortItem := redis.Z{Member: comment.CId, Score: float64(comment.Sore)}
		zSetList = append(zSetList, sortItem)
		cId := strconv.FormatInt(comment.CId, 10)

		//单个评论缓存
		comment.IsLike = false // 重置是否点赞
		commentKey := fmt.Sprintf(CommentDetail, cId)
		byteData, jsonErr := json.Marshal(&comment)
		if jsonErr != nil {
			return errors.Wrap(jsonErr, "SetRootComment Marshal Fail")
		}
		pipe.Set(ctx, commentKey, byteData, 24*time.Hour)

		// 设置单个评论的计数
		countKey := fmt.Sprintf(CommentCount, cId)
		pipe.HSetNX(ctx, countKey, "Likes", comment.Likes)
		pipe.Expire(ctx, countKey, 3*24*time.Hour)

	}

	//通过评论组不够 20条 说明到了最后一页 在缓存中添加标识
	commentLen := len(comments)
	if commentLen < 20 {
		sortItem := redis.Z{Member: "end", Score: 0}
		zSetList = append(zSetList, sortItem)
	}

	// 添加到 redis中
	pipe.ZAdd(ctx, zSortKey, zSetList...)
	pipe.Expire(ctx, zSortKey, 24*time.Hour*7)
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "SetRootComment Cache Fail")
		return
	}

	return
}

// GetRootCommentId 获取作品根评论缓存id
func GetRootCommentId(ownId, sore string, count int) (res []redis.Cmder, isEnd bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zSortKey := fmt.Sprintf(CommentRootList, ownId)
	opt := &redis.ZRangeBy{Max: sore, Min: "-inf", Offset: 0, Count: int64(count)}
	cidList, err := Rdb.ZRevRangeByScore(ctx, zSortKey, opt).Result()
	if err != nil {
		err = errors.Wrap(err, "GetRootCommentId ZRangeBy Cache Fail")
		return
	}

	pipe := Rdb.Pipeline()
	for i, cid := range cidList {
		// 说明没有评论
		if cid == "end" && i == 0 {
			isEnd = true
			break
		}
		// 最后一个是 end 不要查询
		if cid == "end" {
			isEnd = true
			break
		}
		//如果不是第一页 第一条都不要 因为是重复的
		if sore != "+inf" && i == 0 {
			continue
		}
		commentKey := fmt.Sprintf(CommentDetail, cid)
		pipe.Get(ctx, commentKey)
	}

	// 没有评论的情况
	if isEnd && len(cidList) == 0 {
		return
	}

	res, err = pipe.Exec(ctx)
	if err != nil {
		// 如果返回的错误是key不存在 说明有评论主题的缓存丢失 到数据库查询
		if errors.Is(err, redis.Nil) {
			return
		}
		// 出其他错了
		err = errors.Wrap(err, "GetRootCommentId Get Cache Fail")
		return
	}

	return
}

// AddOneComment 添加一个评论到评论缓存队列
func AddOneComment(comment m.ReturnComment, cType string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()

	countKey := ""
	if cType == "aw" {
		countKey = fmt.Sprintf(ArtworkCount, comment.OwnId)
	} else {
		countKey = fmt.Sprintf(TrendCount, comment.OwnId)
	}
	// 给评论数 + 1
	pipe.HIncrBy(ctx, countKey, "Comments", 1)

	//评论zSort缓存
	zSortKey := ""
	commentKey := ""
	cId := strconv.FormatInt(comment.CId, 10)
	// = 0 是根回复
	if comment.RootId == 0 {
		zSortKey = fmt.Sprintf(CommentRootList, comment.OwnId)
	} else {
		zSortKey = fmt.Sprintf(CommentChildList, strconv.FormatInt(comment.RootId, 10))
	}
	commentKey = fmt.Sprintf(CommentDetail, cId)

	// 缓存到集合
	sortItem := redis.Z{Member: comment.CId, Score: float64(comment.Sore)}
	pipe.ZAdd(ctx, zSortKey, sortItem)

	//单个评论缓存
	byteData, jsonErr := json.Marshal(&comment)
	if jsonErr != nil {
		return errors.Wrap(jsonErr, "AddOneComment Marshal Fail")
	}
	pipe.Set(ctx, commentKey, byteData, 24*time.Hour)

	// 设置单个评论的计数
	commentCountKey := fmt.Sprintf(CommentCount, cId)
	pipe.HSetNX(ctx, commentCountKey, "Likes", comment.Likes)
	pipe.Expire(ctx, commentCountKey, 3*24*time.Hour)

	//如果不是跟回复 还要删除 回复的那个评论 否则 总回复个数会不准确,导致前端点击不了进入子回复列表
	if comment.RootId != 0 {
		rootCommentKey := fmt.Sprintf(CommentDetail, strconv.FormatInt(comment.RootId, 10))
		pipe.Del(ctx, rootCommentKey)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "AddOneComment Cache Fail")
		return
	}

	// 查看是否超过了310个 超过则删除 最后一个
	count, err := Rdb.ZCount(ctx, zSortKey, "-inf", "+inf").Result()
	if err != nil {
		logger.ErrZapLog(err, "SetRootComment ZCount fail ")
	}
	if count > 310 {
		if comment.RootId == 0 {
			_, err = Rdb.ZRemRangeByRank(ctx, zSortKey, 0, 0).Result()
		} else {
			_, err = Rdb.ZRemRangeByRank(ctx, zSortKey, -1, -1).Result()
		}
	}
	if err != nil {
		err = errors.Wrap(err, "AddOneComment ZRemRangeByRank Fail")
		return
	}
	return
}

// SetChildComment 设置子评论的缓存
func SetChildComment(rootComments []m.Comment, rootId string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zSortKey := fmt.Sprintf(CommentChildList, rootId)
	count, err := Rdb.ZCount(ctx, zSortKey, "-inf", "+inf").Result()
	if err != nil {
		logger.ErrZapLog(err, "SetChildComment ZCount fail ")
	}
	//一个sort 只保存 300条 超过不再添加缓存
	if count >= 300 {
		return
	}

	pipe := Rdb.Pipeline()

	//1.构建 sort 需要的数据
	var zSetList []redis.Z
	for _, comment := range rootComments {
		comment.IsLike = false
		sortItem := redis.Z{Member: comment.CId, Score: float64(comment.Sore)}
		zSetList = append(zSetList, sortItem)

		cId := strconv.FormatInt(comment.CId, 10)
		//单个评论缓存
		commentKey := fmt.Sprintf(CommentDetail, cId)
		byteData, jsonErr := json.Marshal(&comment)
		if jsonErr != nil {
			return errors.Wrap(jsonErr, "SetChildComment Marshal Fail")
		}
		pipe.Set(ctx, commentKey, byteData, 24*time.Hour)

		// 设置单个评论的计数
		countKey := fmt.Sprintf(CommentCount, cId)
		pipe.HSetNX(ctx, countKey, "Likes", comment.Likes)
		pipe.Expire(ctx, countKey, 3*24*time.Hour)
	}

	//通过评论组不够 20条 说明到了最后一页 在缓存中添加标识
	commentLen := len(rootComments)
	// 因为 sore 是时间戳秒 最大10为 所以给11位数
	if commentLen < 20 {
		sortItem := redis.Z{Member: "end", Score: 10000000001}
		zSetList = append(zSetList, sortItem)
	}

	// 添加到 redis中
	pipe.ZAdd(ctx, zSortKey, zSetList...)
	pipe.Expire(ctx, zSortKey, 24*time.Hour*7)
	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "SetChildComment Cache Fail")
		return
	}

	return
}

// GetChildComment 获取子评论的缓存
func GetChildComment(rootId string, sore string, count int) (res []redis.Cmder, isEnd bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zSortKey := fmt.Sprintf(CommentChildList, rootId)
	opt := &redis.ZRangeBy{Max: "+inf", Min: sore, Offset: 0, Count: int64(count)}
	cidList, err := Rdb.ZRangeByScore(ctx, zSortKey, opt).Result()
	if err != nil {
		err = errors.Wrap(err, "GetChildComment ZRangeBy Cache Fail")
		return
	}

	pipe := Rdb.Pipeline()
	for i, cid := range cidList {
		// 说明没有评论
		if cid == "end" && i == 0 {
			isEnd = true
			break
		}
		// 最后一个是 end 不要查询
		if cid == "end" {
			isEnd = true
			break
		}
		//如果不是第一页 第一条都不要 因为是重复的
		if sore != "-inf" && i == 0 {
			continue
		}
		//单个评论缓存
		commentKey := fmt.Sprintf(CommentDetail, cid)
		pipe.Get(ctx, commentKey)
	}

	// 没有评论的情况
	if isEnd && len(cidList) == 0 {
		return
	}

	res, err = pipe.Exec(ctx)
	if err != nil {
		// 如果返回的错误是key不存在 说明有评论主题的缓存丢失 到数据库查询
		if errors.Is(err, redis.Nil) {
			return
		}
		// 出其他错了
		err = errors.Wrap(err, "GetRootCommentId Get Cache Fail")
		return
	}

	return
}

// GetOneComment 获取一条评论缓存
func GetOneComment(cid string) (res string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var key string
	key = fmt.Sprintf(CommentDetail, cid)

	res, err = Rdb.Get(ctx, key).Result()
	if err != nil {
		err = errors.Wrap(err, "GetOneComment Cache Fail")
		return
	}
	return
}

// SetOneCommentCache 设置一个评论缓存
func SetOneCommentCache(comment m.Comment) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cId := strconv.FormatInt(comment.CId, 10)

	pipe := Rdb.Pipeline()
	// 统计缓存
	countKey := fmt.Sprintf(CommentCount, cId)
	pipe.HSetNX(ctx, countKey, "Likes", comment.Likes)
	pipe.Expire(ctx, countKey, 3*24*time.Hour)

	//信息缓存
	comment.IsLike = false // 重置是否点赞
	var commentKey string

	commentKey = fmt.Sprintf(CommentDetail, cId)
	byteData, jsonErr := json.Marshal(&comment)
	if jsonErr != nil {
		return errors.Wrap(jsonErr, "SetOneCommentCache Marshal Fail")
	}
	pipe.Set(ctx, commentKey, byteData, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, "SetOneCommentCache fail")
	}
	return
}

// SetCommentLike 设置评论点赞相关缓存
func SetCommentLike(userId string, likeData m.PostCommentLike) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	// 评论统计缓存 key
	cKey := fmt.Sprintf(CommentCount, likeData.CId)
	// 给点赞者 点赞列表添加/删除
	uSortKey := fmt.Sprintf(UserLike, userId)

	if !likeData.IsCancel {
		pipe.HIncrBy(ctx, cKey, "Likes", 1)
		pipe.ZAdd(ctx, uSortKey, redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: likeData.CId,
		})
	} else {
		pipe.HIncrBy(ctx, cKey, "Likes", -1)
		pipe.ZRem(ctx, uSortKey, likeData.CId)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		err = errors.Wrap(err, "SetCommentLike Cache Fail")
		return err
	}
	return
}

// GetCommentLikeCount 获取评论点赞
func GetCommentLikeCount(cIds []string) (data map[int64]map[string]string, err error) {
	var keys []string
	for _, id := range cIds {
		keys = append(keys, fmt.Sprintf(CommentCount, id))
	}
	data, _, err = BatchGetTypeOfHash(keys)
	if err != nil {
		err = errors.Wrap(err, "GetCommentLike fail")
	}
	return
}

// DelCommentCache 删除一个评论
func DelCommentCache(comment m.Comment) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pipe := Rdb.Pipeline()

	cId := strconv.FormatInt(comment.CId, 10)

	var zSetKey string
	if comment.RootId == 0 {
		zSetKey = fmt.Sprintf(CommentRootList, comment.OwnId)
	} else {
		rootId := strconv.FormatInt(comment.RootId, 10)
		zSetKey = fmt.Sprintf(CommentChildList, rootId)
		// 子回复可能显示在 子评论两条预览里面 所以 删子回复还要删根回复的详情
		rKey := fmt.Sprintf(CommentDetail, rootId)
		pipe.Del(ctx, rKey)
	}

	detailKey := fmt.Sprintf(CommentDetail, cId)

	pipe.Del(ctx, detailKey)     //评论详情删除
	pipe.ZRem(ctx, zSetKey, cId) // 评论列表删除

	var countKey string
	if comment.OwnType == "aw" {
		countKey = fmt.Sprintf(ArtworkCount, comment.OwnId)
	} else {
		countKey = fmt.Sprintf(TrendCount, comment.OwnId)
	}
	incr := comment.RootCount + 1 // 减去的评论数 包括根回复
	pipe.HIncrBy(ctx, countKey, "Comments", int64(-incr))

	_, err = pipe.Exec(ctx)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("DelCommentCache fail %s", cId))
	}
	return
}
