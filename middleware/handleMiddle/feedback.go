package handleMiddle

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	m "onpaper-api-go/models"
	"unicode/utf8"
)

func HandlePostReport(ctx *gin.Context) {
	var data m.PostReport
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}
	//如果文本长度大于500 则返回错误
	if utf8.RuneCountInString(data.Describe) > 500 {
		ctl.ResponseError(ctx, ctl.CodeJsonFormatError)
		return
	}

	ctx.Set("report", data)
}
