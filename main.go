package main

import (
	"fmt"
	"log"
	"net/http"
	"onpaper-api-go/app"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/settings"
	"onpaper-api-go/utils/quite"
)

func main() {
	// 初始化服务
	router := app.Init()

	// 启动服务
	//定义 server 结构体
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: router,
	}

	// 开启一个goroutine启动服务 启动监听
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 平滑关机
	quite.SmoothQuite(server)
	// 退出时关闭数据库链接
	defer mysql.Close()
	defer cache.Close()
	defer mongo.Close()
}
