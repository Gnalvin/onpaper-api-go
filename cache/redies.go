package cache

import (
	"context"
	"github.com/go-redis/redis/v9"
	"onpaper-api-go/settings"
	"time"
)

var Rdb *redis.Client

func Init(config *settings.RedisConfig) (err error) {

	Rdb = redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Username:     config.UserName,
		Password:     config.Password, // 密码
		DB:           0,               // 数据库
		PoolSize:     config.PoolSize, // 连接池大小
		MinIdleConns: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = Rdb.Ping(ctx).Result()
	defer cancel()

	return
}

// Close 暴露close 方法
func Close() {
	_ = Rdb.Close()
}
