package cacheMiddle

import (
	"github.com/gin-gonic/gin"
	c "onpaper-api-go/cache"
	"onpaper-api-go/logger"
	m "onpaper-api-go/models"
	"strconv"
)

// SetActiveData 设置活跃数据
func SetActiveData(ctx *gin.Context) {
	ctxData, _ := ctx.Get("userInfo")
	loginUser := ctxData.(m.UserTokenPayload)

	userId, _ := strconv.ParseInt(loginUser.Id, 10, 64)
	err := c.SetDayActive(userId)
	if err != nil {
		logger.ErrZapLog(err, "SetDayActive fail")
	}

	err = c.SetActiveTimeAndIp(loginUser.Id, ctx.ClientIP())
	if err != nil {
		logger.ErrZapLog(err, "SetActiveTimeAndIp fail")
	}
}
