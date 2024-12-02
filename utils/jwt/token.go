package jwt

import (
	"github.com/pkg/errors"
	"onpaper-api-go/models"
	"onpaper-api-go/settings"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

var publicKeyByte []byte
var privateKeyByte []byte

func Init() (err error) {
	// 获取公钥字节
	publicKeyByte, err = os.ReadFile(settings.Conf.TokenPublicKeyPath)
	if err != nil {
		return
	}
	// 获取私钥字节
	privateKeyByte, err = os.ReadFile(settings.Conf.TokenPrivateKeyPath)
	if err != nil {
		return
	}

	return
}

// CreateRefreshToken createToken 生成一个RS256验证的Token
// Token里面包括的值，可以自己根据情况添加，
// 非对称加密 公钥解密 私钥颁发
func CreateRefreshToken(userInfo *models.UserTokenPayload) (tokenStr string, err error) {
	//解析私钥
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyByte)
	if err != nil {
		return
	}

	var expTime time.Duration
	// 补时 将 RefreshToken 过期时间补到 凌晨4点
	var overTime int
	nowHour := time.Now().Hour() //获取当前小时
	switch {
	case nowHour > 4:
		overTime = 24 - nowHour + 4
	default:
		overTime = 4 - nowHour
	}
	expTime = time.Hour * (24*31*6 + time.Duration(overTime)) // 过期时间，目前是六个月后的凌晨4点

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat":   time.Now().Unix(),              // Token颁发时间
		"nbf":   time.Now().Unix(),              // Token生效时间
		"exp":   time.Now().Add(expTime).Unix(), // Token过期时间
		"iss":   "onpaper.cn",                   // 颁发者
		"sub":   "RefreshToken",                 // 主题
		"uId":   userInfo.Id,
		"email": userInfo.Email,
		"phone": userInfo.Phone,
		"md5":   userInfo.MD5,
	})

	// 生成token
	tokenStr, err = token.SignedString(privateKey)
	if err != nil {
		return
	}
	return
}

// CreateAccessToken 时间短获取资源的临时token
func CreateAccessToken(userInfo *models.UserTokenPayload, exp time.Duration) (tokenStr string, err error) {
	//解析私钥
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyByte)
	if err != nil {
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat":   time.Now().Unix(),          // Token颁发时间
		"nbf":   time.Now().Unix(),          // Token生效时间
		"exp":   time.Now().Add(exp).Unix(), // Token过期时间
		"iss":   "onpaper.cn",               // 颁发者
		"sub":   "AccessToken",              // 主题
		"uId":   userInfo.Id,
		"email": userInfo.Email,
		"phone": userInfo.Phone,
		"md5":   ".",
	})

	// 生成token
	tokenStr, err = token.SignedString(privateKey)
	if err != nil {
		return
	}
	return
}

// ParseToken 解析认证token
func ParseToken(tokenStr string) (userInfo models.UserTokenPayload, err error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyByte)

	// 解析token 需要传入一个返回publicKey的 回调函数
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// 基于JWT的第一部分中的alg字段值进行一次验证
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("验证Token的加密类型错误")
		}
		return publicKey, nil
	})
	if err != nil {
		err = errors.Wrap(err, "jwt.Parse fail")
		return
	}

	// 解析成功可以拿到 token对象获取 payload
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Json反序列化数字到interface{}类型的值中，默认解析为float64类型,需要在 int()强转一次
		// 直接 断言int 会报错
		if (claims["uId"] == nil) || (claims["phone"] == nil) || (claims["md5"] == nil) || (claims["email"] == nil) {
			err = errors.New("token payload error")
			userInfo = models.UserTokenPayload{}
			return
		}

		userInfo.Id = claims["uId"].(string)
		userInfo.Phone = claims["phone"].(string)
		userInfo.Email = claims["email"].(string)
		userInfo.MD5 = claims["md5"].(string)
		userInfo.TokenType = claims["sub"].(string)

	}
	return
}
