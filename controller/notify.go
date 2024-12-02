package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
	"time"
)

// GetNotifyUnread 获取未读通知数
func GetNotifyUnread(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	userData, _ := ctx.Get("userInfo")
	userInfo, _ := userData.(m.UserTokenPayload)

	countData, err := mysql.GetNotifyUnreadCount(userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, countData)

}

// SetLikeOrCollectNotify 设置喜欢通知
func SetLikeOrCollectNotify(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	userData, _ := ctx.Get("userInfo")
	data, _ := ctx.Get("interact")

	interactData, _ := data.(*m.PostInteractData)
	userInfo, _ := userData.(m.UserTokenPayload)
	ctx.Set("userId", interactData.AuthorId)

	//	如果是取消点赞 或是自己点赞自己收藏自己
	if interactData.IsCancel || interactData.AuthorId == userInfo.Id {
		ctx.Abort()
		return
	}

	// 查找需要通知人的通知配置
	config, haveCache, err := GetUserNotifyConfig(interactData.AuthorId)
	ctx.Set("config", config)
	if err != nil {
		logger.ErrZapLog(err, "SetLikeOrCollectNotify GetUserNotifyConfig fail")
		ctx.Abort()
		return
	}
	if haveCache {
		ctx.Abort()
	}

	var notifyConfig uint8
	if interactData.Action == "like" {
		notifyConfig = config.Like
	} else {
		notifyConfig = config.Collect
	}

	// 设置了不通知 不做操作
	if notifyConfig == 0 {
		return
	}

	//仅关注的人 查询是否关注
	if notifyConfig == 2 {
		isFocus, _err := cache.CheckUserFollow([]string{userInfo.Id}, interactData.AuthorId)
		if _err != nil {
			logger.ErrZapLog(_err, "SetLikeOrCollectNotify CheckUserFollow fail")
		}
		// 说明没有关注
		if isFocus[userInfo.Id] == 0 || len(isFocus) == 0 {
			return
		}
	}

	notify := m.NotifyBody{
		BaseNotify: m.BaseNotify{
			Type:       "remind",
			TargetId:   interactData.MsgId,
			TargetType: interactData.Type,
			Action:     interactData.Action,
			Sender:     m.UserSimpleInfo{UserId: userInfo.Id},
			ReceiverId: interactData.AuthorId,
		},
		Content: nil,
	}
	isNew, err := mongo.SendNotify(notify)
	if err != nil {
		logger.ErrZapLog(err, interactData)
	}

	// 点赞又取消再点赞 不算新增的数据 不添加计数
	if !isNew {
		return
	}
	err = mysql.SetNotifyUnread(interactData.AuthorId, interactData.Action)
	if err != nil {
		logger.ErrZapLog(err, interactData)
	}
}

// SetLikeCommentNotify 设置点赞评论通知
func SetLikeCommentNotify(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("likeData")
	likeData, _ := ctxData.(m.PostCommentLike)

	//	如果是取消点赞 或是自己点赞自己
	if likeData.IsCancel || likeData.AuthorId == userInfo.Id {
		ctx.Abort()
		return
	}

	// 查找需要通知人的通知配置
	config, haveCache, err := GetUserNotifyConfig(likeData.AuthorId)
	if err != nil {
		logger.ErrZapLog(err, "SetLikeCommentNotify GetUserNotifyConfig fail")
		ctx.Abort()
		return
	}
	if haveCache {
		ctx.Abort()
	}

	// 设置了不通知 不做操作
	if config.Like == 0 {
		return
	}

	//仅关注的人 查询是否关注
	if config.Like == 2 {
		isFocus, _err := cache.CheckUserFollow([]string{userInfo.Id}, likeData.AuthorId)
		if _err != nil {
			logger.ErrZapLog(_err, "SetLikeCommentNotify CheckUserFollow fail")
		}
		// 说明没有关注
		if isFocus[userInfo.Id] == 0 || len(isFocus) == 0 {
			return
		}
	}
	likeCId, _ := strconv.ParseInt(likeData.CId, 10, 64)
	content := m.LikeCommentNotify{
		BeLikeCId: likeCId,
	}

	notify := m.NotifyBody{
		BaseNotify: m.BaseNotify{
			Type:       "remind",
			TargetId:   likeData.CId,
			TargetType: "cm",
			Action:     "like",
			Sender:     m.UserSimpleInfo{UserId: userInfo.Id},
			ReceiverId: likeData.AuthorId,
			UpdateAt:   time.Now(),
		},
		Content: content,
	}

	isNew, err := mongo.SendNotify(notify)
	if err != nil {
		logger.ErrZapLog(err, likeData)
	}

	// 点赞又取消再点赞 不算新增的数据 不添加计数
	if !isNew {
		return
	}
	err = mysql.SetNotifyUnread(likeData.AuthorId, "like")
	if err != nil {
		logger.ErrZapLog(err, likeData)
	}

	ctx.Set("config", config)
	ctx.Set("userId", likeData.AuthorId)
}

// GetLikeAndCollectNotify 获取点赞收藏的提醒
func GetLikeAndCollectNotify(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)
	ctxData, _ = ctx.Get("query")
	queryData, _ := ctxData.(m.NotifyQuery)

	notify, err := mongo.GetLikeAndCollectNotify(userInfo.Id, queryData.NextId)
	if err != nil {
		err = errors.Wrap(err, "GetLikeAndCollectNotify fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	if notify == nil {
		ResponseSuccess(ctx, make([]struct{}, 0))
		ctx.Abort()
		return
	}

	var uIds []string   // 查找用户信息
	var findAw []string // 查找点赞或收藏的作品
	var findTr []int64  //  查找点赞或收藏的动态
	var findCm []int64
	for _, n := range notify {
		uIds = append(uIds, n.Sender.UserId)
		if n.TargetType == "aw" {
			findAw = append(findAw, n.TargetId)
		} else if n.TargetType == "tr" {
			i, _ := strconv.ParseInt(n.TargetId, 10, 64)
			findTr = append(findTr, i)
		} else {
			content := n.Content.(m.LikeCommentNotify)
			findCm = append(findCm, content.BeLikeCId)
		}
	}
	// 查找评论
	commentMap, err := mongo.BatchGetComment(findCm)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	for _, comment := range commentMap {
		if comment.OwnType == "aw" {
			findAw = append(findAw, comment.OwnId)
		} else {
			i, _ := strconv.ParseInt(comment.OwnId, 10, 64)
			findTr = append(findTr, i)
		}
	}

	userMap, err := mysql.GetBatchUserSimpleInfo(uIds)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	artMap, trendMap, findData, err := BatchGetNotifyFactorInfo([]string{userInfo.Id}, findAw, findTr)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	for i, body := range notify {
		notify[i].Sender = userMap[body.Sender.UserId]
		if body.TargetType == "aw" {
			notify[i].Content = artMap[body.TargetId]
		} else if body.TargetType == "tr" {
			notify[i].Content = trendMap[body.TargetId]
		} else {
			content := notify[i].Content.(m.LikeCommentNotify)
			likeCid := content.BeLikeCId
			content.OwnId = commentMap[likeCid].OwnId
			content.OwnType = commentMap[likeCid].OwnType
			content.CommentIsDel = commentMap[likeCid].IsDelete
			content.RootId = commentMap[likeCid].RootId
			if content.CommentIsDel {
				content.BeLikeContent = "「该评论已删除」"
			} else {
				content.BeLikeContent = commentMap[likeCid].Text
			}
			if content.OwnType == "aw" {
				content.Author = artMap[content.OwnId].Author
				content.Cover = artMap[content.OwnId].Cover
				content.OwnerIsDel = artMap[content.OwnId].IsDelete
			} else {
				content.Author = trendMap[content.OwnId].Author
				content.Cover = trendMap[content.OwnId].Cover
				content.OwnerIsDel = trendMap[content.OwnId].IsDelete
			}
			notify[i].Content = content
		}
	}

	ResponseSuccess(ctx, notify)
	// 清除未读
	if queryData.NextId == "0" {
		err = mysql.AckNotifyUnread(userInfo.Id, "likeAndCollect")
		if err != nil {
			logger.ErrZapLog(err, "AckNotifyUnread likeAndCollect fail")
		}
	}

	//需要缓存的数据
	ctx.Set("artData", findData)
}

// GetFocusNotify 获取关注通知
func GetFocusNotify(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)
	ctxData, _ = ctx.Get("query")
	queryData, _ := ctxData.(m.NotifyQuery)

	notify, err := mongo.GetFocusNotify(userInfo.Id, queryData.NextId)
	if err != nil {
		err = errors.Wrap(err, "GetFocusNotify fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	if notify == nil {
		ResponseSuccess(ctx, make([]struct{}, 0))
		return
	}

	var uIds []string // 查找用户信息
	for _, n := range notify {
		uIds = append(uIds, n.Sender.UserId)
	}
	userMap, err := mysql.GetBatchUserSimpleInfo(uIds)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	focusIdMap, err := cache.CheckUserFollow(uIds, userInfo.Id)

	for i, body := range notify {
		notify[i].Sender = userMap[body.Sender.UserId]
		isFocus := focusIdMap[body.Sender.UserId]
		notify[i].Content = m.UserIsFocus{
			UserId:  body.Sender.UserId,
			IsFocus: isFocus,
		}
	}

	ResponseSuccess(ctx, notify)
	// 清除未读
	if queryData.NextId == "0" {
		err = mysql.AckNotifyUnread(userInfo.Id, "focus")
		if err != nil {
			logger.ErrZapLog(err, "AckNotifyUnread likeAndCollect fail")
		}
	}
}

// SetCommentNotify 设置评论通知
func SetCommentNotify(ctx *gin.Context) {
	//// 1. 获取传递的数据
	dataCtx, _ := ctx.Get("newComment")
	newComment := dataCtx.(m.ReturnComment)

	userData, _ := ctx.Get("userInfo")
	userInfo, _ := userData.(m.UserTokenPayload)

	dataCtx, _ = ctx.Get("comment")
	postComment := dataCtx.(m.PostCommentData)

	// 自己评论自己 不提醒
	if newComment.ReplyUserId == userInfo.Id {
		ctx.Abort()
		return
	}

	// 查找需要通知人的通知配置
	config, haveCache, err := GetUserNotifyConfig(newComment.ReplyUserId)
	if err != nil {
		logger.ErrZapLog(err, "SetLikeOrCollectNotify GetUserNotifyConfig fail")
		ctx.Abort()
		return
	}
	if haveCache {
		ctx.Abort()
	}

	// 设置了不通知 不做操作
	if config.Comment == 0 {
		return
	}

	//仅关注的人 查询是否关注
	if config.Comment == 2 {
		isFocus, _err := cache.CheckUserFollow([]string{userInfo.Id}, newComment.ReplyUserId)
		if _err != nil {
			logger.ErrZapLog(_err, "SetLikeOrCollectNotify CheckUserFollow fail")
		}
		// 说明没有关注
		if isFocus[userInfo.Id] == 0 || len(isFocus) == 0 {
			return
		}
	}

	var targetId string
	var targetType string
	content := m.NotifyCommentInfo{
		SendCId: newComment.CId,
	}
	if newComment.RootId == 0 {
		// 根回复 -> 评论了你的作品/动态
		targetId = newComment.OwnId
		targetType = postComment.Type
	} else {
		// 子回复 -> 回复了你的评论
		targetId = strconv.FormatInt(newComment.RootId, 10)
		targetType = "cm"
		if newComment.ReplyId == 0 {
			// 回复的是根回复
			content.BeReplyCId = newComment.RootId
		} else {
			//回复的是子回复论
			content.BeReplyCId = newComment.ReplyId
		}

	}

	notify := m.NotifyBody{
		BaseNotify: m.BaseNotify{
			Type:       "remind",
			TargetId:   targetId,
			TargetType: targetType,
			Action:     "comment",
			Sender:     m.UserSimpleInfo{UserId: userInfo.Id},
			ReceiverId: newComment.ReplyUserId,
			UpdateAt:   time.Now(),
		},
		Content: content,
	}
	err = mongo.SendRepetitionNotify(notify)
	if err != nil {
		logger.ErrZapLog(err, notify)
	}

	err = mysql.SetNotifyUnread(newComment.ReplyUserId, "comment")
	if err != nil {
		logger.ErrZapLog(err, notify)
	}

	ctx.Set("config", config)
	ctx.Set("userId", newComment.ReplyUserId)
}

// GetCommentNotify 获取评论通知
func GetCommentNotify(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("query")
	queryData, _ := ctxData.(m.NotifyQuery)

	notify, err := mongo.GetCommentNotify(userInfo.Id, queryData.NextId)
	if err != nil {
		err = errors.Wrap(err, "GetFocusNotify fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	if notify == nil {
		ResponseSuccess(ctx, make([]struct{}, 0))
		ctx.Abort()
		return
	}

	var uIds []string   // 查找用户信息
	var findAw []string // 查找点赞或收藏的作品
	var findTr []int64  //  查找点赞或收藏的动态
	var checkLikecIds []string
	var allCId []int64
	for _, n := range notify {
		uIds = append(uIds, n.Sender.UserId)
		checkLikecIds = append(checkLikecIds, strconv.FormatInt(n.Content.SendCId, 10))
		allCId = append(allCId, n.Content.SendCId, n.Content.BeReplyCId)
	}

	commentMap, err := mongo.BatchGetComment(allCId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	for _, comment := range commentMap {
		if comment.OwnType == "aw" {
			findAw = append(findAw, comment.OwnId)
		} else {
			i, _ := strconv.ParseInt(comment.OwnId, 10, 64)
			findTr = append(findTr, i)
		}
	}

	userMap, err := mysql.GetBatchUserSimpleInfo(uIds)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	artMap, trendMap, findData, err := BatchGetNotifyFactorInfo([]string{userInfo.Id}, findAw, findTr)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	likeMap, err := cache.CheckUserLike(checkLikecIds, userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	for i, body := range notify {
		notify[i].Sender = userMap[body.Sender.UserId]
		sendComment := commentMap[body.Content.SendCId]
		beReplyComment := commentMap[body.Content.BeReplyCId]

		notify[i].Content.OwnType = sendComment.OwnType
		notify[i].Content.OwnId = sendComment.OwnId
		notify[i].Content.SendIsDel = sendComment.IsDelete
		if sendComment.IsDelete {
			notify[i].Content.SendContent = "「该评论已删除」"
		} else {
			notify[i].Content.SendContent = sendComment.Text
		}
		if beReplyComment.IsDelete {
			notify[i].Content.BeReplyContent = "「该评论已删除」"
		} else {
			notify[i].Content.BeReplyContent = beReplyComment.Text
		}

		if isLike, ok := likeMap[strconv.FormatInt(body.Content.SendCId, 10)]; ok {
			notify[i].Content.IsLike = isLike
		}
		if notify[i].Content.OwnType == "aw" {
			if artInfo, ok := artMap[sendComment.OwnId]; ok {
				notify[i].Content.Cover = artInfo.Cover
				notify[i].Content.Author = artInfo.Author
				notify[i].Content.OwnerIsDel = artInfo.IsDelete
			}
		} else {
			if trendInfo, ok := trendMap[sendComment.OwnId]; ok {
				notify[i].Content.Cover = trendInfo.Cover
				notify[i].Content.Author = trendInfo.Author
				notify[i].Content.OwnerIsDel = trendInfo.IsDelete
			}
		}
	}

	ResponseSuccess(ctx, notify)
	// 清除未读
	if queryData.NextId == "0" {
		err = mysql.AckNotifyUnread(userInfo.Id, "comment")
		if err != nil {
			logger.ErrZapLog(err, "AckNotifyUnread likeAndCollect fail")
		}
	}

	//需要缓存的数据
	ctx.Set("artData", findData)
}

// GetNotifySetting 获取通知设置
func GetNotifySetting(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	config, haveCache, err := GetUserNotifyConfig(userInfo.Id)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	if haveCache {
		ctx.Abort()
	}
	ResponseSuccess(ctx, config)
	ctx.Set("userId", userInfo.Id)
	ctx.Set("config", config)
}

// UpdateNotifySetting 更新通知配置
func UpdateNotifySetting(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("config")
	config, _ := ctxData.(m.NotifyConfig)

	err := mysql.SettingNotifySetting(userInfo.Id, config)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, config)
	ctx.Set("userId", userInfo.Id)
}

// SetFocusNotify 设置关注通知
func SetFocusNotify(ctx *gin.Context) {
	// 获取传递的参数
	ctxData, _ := ctx.Get("Focus")
	focusInfo := ctxData.(m.VerifyUserFocus)

	ctxData, _ = ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	// 取消关注 不提醒
	if focusInfo.IsCancel {
		err := mongo.DelFocusNotify(focusInfo.FocusId, userInfo.Id)
		if err != nil {
			logger.ErrZapLog(err, "DelFocusNotify fail")
		}
		ctx.Abort()
		return
	}

	// 查找需要通知人的通知配置
	config, haveCache, err := GetUserNotifyConfig(focusInfo.FocusId)
	if err != nil {
		logger.ErrZapLog(err, "SetLikeOrCollectNotify GetUserNotifyConfig fail")
		ctx.Abort()
		return
	}
	if haveCache {
		ctx.Abort()
	}

	// 设置了不通知 不做操作
	if config.Follow == 0 {
		return
	}

	//仅关注的人 查询是否关注
	if config.Follow == 2 {
		isFocus, _err := cache.CheckUserFollow([]string{userInfo.Id}, focusInfo.FocusId)
		if _err != nil {
			logger.ErrZapLog(_err, "SetLikeOrCollectNotify CheckUserFollow fail")
		}
		// 说明没有关注
		if isFocus[userInfo.Id] == 0 || len(isFocus) == 0 {
			return
		}
	}

	// 设置关注提醒
	notify := m.NotifyBody{
		BaseNotify: m.BaseNotify{
			Type:       "remind",
			TargetId:   focusInfo.FocusId,
			TargetType: "usr",
			Action:     "focus",
			Sender:     m.UserSimpleInfo{UserId: userInfo.Id},
			ReceiverId: focusInfo.FocusId,
		},
		Content: nil,
	}
	isNew, err := mongo.SendNotify(notify)
	if err != nil {
		logger.ErrZapLog(err, notify)
	}

	// 点赞又取消再点赞 不算新增的数据 不添加计数
	if !isNew {
		return
	}
	err = mysql.SetNotifyUnread(focusInfo.FocusId, "focus")
	if err != nil {
		logger.ErrZapLog(err, notify)
	}

	ctx.Set("config", config)
	ctx.Set("userId", focusInfo.FocusId)
}

// SetCommissionNotify 设置约稿通知
func SetCommissionNotify(ctx *gin.Context) {
	ctxData, ok := ctx.Get("planNext")
	// 如果没有传则不需要设置通知
	if !ok {
		return
	}
	planNext := ctxData.(m.PlanNext)

	ctxData, _ = ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("planUser")
	planUser := ctxData.(m.PlanUserInfo)

	var receiver string
	if loginUser.Id == planUser.ArtistId {
		receiver = planUser.Sender
	} else {
		receiver = planUser.ArtistId
	}

	// 设置关注提醒
	notify := m.NotifyBody{
		BaseNotify: m.BaseNotify{
			Type:       "remind",
			TargetId:   strconv.FormatInt(planNext.InviteId, 10),
			TargetType: "com",
			Action:     "update", //修改了方案状态
			Sender:     m.UserSimpleInfo{UserId: loginUser.Id},
			ReceiverId: receiver,
			UpdateAt:   time.Now(),
		},
		Content: m.NotifyCommissionInfo{Status: planNext.Status}, // 修改的方案状态
	}

	err := mongo.SendRepetitionNotify(notify)
	if err != nil {
		logger.ErrZapLog(err, notify)
	}

	err = mysql.SetNotifyUnread(receiver, "commission")
	if err != nil {
		logger.ErrZapLog(err, notify)
	}
}

// GetCommission 获取约稿的通知
func GetCommission(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	userInfo, _ := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("query")
	queryData, _ := ctxData.(m.NotifyQuery)

	notify, err := mongo.GetCommissionNotify(userInfo.Id, queryData.NextId)
	if err != nil {
		err = errors.Wrap(err, "GetFocusNotify fail")
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	if notify == nil {
		ResponseSuccess(ctx, make([]struct{}, 0))
		return
	}

	var uIds []string   // 查找用户信息
	var invites []int64 // 查找约稿信息
	for _, n := range notify {
		uIds = append(uIds, n.Sender.UserId)
		id, _ := strconv.ParseInt(n.TargetId, 10, 64)
		invites = append(invites, id)
	}

	commissionMap, err := mongo.BatchGetCommissionNotifyInfo(invites)

	userMap, err := mysql.GetBatchUserSimpleInfo(uIds)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	for i, n := range notify {
		notify[i].Sender = userMap[n.Sender.UserId]
		notify[i].Content = commissionMap[n.TargetId]
		notify[i].Content.Status = n.Content.Status
	}

	ResponseSuccess(ctx, notify)

	// 清除未读
	if queryData.NextId == "0" {
		err = mysql.AckNotifyUnread(userInfo.Id, "commission")
		if err != nil {
			logger.ErrZapLog(err, "AckNotifyUnread commission fail")
		}
	}

}
