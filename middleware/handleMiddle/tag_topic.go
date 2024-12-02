package handleMiddle

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	m "onpaper-api-go/models"
)

// HandleQueryTagArt 验证查询tag对应作品的 参数
func HandleQueryTagArt(ctx *gin.Context) {
	var data m.TagQueryArtParam
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
	}

	ctx.Set("queryData", data)
}

// HandleQueryTag 验证查询 tag相关信息的参数
func HandleQueryTag(ctx *gin.Context) {
	var data m.TagQueryParam
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
	}

	ctx.Set("queryData", data)
}

// HandleQueryTopic 验证查询的Topic
func HandleQueryTopic(ctx *gin.Context) {
	var data m.TopicQueryParam
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
	}

	ctx.Set("queryData", data)
}

// HandleQueryTopicTrend 验证话题相关动态参数
func HandleQueryTopicTrend(ctx *gin.Context) {
	var data m.TopicQueryTrendParam
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
	}

	ctx.Set("queryData", data)
}
