package handleMiddle

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	m "onpaper-api-go/models"
)

// HandleNextIdQuery 验证Feed需要参数
func HandleNextIdQuery(ctx *gin.Context) {
	var data m.VerifyNextId
	err := ctx.ShouldBindQuery(&data)
	if err != nil {
		ctl.ResponseError(ctx, ctl.CodeParamsError)
		return
	}

	ctx.Set("nextId", data.NextId)
}
