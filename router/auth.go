package router

import (
	"github.com/gin-gonic/gin"
	ctl "onpaper-api-go/controller"
	cm "onpaper-api-go/middleware/cacheMiddle"
	hm "onpaper-api-go/middleware/handleMiddle"
)

// authRouter 验证信息相关的路由
func authRouter(router *gin.Engine) {
	//定义路由组
	r := router.Group("/auth")
	rMustAuth := router.Group("/auth", hm.VerifyAuthMust)

	//用户密码登陆接口
	r.POST("/login/password", hm.VerifyLogin, cm.InitUserData, ctl.SignIn, cm.SetActiveData)
	// 用户验证码登陆/注册接口
	r.POST("/login/phone", hm.VerifyAccountForm, ctl.HandleLoginOrRegister, cm.InitUserData, ctl.SignIn, cm.SetActiveData)
	// 用户邮箱验证码登陆
	r.POST("/login/email", hm.VerifyEmailLogin, ctl.HandleEmailLogin, cm.InitUserData, ctl.SignIn, cm.SetActiveData)

	//用于请求 AccessToken的 接口
	r.GET("/accesstoken", hm.VerifyRefreshToken, ctl.CreateToken, cm.SetActiveData)
	//发送邮件验证码
	r.GET("/emailcode", hm.VerifyAuth, hm.VerifyEmailFormat, hm.VerifyQuerySign, ctl.SendEmailCode)
	//发送手机验证码
	r.GET("/phonecode", hm.VerifyPhoneFormat, hm.VerifyQuerySign, ctl.VerifyLoginPhone, ctl.SendPhoneCode)

	//登陆后向密保手机发送验证码
	rMustAuth.GET("/code", hm.VerifyQuerySign, ctl.SendSafetyPhoneCode)
	//通过手机验证码验证所有权发放权限
	rMustAuth.GET("/owner", hm.VerifySafetyCode, ctl.GetAuthToken)
	//登陆后向新绑定的手机发送验证码
	rMustAuth.GET("/newphone", hm.VerifyPhoneFormat, hm.VerifyQuerySign, ctl.SendPhoneCode)

	//获取相关安全绑定信息
	rMustAuth.GET("/binding", ctl.GetBindingInfo)
	//换绑定
	//换邮箱绑定
	rMustAuth.PATCH("/change/email", hm.VerifyChangeEmail, ctl.ChangBindingEmail)
	//换密码
	rMustAuth.PATCH("/change/password", hm.VerifyChangePassword, ctl.ChangePassword)
	//换绑定手机
	rMustAuth.PATCH("/change/phone", hm.VerifyChangePhone, ctl.ChangeBindingPhone)

	// 查询用户是否存在
	r.GET("/verifyname", ctl.VerifyName)
	// 查询邮箱是否存在
	r.GET("/emailexist", hm.VerifyEmailFormat, ctl.VerifyEmailExist)
	// 检查手机是否存在
	r.GET("/phoneexist", hm.VerifyPhoneFormat, ctl.VerifyPhoneExist)

	// 小程序获取 微信id
	r.GET("/wxid", hm.VerifyMiniProgramCode, ctl.GetUserWxId)
	// 小程序登陆/注册
	r.GET("/wx_login", hm.VerifyMiniProgramCode, ctl.WxGetUserPhone, ctl.HandleWxLogin, cm.InitUserData, ctl.SignIn, cm.SetActiveData)

	// 获取小程序跳转url
	r.GET("/wx_url", ctl.GetWxUrl)

	// 请求获取 图片上传 sts授权接口
	router.GET("/uploadimg/sts", hm.VerifyAuthMust, ctl.ReturnSTSData)
}
