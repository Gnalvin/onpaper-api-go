package mysql

import (
	"github.com/pkg/errors"
	m "onpaper-api-go/models"
)

// CheckUserExistByName 通过用户名查询 是否存在
func CheckUserExistByName(userName string) (err error) {
	sqlStr := `SELECT  count(snow_id) FROM user WHERE username = ?;`
	// 查询 数据库
	var count int64
	err = db.Get(&count, sqlStr, userName)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "GetUserByName: sql get fail")
		return
	}
	//如果用户不存在 返回用户不存在错误
	if count == 0 {
		return ErrorUserNotExist
	}
	//如果用户存在 返回用户存在错误
	if count != 0 {
		return ErrorUserExist
	}
	return
}

// CheckUserExistByEmail 通过邮箱查询 是否存在
func CheckUserExistByEmail(userEmail string) (err error) {
	sqlStr := `SELECT  count(snow_id) FROM user WHERE email = ?;`
	var count int64
	// 查询 数据库
	err = db.Get(&count, sqlStr, userEmail)
	if err != nil {
		err = errors.Wrap(err, "GetUserByName: sql get fail")
		return
	}
	//如果邮箱不存在 返回用户不存在错误
	if count == 0 {
		return ErrorEmailNotExist
	}
	//如果邮箱存在 返回用户存在错误
	if count != 0 {
		return ErrorEmailExist
	}
	return
}

// CheckUserExistByPhone 通过手机查询 是否存在
func CheckUserExistByPhone(phone string) (isExist bool, err error) {
	sqlStr := `SELECT  count(snow_id) FROM user WHERE phone = ?;`
	var count int64
	// 查询 数据库
	err = db.Get(&count, sqlStr, phone)
	if err != nil {
		err = errors.Wrap(err, "CheckUserExistByPhone: sql get fail")
		return
	}
	//如果手机不存在 返回用户不存在错误
	if count == 0 {
		return false, nil
	}
	//如果手机存在 返回用户存在错误
	if count != 0 {
		return true, nil
	}
	return
}

// CheckInviteCode 验证邀请码有效性
func CheckInviteCode(code string) (err error) {
	if code == "" {
		return ErrorInviteCodeInvalid
	}
	sqlStr := `SELECT  count(*) FROM invite_code WHERE code = ? and used = 0;`
	var count int64
	// 查询 数据库
	err = db.Get(&count, sqlStr, code)
	if err != nil {
		err = errors.Wrap(err, "CheckInviteCode: sql get fail")
		return
	}
	//如果没有查到则无效
	if count == 0 {
		return ErrorInviteCodeInvalid
	}
	return
}

// GetBindingInfo 获取账号绑定相关信息
func GetBindingInfo(userId string) (info m.UserBindingInfo, err error) {
	sql := `select phone,email,password from user where snow_id = ?`
	err = db.Get(&info, sql, userId)
	if err != nil {
		err = errors.Wrap(err, "GetBindingInfo: sql get fail")
		return
	}
	return
}

// ChangeBindingEmail 修改绑定邮箱
func ChangeBindingEmail(userId, email string) (err error) {
	sql := `update user set email = ? where snow_id = ?`
	_, err = db.Exec(sql, email, userId)
	if err != nil {
		err = errors.Wrap(err, "ChangeBindingEmail: sql fail")
		return
	}
	return
}

// ChangePassword 修改密码
func ChangePassword(userId, password string) (err error) {
	sql := `update user set password = ? where snow_id = ?`
	_, err = db.Exec(sql, password, userId)
	if err != nil {
		err = errors.Wrap(err, "ChangePassword: sql fail")
		return
	}
	return
}

// ChangePhone 修改密码
func ChangePhone(userId, phone string) (err error) {
	sql := `update user set phone = ? where snow_id = ?`
	_, err = db.Exec(sql, phone, userId)
	if err != nil {
		err = errors.Wrap(err, "ChangePhone: sql fail")
		return
	}
	return
}
