package app

import (
	"fmt"
	"onpaper-api-go/cache"
	"onpaper-api-go/dao/mongo"
	"onpaper-api-go/dao/mysql"
	"onpaper-api-go/logger"
	"onpaper-api-go/router"
	"onpaper-api-go/settings"
	"onpaper-api-go/utils/jwt"
	"onpaper-api-go/utils/snowflake"

	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
)

// Init Index 初始化所有配置
func Init() (r *gin.Engine) {
	// 1.加载配置文件
	if err := settings.ConfigInit(); err != nil {
		fmt.Printf("settings.Init() 初始化配置失败 ,err:%v\n", err)
		return
	}
	// 2.初始化日志
	if err := logger.Init(settings.Conf.LogConfig); err != nil {
		fmt.Printf("logger.Init() 初始化日志失败, err:%v\n", err)
		return
	}
	defer zap.L().Sync() //退出程序时将缓存的日记落盘
	zap.L().Info("logger init success...")

	// 3.初始化Mysql 链接
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	zap.L().Info("Mysql init success...")

	// 4.初始化Redis 链接
	if err := cache.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init Redis failed, err:%v\n", err)
		return
	}
	zap.L().Info("Redis init success...")

	// 5.初始化 MongoDB 链接
	if err := mongo.Init(settings.Conf.MongodbConfig); err != nil {
		fmt.Printf("init mongodb failed, err:%v\n", err)
		return
	}
	zap.L().Info("MongoDB init success...")

	// 5.注册路由
	r = router.Setup(settings.Conf.Mode)
	zap.L().Info("router init success...")

	//6.初始化雪花算法
	err := snowflake.Init(settings.Conf.SnowStartTime, settings.Conf.MachineId)
	if err != nil {
		zap.L().Error("snowflake init fail", zap.Error(err))
	}
	zap.L().Info("snowflake init success...")

	//7.初始化 jwt
	err = jwt.Init()
	if err != nil {
		zap.L().Error("jwt init fail", zap.Error(err))
	}
	zap.L().Info("jwt init success...")
	return
}
