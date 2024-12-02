package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/pkg/errors"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/encrypt"
	"time"
)

// SetEmailVerifyCode  到Redis里面保存 邮箱验证码
func SetEmailVerifyCode(email, code string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf(AuthEmail, email)
	_, err = Rdb.Set(ctx, key, code, time.Minute*15).Result()
	if err != nil {
		return err
	}
	return
}

// GetEmailVerifyCode Redis获取邮箱验证码
func GetEmailVerifyCode(email string) (val string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf(AuthEmail, email)
	val = Rdb.Get(ctx, key).Val()
	if err != nil {
		// 如果返回的错误是key不存在
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}

	return
}

// SetPhoneVerifyCode  到Redis里面保存 手机验证码
func SetPhoneVerifyCode(phone, code string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf(AuthPhone, phone)
	//短信提示5分钟过期 这里10分钟 留些余地
	_, err = Rdb.Set(ctx, key, code, time.Minute*10).Result()
	if err != nil {
		return err
	}
	return
}

// GetPhoneVerifyCode Redis获取手机验证码
func GetPhoneVerifyCode(phone string) (val string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf(AuthPhone, phone)
	val = Rdb.Get(ctx, key).Val()
	if err != nil {
		// 如果返回的错误是key不存在
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return
}

func SetTokenMd5(userInfo m.UserTokenPayload) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf(AuthToken, userInfo.Id)
	md5 := encrypt.CreatMd5(userInfo.MD5 + userInfo.Phone + userInfo.Email)
	_, err = Rdb.Set(ctx, key, md5, time.Hour*24*90).Result()
	if err != nil {
		return err
	}

	return
}

func GetTokenMd5(userId string) (md5 string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := fmt.Sprintf(AuthToken, userId)
	md5, err = Rdb.Get(ctx, key).Result()
	if err != nil {
		// 如果返回的错误是key不存在
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return
}

func GetWxToken() (token string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	token, err = Rdb.Get(ctx, WxAccessToken).Result()
	if err != nil {
		err = errors.Wrap(err, "GetWxToken fail")
		return
	}
	return
}
