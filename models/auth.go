package models

// UserLoginInfo 用户登录时上传的信息
type UserLoginInfo struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserTokenPayload token里面携带的用户信息
type UserTokenPayload struct {
	Id        string
	Phone     string
	Email     string
	MD5       string
	TokenType string
}

type STSResult struct {
	Token    Credentials `json:"token"`
	FileName []string    `json:"fileName"`
}

type Credentials struct {
	AccessKeyId     *string `json:"accessKeyId"`
	AccessKeySecret *string `json:"accessKeySecret"`
	SecurityToken   *string `json:"securityToken"`
	Expiration      *string `json:"expiration"`
}

// UserEmailLogin 用户邮箱验证码登陆需要的数据
type UserEmailLogin struct {
	Email      string `json:"email" binding:"required"`
	VerifyCode string `json:"verifyCode" binding:"required"`
}

// UserBindingInfo 用户设置的安全信息
type UserBindingInfo struct {
	Phone    string  `json:"phone" db:"phone"`
	Email    *string `json:"email" db:"email"`
	Password string  `json:"password" db:"password"`
}

// MiniProgramCode	小程序上传的验证Code
type MiniProgramCode struct {
	Code string `form:"code" binding:"required"`
}

// WXIdRes 小程序获取OpenId 返回的结果
type WXIdRes struct {
	Openid     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionId    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WxGetUserPhoneRes 微信获取用户手机 返回结果
type WxGetUserPhoneRes struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	PhoneInfo struct {
		PhoneNumber     string `json:"phoneNumber"`
		PurePhoneNumber string `json:"purePhoneNumber"`
		CountryCode     int    `json:"countryCode"`
		Watermark       struct {
			Timestamp int    `json:"timestamp"`
			Appid     string `json:"appid"`
		} `json:"watermark"`
	} `json:"phone_info"`
}

// WxUrl 小程序获取跳转链接的返回结果
type WxUrl struct {
	Errcode  int    `json:"errcode"`
	Errmsg   string `json:"errmsg"`
	Openlink string `json:"openlink"`
}
