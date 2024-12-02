package handleMiddle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	m "onpaper-api-go/models"
	"strconv"
)

// HandleCommentQuery 处理根评论查询参数
func HandleCommentQuery(ctx *gin.Context) {
	var data m.VerifyGetCommentQuery
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	// 把它传递到上下文
	ctx.Set("queryData", data)
}

// HandleRootCommentReply 获取单个作品中根评论的所有子评论
func HandleRootCommentReply(ctx *gin.Context) {
	var data m.VerifyGetCommentReply
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	//将上传的 string 格式 转成int64比较
	num, err := strconv.ParseInt(data.RootId, 10, 64)
	if num < 10000 {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	// 把它传递到上下文
	ctx.Set("queryData", data)
}

// HandleOneRootComment 查询一条根评论
func HandleOneRootComment(ctx *gin.Context) {
	var queryData m.QueryOneRoot
	err := ctx.ShouldBindQuery(&queryData)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	ctx.Set("query", queryData)
}

// HandleCommentLike 处理评论点赞
func HandleCommentLike(ctx *gin.Context) {
	var data m.PostCommentLike
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		// 如果参数不对
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 把它传递到上下文
	ctx.Set("likeData", data)
}

func HandleCommentDelete(ctx *gin.Context) {
	var data m.DeleteComment
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		// 如果参数不对
		fmt.Println(err)
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	// 把它传递到上下文
	ctx.Set("delete", data)
}
