package cacheMiddle

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	c "onpaper-api-go/cache"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
)

// SetRootComment 设置作品评论缓存
func SetRootComment(ctx *gin.Context) {
	// 1. 获取传递的数据
	dataCtx, _ := ctx.Get("queryData")
	queryData := dataCtx.(m.VerifyGetCommentQuery)

	dataCtx, _ = ctx.Get("rootComment")
	rootComment := dataCtx.(m.ReturnComments)

	err := c.SetRootComment(rootComment, queryData.OwnId, queryData.Type)
	if err != nil {
		logger.ErrZapLog(err, queryData.OwnId)
	}
}

// GetRootComment 获取作品评论缓存
// tip: 评论缓存有个bug 因为是zSore 添加，正常情况下先加第一页缓存加第二页缓存 查询的时候 按时间排列没问题
// 如果 有爬虫 跳过了第一页 先查询了第n页 第n页添加进缓存 这时候 查询缓存的数据从第n页开始 会造成不准确，缓存无法判断是否最新评论开始
func GetRootComment(ctx *gin.Context) {
	var data m.VerifyGetCommentQuery
	_ = ctx.ShouldBindQuery(&data)

	count := 21
	// 如果 lastCid = 0 说明是第一页
	if data.LastCid == "0" {
		data.Sore = "+inf"
		count = 20
	}

	var cacheValue c.CtxCacheVale
	var tempList m.ReturnComments

	commentList, isEnd, err := c.GetRootCommentId(data.OwnId, data.Sore, count)
	if err != nil {
		//可能有缓存丢失 交给数据库查询
		if errors.Is(err, redis.Nil) {
			ctx.Set("comment", cacheValue)
			ctx.Set("noComment", false)
			return
		}
		logger.ErrZapLog(err, data.OwnId)
	}

	// 说明有缓存 没有评论
	if isEnd && len(commentList) == 0 {
		ctx.Set("comment", cacheValue)
		ctx.Set("noComment", true)
		return
	}
	// 如果查询的少于20条 又不是到头 说明 查询到末尾 又有新评论添加 挤出了部分数据
	// 例子： [3,2,1] -> [4,3,2] 维护 队列长度为3个  后面的 1 被删了 查询3-1 少了一个
	// 没有缓存也会来到这！！！
	if !isEnd && len(commentList) < 20 {
		ctx.Set("comment", cacheValue)
		ctx.Set("noComment", false)
		return
	}

	for _, cmder := range commentList {
		var temp m.ReturnComment
		val := c.GetCmderStringResult(cmder)
		err = json.Unmarshal([]byte(val), &temp)
		if err != nil {
			err = errors.Wrap(err, "GetRootComment Unmarshal fail ")
			logger.ErrZapLog(err, val)
			ctx.Set("comment", cacheValue)
			ctx.Set("noComment", false)
			return
		} else {
			tempList = append(tempList, temp)
		}
	}
	cacheValue.Val = tempList
	cacheValue.HaveCache = true
	ctx.Set("comment", cacheValue)
	ctx.Set("noComment", false)
}

// AddComment 添加新的评论到缓存中
func AddComment(ctx *gin.Context) {
	// 1. 获取传递的数据
	dataCtx, _ := ctx.Get("newComment")
	newComment := dataCtx.(m.ReturnComment)
	dataCtx, _ = ctx.Get("comment")
	postComment := dataCtx.(m.PostCommentData)

	err := c.AddOneComment(newComment, postComment.Type)
	if err != nil {
		logger.ErrZapLog(err, newComment.CId)
	}
}

// SetChildrenComment 设置子评论的缓存
func SetChildrenComment(ctx *gin.Context) {
	//获取传递的参数
	data, _ := ctx.Get("queryData")
	queryData := data.(m.VerifyGetCommentReply)

	data, _ = ctx.Get("childComment")
	comments := data.(m.Comments)

	err := c.SetChildComment(comments, queryData.RootId)
	if err != nil {
		logger.ErrZapLog(err, queryData.RootId)
	}
}

// GetChildrenComment 获取子评论的缓存
func GetChildrenComment(ctx *gin.Context) {
	//获取传递的参数
	data, _ := ctx.Get("queryData")
	queryData := data.(m.VerifyGetCommentReply)

	count := 21
	// 如果 lastCid = 0 说明是第一页
	if queryData.LastCid == "0" {
		queryData.Sore = "-inf"
		count = 20
	}
	var cacheValue c.CtxCacheVale
	var tempList m.Comments

	commentList, isEnd, err := c.GetChildComment(queryData.RootId, queryData.Sore, count)
	if err != nil {
		//可能有缓存丢失 交给数据库查询
		if errors.Is(err, redis.Nil) {
			ctx.Set("comment", cacheValue)
			ctx.Set("noComment", false)
			return
		}
		logger.ErrZapLog(err, queryData.RootId)
	}

	// 说明有缓存 没有评论
	if isEnd && len(commentList) == 0 {
		ctx.Set("comment", cacheValue)
		ctx.Set("noComment", true)
		return
	}

	// 如果查询的少于20条 又不是到头 说明 查询到末尾 又有新评论添加 挤出了部分数据
	// 例子： [3,2,1] -> [4,3,2] 维护 队列长度为3个  后面的 1 被删了 查询3-1 少了一个
	// 没有缓存也会来到这
	if !isEnd && len(commentList) < 20 {
		ctx.Set("comment", cacheValue)
		ctx.Set("noComment", false)
		return
	}

	for _, cmder := range commentList {
		var temp m.Comment
		val := c.GetCmderStringResult(cmder)
		err = json.Unmarshal([]byte(val), &temp)
		if err != nil {
			err = errors.Wrap(err, "GetRootComment Unmarshal fail ")
			logger.ErrZapLog(err, val)
			ctx.Set("comment", cacheValue)
			ctx.Set("noComment", false)
			return
		} else {
			tempList = append(tempList, temp)
		}
	}
	cacheValue.Val = tempList
	cacheValue.HaveCache = true
	ctx.Set("comment", cacheValue)
	ctx.Set("noComment", false)
}
