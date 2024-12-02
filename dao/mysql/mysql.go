package mysql

import (
	"fmt"
	"onpaper-api-go/settings"
	"time"

	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func Init(config *settings.MySQLConfig) (err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DbName,
	)
	// 也可以使用MustConnect连接不成功就panic
	// 链接数据库
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		zap.L().Error("链接数据库失败", zap.Error(err))
		return
	}

	db.SetMaxOpenConns(config.MaxOpenConns) // 最大链接数
	db.SetMaxIdleConns(config.MaxIdleConns) // 最大空闲数
	db.SetConnMaxLifetime(time.Minute * 10) //设置数据库闲链接超时时间
	return
}

// Close 暴露close 方法
func Close() {
	_ = db.Close()
}
