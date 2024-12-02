package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"time"
)

// SetDayActive 设置日活统计
func SetDayActive(userId int64) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pipe := Rdb.Pipeline()
	now := time.Now()
	// 记录日活
	today := now.Format("2006-01-02")
	dKey := fmt.Sprintf(ActiveDay, today)
	pipe.SetBit(ctx, dKey, userId, 1)

	// 记录月活
	month := now.Format("2006-01")
	mKey := fmt.Sprintf(ActiveMonth, month)
	pipe.SetBit(ctx, mKey, userId, 1)

	_, err = pipe.Exec(ctx)
	return
}

// SetActiveTimeAndIp 设置活跃时间和Ip地址
func SetActiveTimeAndIp(userId string, ip string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 记录用户最后活跃时间
	now := time.Now().Unix()
	err = Rdb.ZAdd(ctx, ActiveTime, redis.Z{Score: float64(now), Member: userId + "&" + ip}).Err()

	return
}
