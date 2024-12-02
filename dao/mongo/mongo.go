package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"onpaper-api-go/logger"
	"onpaper-api-go/settings"
	"time"
)

var Mgo *mongo.Database
var client *mongo.Client

func Init(config *settings.MongodbConfig) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("mongodb://%s:%s@%s", config.User, config.Password, config.Host)
	option := options.Client().ApplyURI(url)
	option.SetMaxPoolSize(config.PoolSize)
	option.SetMinPoolSize(5)             //最少保持5个链接
	option.SetMaxConnIdleTime(time.Hour) // 连接池中某个连接的空闲时间超过该值，将丢弃该连接并重新新建立一个连接

	client, err = mongo.Connect(ctx, option)
	if err != nil {
		return
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return
	}

	Mgo = client.Database(config.DbName)

	return
}

// Close 暴露close 方法
func Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.Disconnect(ctx)
	if err != nil {
		logger.ErrZapLog(err, "close mongodb fail")
	}
}
