package router

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"onpaper-api-go/logger"
	hm "onpaper-api-go/middleware/handleMiddle"
)

func Setup(mode string) (router *gin.Engine) {
	// 配置 gin 运行环境
	switch mode {
	// 生产环境需要把这 config配置改成 release
	case "release":
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		//router.Use(hm.Cors()) nginx 解决了就不需要这个
		router.Use(hm.OverrideMethod(router))
		router.Use(logger.GinLogger(), logger.GinRecovery(true))
	case "debug":
		gin.SetMode(gin.DebugMode)
		// 生成路由对象
		router = gin.New()
		pprof.Register(router)
		router.Use(hm.OverrideMethod(router))
		// 添加上官方的 logger 在终端显示
		router.Use(gin.Logger(), logger.GinLogger(), logger.GinRecovery(true))
	case "test":
		gin.SetMode(gin.TestMode)
	}

	// 注册用户路由
	type routerFunc func(*gin.Engine)
	// 定义 router类型路由切片
	var routerList []routerFunc

	routerList = append(routerList,
		userRouter,
		authRouter,
		fileRouter,
		artworkRouter,
		feedRouter,
		tagAndTopicRouter,
		trendRouter,
		commentRouter,
		messageRouter,
		notifyRouter,
		feedbackRouter,
		commissionRouter,
	)

	for _, routerFuncItem := range routerList {
		// 取出路由 将router传入
		routerFuncItem(router)
	}

	return
}
