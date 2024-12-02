package singleFlight

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
	"time"
)

var sg = &singleflight.Group{}

// Do 阻塞函数
func Do(ctx context.Context, key string, fn func() (interface{}, error)) (res interface{}, err error) {
	// 启动一个 Goroutine 定时 forget 一下 提高单位时间 fn 的使用率
	// 相当于将 rps 从 1rps 提高到了 10rps
	go func() {
		time.Sleep(100 * time.Millisecond)
		sg.Forget(key)
	}()

	result := sg.DoChan(key, fn)

	// 超时控制
	select {
	case r := <-result:
		return r.Val, r.Err
	case <-ctx.Done():
		err = errors.Wrap(ctx.Err(), fmt.Sprintf("%s singleflight Context is timeout", key))
		return nil, err
	}
}
