package handleMiddle

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	m "onpaper-api-go/models"
	"unicode/utf8"
)

// HandleSendMsg 验证发送的私信消息
func HandleSendMsg(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	var data m.SendMessage
	// 取出上传的 json data 信息到 key
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		//绑定失败 参数错误
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}
	// 字数限制
	if utf8.RuneCountInString(data.Content) > 450 {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	// 用登录的 换掉上传的 省去验证
	data.Sender = loginUser.Id
	ctx.Set("message", data)
}

// HandleGetChatList 获取会话列表
func HandleGetChatList(ctx *gin.Context) {
	var data m.VerifyNextId
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
	}

	ctx.Set("nextId", data.NextId)
}

// HandleGetChatRecord 获取聊天记录需要参数
func HandleGetChatRecord(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	var data m.ReceiveMessage
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
	}
	// 用登录的 换掉上传的 省去验证
	data.Sender = loginUser.Id
	ctx.Set("queryData", data)
}
