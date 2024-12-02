package mysql

import "errors"

var (
	ErrorUserExist         = errors.New("用户已存在")
	ErrorUserNotExist      = errors.New("用户不存在")
	ErrorEmailExist        = errors.New("邮箱已存在")
	ErrorEmailNotExist     = errors.New("邮箱不存在")
	ErrorPhoneExist        = errors.New("手机已存在")
	ErrorPhoneNotExist     = errors.New("手机不存在")
	ErrorInviteCodeInvalid = errors.New("邀请码无效")
)
