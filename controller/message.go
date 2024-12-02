package controller

import (
	"github.com/gin-gonic/gin"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/snowflake"
	"time"
)

// SaveMsg 保存消息
func SaveMsg(ctx *gin.Context) {
	//1.获取传递的 作品信息
	ctxData, _ := ctx.Get("message")
	sendMsg := ctxData.(m.SendMessage)

	chatId, isExist, err := mongo.FindChatId(sendMsg.Sender, sendMsg.Receiver)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	if !isExist {
		// 如果没有id 说明是新的会话 生成新的id
		chatId = snowflake.CreateID()
	}

	saveMsg := m.MessageBody{
		ChatId:   chatId,
		MsgId:    snowflake.CreateID(),
		Sender:   sendMsg.Sender,
		Receiver: sendMsg.Receiver,
		Content:  sendMsg.Content,
		MsgType:  sendMsg.MsgType,
		SendTime: time.Now(),
		Width:    sendMsg.Width,
		Height:   sendMsg.Height,
	}

	// 优先保存消息
	err = mongo.SaveMsg(saveMsg)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 建立会话关系
	err = mongo.SetChatRelation(saveMsg, isExist)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	ResponseSuccess(ctx, saveMsg)

}

// GetChatList 获取会话列表
func GetChatList(ctx *gin.Context) {
	// 取出 ctx 传递的数据
	ctxData, _ := ctx.Get("userInfo")
	userInfo := ctxData.(m.UserTokenPayload)

	ctxData, _ = ctx.Get("nextId")
	msgId := ctxData.(*int64)
	chatList, err := mongo.GetChatList(userInfo.Id, *msgId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	var findUser []string
	for _, data := range chatList {
		findUser = append(findUser, data.SenderId)
		findUser = append(findUser, data.ReceiverId)
	}
	// 查询名字和头像
	userMap, err := mysql.GetBatchUserSimpleInfo(findUser)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}
	// 拼接数据
	for i, relation := range chatList {
		chatList[i].Receiver = userMap[relation.ReceiverId]
		chatList[i].Sender = userMap[relation.SenderId]
	}

	ResponseSuccess(ctx, chatList)

}

// GetChatRecord 获取聊天记录
func GetChatRecord(ctx *gin.Context) {
	ctxData, _ := ctx.Get("queryData")
	queryData := ctxData.(m.ReceiveMessage)

	code := CodeSuccess        //返回的状态码 默认返回 200
	configHaveCache := true    // 权限配置是否有缓存
	config := m.NotifyConfig{} // 私信权限配置

	chatId := *queryData.ChatId
	// 如果没有传 chatI d 查找会话id （第一次打开会话框会来到这）
	if chatId == 0 {
		var err error
		// 查找私信权限
		config, configHaveCache, err = GetUserNotifyConfig(queryData.Receiver)
		if err != nil {
			logger.ErrZapLog(err, "SetLikeOrCollectNotify GetUserNotifyConfig fail")
			ctx.Abort()
			return
		}

		// 设置了不许发送私信
		if config.Message == 0 {
			code = CodeUnPermission
		}

		//仅关注的人 查询是否关注
		if config.Message == 2 {
			isFocus, _err := cache.CheckUserFollow([]string{queryData.Sender}, queryData.Receiver)
			if _err != nil {
				logger.ErrZapLog(_err, "SetLikeOrCollectNotify CheckUserFollow fail")
			}
			// 说明没有关注
			if isFocus[queryData.Sender] == 0 || len(isFocus) == 0 {
				code = CodeOnlyHeFocusUserCanDo
			}
		}

		var isExist bool
		chatId, isExist, err = mongo.FindChatId(queryData.Sender, queryData.Receiver)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
		// 没有会话id 说明之前没有联系过 直接返回空数组
		if !isExist {
			Response(ctx, code, []struct{}{})
			// 如果有缓存 不再设置
			if configHaveCache {
				ctx.Abort()
			} else {
				ctx.Set("userId", queryData.Receiver)
				ctx.Set("config", config)
			}
			return
		}
	}

	msg, err := mongo.GetChatRecord(chatId, *queryData.NextId)
	if err != nil {
		ResponseErrorAndLog(ctx, CodeServerBusy, err)
		return
	}

	// 说明查询新消息 把未读改0
	if *queryData.NextId == 0 {
		err = mongo.AckChatUnread(queryData.Sender, queryData.Receiver)
		if err != nil {
			ResponseErrorAndLog(ctx, CodeServerBusy, err)
			return
		}
	}

	Response(ctx, code, msg)
	// 如果有缓存 不再设置
	if configHaveCache {
		ctx.Abort()
	} else {
		ctx.Set("userId", queryData.Receiver)
		ctx.Set("config", config)
	}

}
