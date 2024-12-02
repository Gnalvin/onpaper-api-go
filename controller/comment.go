package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/snowflake"
	"strconv"
	"time"
)

// GetRootComment 获取作品评论信息
func GetRootComment(ctx *gin.Context) {
	// 1. 获取缓存的数据
	dataCtx, _ := ctx.Get("userInfo")
	loginUser, _ := dataCtx.(m.UserTokenPayload)

	dataCtx, _ = ctx.Get("comment")
	commentCtx := dataCtx.(cache.CtxCacheVale)

	dataCtx, exi := ctx.Get("noComment")
	noComment := dataCtx.(bool)
	// 如果没有评论
	if noComment && exi {
		// 返回数据
		ResponseError(ctx, CodeCommentNoHave)
		return
	}

	var returnComment m.ReturnComments

	if commentCtx.HaveCache {
		// 如果有缓存
		returnComment = commentCtx.Val.(m.ReturnComments)
		ctx.Abort()
	} else {
		//2. 没有缓存到数据库查询
		data, _ := ctx.Get("queryData")
		queryData := data.(m.VerifyGetCommentQuery)

		rootComment, childComment, findUser, err := mongo.GetRootComment(queryData.OwnId, queryData.LastCid)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		// 如果没有评论
		if rootComment == nil {
			//传送给缓存
			ctx.Set("rootComment", make(m.ReturnComments, 0))
			ctx.Next()
			// 返回数据
			ResponseError(ctx, CodeCommentNoHave)
			return
		}

		// 评论区的 用户信息
		userMap, err := mysql.GetBatchUserSimpleInfo(findUser)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		normTime, _ := time.ParseInLocation("2006-01-02", "2005-01-01", time.Local)
		//将评论内容和 用户信息赋值给 查询返回的数据
		for ic, comment := range rootComment {
			// 时间相减为 redis zSet 的 sore 减少数字量级
			rootComment[ic].Sore = comment.CreateAT.Unix() - normTime.Unix()
			rootComment[ic].ChildComments = make([]m.Comment, 0)

			user := userMap[comment.UserId]
			rootComment[ic].UserName = user.UserName
			rootComment[ic].Avatar = user.Avatar
			rootComment[ic].VTag = user.VTag
			rootComment[ic].VStatus = user.VStatus
			if comment.IsDelete {
				rootComment[ic].Text = "「该评论已删除」"
			}
		}

		for _, item := range childComment {
			for ic, comment := range item.Comment {
				user := userMap[comment.UserId]
				// 匹配发评论的用户信息
				item.Comment[ic].UserName = user.UserName
				item.Comment[ic].Avatar = user.Avatar
				//匹配被回复评论的用户信息 把 userID 换成 用户名
				item.Comment[ic].ReplyUserName = userMap[comment.ReplyUserId].UserName
			}
		}

		//将子评论 赋值到 根评论里 形成解构树
		for _, comment := range childComment {
			for ir, root := range rootComment {
				if comment.RootId == root.CId {
					rootComment[ir].ChildComments = comment.Comment
					break
				}
			}
		}

		returnComment = rootComment
	}

	checkId := returnComment.GetAllCId()
	// 检查是否点赞过评论
	likeMap, err := cache.CheckUserLike(checkId, loginUser.Id)
	if err != nil {
		logger.ErrZapLog(err, "GetRootComment CheckUserLike fail ")
	}
	// 查找评论点赞数缓存
	countMap, err := cache.GetCommentLikeCount(checkId)
	if err != nil {
		logger.ErrZapLog(err, "GetRootComment GetCommentLikeCount fail")
	}

	for i, comment := range returnComment {
		cid := strconv.FormatInt(comment.CId, 10)
		if isLike, ok := likeMap[cid]; ok {
			returnComment[i].IsLike = isLike
		}
		if count, ok := countMap[comment.CId]; ok {
			returnComment[i].Likes, _ = strconv.Atoi(count["Likes"])
		}

		for ic, childComment := range comment.ChildComments {
			cid = strconv.FormatInt(childComment.CId, 10)
			if isLike, ok := likeMap[cid]; ok {
				returnComment[i].ChildComments[ic].IsLike = isLike
			}
			if count, ok := countMap[childComment.CId]; ok {
				returnComment[i].ChildComments[ic].Likes, _ = strconv.Atoi(count["Likes"])
			}
		}
	}

	// 返回数据
	ResponseSuccess(ctx, returnComment)
	//传送给缓存
	ctx.Set("rootComment", returnComment)
}

// GetCommentReply 获取作品根评论中的所有子评论
func GetCommentReply(ctx *gin.Context) {
	// 1. 获取缓存的数据
	dataCtx, _ := ctx.Get("userInfo")
	loginUser, _ := dataCtx.(m.UserTokenPayload)

	dataCtx, _ = ctx.Get("comment")
	commentCtx := dataCtx.(cache.CtxCacheVale)

	dataCtx, exi := ctx.Get("noComment")
	noComment := dataCtx.(bool)
	// 如果没有评论
	if noComment && exi {
		// 返回数据
		ResponseError(ctx, CodeCommentNoHave)
		return
	}

	var returnComment m.Comments

	// 如果有缓存 直接缓存返回
	if commentCtx.HaveCache {
		returnComment = commentCtx.Val.(m.Comments)
		ctx.Abort()
	} else {
		//2. 没有缓存到数据库查询
		data, _ := ctx.Get("queryData")
		queryData := data.(m.VerifyGetCommentReply)

		childComment, findUser, err := mongo.GetCommentReply(queryData.RootId, queryData.LastCid)
		if err != nil {
			//数据库错误
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		// 如果没有评论
		if childComment == nil {
			// 返回数据
			ResponseError(ctx, CodeCommentNoHave)
			return
		}

		userMap, err := mysql.GetBatchUserSimpleInfo(findUser)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}

		normTime, _ := time.ParseInLocation("2006-01-02", "2005-01-01", time.Local)
		//循环获取评论内容和用户信息
		for ic, comment := range childComment {
			// 时间相减为 redis zSet 的 sore 减少数字量级
			childComment[ic].Sore = comment.CreateAT.Unix() - normTime.Unix()
			user := userMap[comment.UserId]
			childComment[ic].UserName = user.UserName
			childComment[ic].Avatar = user.Avatar
		}

		returnComment = childComment
	}

	checkId := returnComment.GetAllCId()
	// 检查是否点赞过评论
	likeMap, err := cache.CheckUserLike(checkId, loginUser.Id)
	if err != nil {
		logger.ErrZapLog(err, "GetCommentReply CheckUserLike fail")
	}
	// 查找评论点赞数缓存
	countMap, err := cache.GetCommentLikeCount(checkId)
	if err != nil {
		logger.ErrZapLog(err, "GetCommentReply GetCommentLikeCount fail")
	}

	for i, comment := range returnComment {
		cid := strconv.FormatInt(comment.CId, 10)
		if isLike, ok := likeMap[cid]; ok {
			returnComment[i].IsLike = isLike
		}
		if count, ok := countMap[comment.CId]; ok {
			returnComment[i].Likes, _ = strconv.Atoi(count["Likes"])
		}
	}

	// 返回数据
	ResponseSuccess(ctx, returnComment)
	//传送给缓存
	ctx.Set("childComment", returnComment)
}

// GetOneRootComment 获取一条根评论
func GetOneRootComment(ctx *gin.Context) {
	data, _ := ctx.Get("query")
	queryData := data.(m.QueryOneRoot)

	dataCtx, _ := ctx.Get("userInfo")
	loginUser, _ := dataCtx.(m.UserTokenPayload)

	cacheRes, err := cache.GetOneComment(strconv.FormatInt(queryData.RootId, 10))
	if err != nil && !errors.Is(err, redis.Nil) {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	var rootComment m.Comment
	// 如果有缓存
	if cacheRes != "" {
		err = json.Unmarshal([]byte(cacheRes), &rootComment)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
	} else {
		rootComment, err = mongo.GetOneComment(queryData.RootId)
		if err != nil {
			if err == mongodb.ErrNoDocuments {
				ResponseError(ctx, CodeCommentNoHave)
				return
			}
			err = errors.Wrap(err, "GetOneComment fail")
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		userMap, _err := mysql.GetBatchUserSimpleInfo([]string{rootComment.UserId})
		if _err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}

		userInfo, ok := userMap[rootComment.UserId]
		if ok {
			rootComment.UserName = userInfo.UserName
			rootComment.Avatar = userInfo.Avatar
		}
	}
	// 检查是否点赞过评论
	cid := strconv.FormatInt(rootComment.CId, 10)
	likeMap, err := cache.CheckUserLike([]string{cid}, loginUser.Id)
	if err != nil {
		logger.ErrZapLog(err, "GetOneRootComment CheckUserLike fail")
	}
	// 查找评论点赞数缓存
	countMap, err := cache.GetCommentLikeCount([]string{cid})
	if err != nil {
		logger.ErrZapLog(err, "GetOneRootComment GetCommentLikeCount fail")
	}

	if isLike, ok := likeMap[cid]; ok {
		rootComment.IsLike = isLike
	}
	if count, ok := countMap[rootComment.CId]; ok {
		rootComment.Likes, _ = strconv.Atoi(count["Likes"])
	}
	if rootComment.IsDelete {
		rootComment.Text = "「该评论已删除」"
	}
	ResponseSuccess(ctx, rootComment)
}

// SaveCommentLike 保存评论点赞
func SaveCommentLike(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("likeData")
	likeData, _ := ctxData.(m.PostCommentLike)

	key := fmt.Sprintf(cache.CommentCount, likeData.CId)
	isExists, err := cache.CheckExistsKey(key)
	// 如果缓存不存在 到数据库中查找并设置缓存
	if isExists == 0 || err != nil {
		cid, _err := strconv.ParseInt(likeData.CId, 10, 64)
		if _err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}
		comment, _err := mongo.GetOneComment(cid)
		if _err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}

		_err = cache.SetOneCommentCache(comment)
		if _err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, _err)
			return
		}
	}

	interactData := m.PostInteractData{
		AuthorId: likeData.AuthorId,
		MsgId:    likeData.CId,
		IsCancel: likeData.IsCancel,
		Type:     "cm",
	}
	// MongoDB 保存
	isChange, err := mongo.SetUserLike(userInfo.Id, interactData)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 返回数据
	ResponseSuccess(ctx, likeData)

	// 如果没有实际更新(重复点赞） 不对缓存操作
	if isChange {
		err = cache.SetCommentLike(userInfo.Id, likeData)
		if err != nil {
			logger.ErrZapLog(err, likeData)
		}
		// 如果是添加点赞 维护ZSort
		if !likeData.IsCancel {
			cSortKey := fmt.Sprintf(cache.UserLike, userInfo.Id)
			err = cache.CheckZSortLen(cSortKey, 1000, 100)
			if err != nil {
				logger.ErrZapLog(err, "")
			}
		}
	} else {
		ctx.Abort()
	}
}

// SaveComment 保存评论数据到数据库
func SaveComment(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	userData, _ := ctx.Get("userInfo")
	data, _ := ctx.Get("comment")

	commentData := data.(m.PostCommentData)
	userInfo := userData.(m.UserTokenPayload)

	// 查找发布评论的用户信息
	userMap, err := mysql.GetBatchUserSimpleInfo([]string{userInfo.Id})
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	sender := userMap[userInfo.Id]

	//生成评论雪花id
	cid := snowflake.CreateID()
	err = mongo.SaveOneComment(cid, userInfo.Id, commentData)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	normTime, _ := time.ParseInLocation("2006-01-02", "2005-01-01", time.Local)
	creatAt := time.Now()

	// 返回数据
	var backData m.ReturnComment

	backData.CId = cid
	backData.OwnId = commentData.OwnId
	backData.UserId = userInfo.Id
	backData.Avatar = commentData.SenderAvatar
	backData.UserName = commentData.SenderName
	backData.ReplyId = commentData.ReplyId
	backData.ReplyUserId = commentData.ReplyUserId
	backData.ReplyUserName = commentData.ReplyUserName
	backData.ChildComments = make([]m.Comment, 0)
	backData.RootId = commentData.RootId
	backData.RootCount = 0
	backData.Text = commentData.Text
	backData.Likes = 0
	backData.Sore = creatAt.Unix() - normTime.Unix()
	backData.VStatus = sender.VStatus
	backData.VTag = sender.VTag
	backData.IsDelete = false
	backData.CreateAT = time.Now()

	ResponseSuccess(ctx, backData)
	ctx.Set("newComment", backData)
}

// DelComment 删除一条评论
func DelComment(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("delete")
	delData := ctxData.(m.DeleteComment)

	comment, err := mongo.GetOneComment(delData.CId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	err = mongo.DelOneComment(comment, userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 删除缓存
	err = cache.DelCommentCache(comment)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"status": "ok",
	})
}
