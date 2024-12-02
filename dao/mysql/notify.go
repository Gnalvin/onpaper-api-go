package mysql

import (
	"database/sql"
	"github.com/pkg/errors"
	m "onpaper-api-go/models"
)

// SetNotifyUnread 设置喜欢通知未读数
func SetNotifyUnread(userId, uType string) (err error) {
	var sqlStr string
	switch uType {
	case "like":
		sqlStr = "UPDATE notify_unread_count SET `like` = `like` + 1 WHERE user_id = ?;"
	case "collect":
		sqlStr = "UPDATE notify_unread_count SET collect = collect + 1 WHERE user_id = ?;"
	case "focus":
		sqlStr = "UPDATE notify_unread_count SET follow = follow + 1 WHERE user_id = ?;"
	case "comment":
		sqlStr = "UPDATE notify_unread_count SET comment = comment + 1 WHERE user_id = ?;"
	case "commission":
		sqlStr = "UPDATE notify_unread_count SET commission = commission + 1 WHERE user_id = ?;"
	}

	_, err = db.Exec(sqlStr, userId)
	if err != nil {
		err = errors.Wrap(err, "SetNotifyUnread update  fail")
	}
	return
}

// GetNotifyUnreadCount 获取通知未读数
func GetNotifyUnreadCount(userId string) (count m.NotifyUnreadCount, err error) {
	sqlStr := "select `like`,follow,comment,at,collect,commission from notify_unread_count WHERE user_id = ?;"
	err = db.Get(&count, sqlStr, userId)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return
		}
		err = errors.Wrap(err, "GetNotifyUnreadCount sql fail")
	}
	return
}

// AckNotifyUnread 取消掉消息未读
func AckNotifyUnread(userId, uType string) (err error) {
	var sqlStr string
	switch uType {
	case "likeAndCollect":
		sqlStr = "UPDATE notify_unread_count SET `like` = 0 , collect=0 WHERE user_id = ?;"
	case "focus":
		sqlStr = "UPDATE notify_unread_count SET follow = 0 WHERE user_id = ?;"
	case "comment":
		sqlStr = "UPDATE notify_unread_count SET comment = 0 WHERE user_id = ?;"
	case "commission":
		sqlStr = "UPDATE notify_unread_count SET commission = 0 WHERE user_id = ?;"
	}

	_, err = db.Exec(sqlStr, userId)
	if err != nil {
		err = errors.Wrap(err, "AckNotifyUnread update  fail")
	}
	return
}

// GetNotifySetting 获取通知设置
func GetNotifySetting(userId string) (config m.NotifyConfig, err error) {
	sqlStr := "select collect,`like`,follow,at,comment,message from notify_config where user_id = ?"

	err = db.Get(&config, sqlStr, userId)
	if err != nil {
		err = errors.Wrap(err, "GetNotifySetting  fail")
	}
	return
}

// SettingNotifySetting 更新通知设置
func SettingNotifySetting(userId string, config m.NotifyConfig) (err error) {
	sqlStr := "UPDATE notify_config SET `like` = ?,comment=?,collect=?,message=?,follow=?,at=? WHERE user_id = ?;"
	_, err = db.Exec(sqlStr, config.Like, config.Comment, config.Collect, config.Message, config.Follow, config.At, userId)
	if err != nil {
		err = errors.Wrap(err, "SettingNotifySetting  fail")
	}
	return
}
