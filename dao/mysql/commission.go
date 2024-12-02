package mysql

import (
	"fmt"
	"github.com/pkg/errors"
	m "onpaper-api-go/models"
	"onpaper-api-go/utils/formatTools"
	"onpaper-api-go/utils/snowflake"
	"strings"
)

func GetUserCommissionScore(sender, artist string) (plan m.UserInvitePlan, err error) {
	sqlStr := `SELECT cc.user_id,username,avatar_name,SUM(send_finish + receive_finish) as finish ,rating FROM commission_count as cc
               left join user_profile as up on cc.user_id = up.user_id
               WHERE cc.user_id in (?,?)
			   GROUP BY cc.user_id,username,avatar_name,rating;`

	var userInfo []m.UserScoreInfo
	err = db.Select(&userInfo, sqlStr, sender, artist)

	for _, info := range userInfo {
		if info.UserId == sender {
			plan.Sender = info
		} else {
			plan.Artist = info
		}
	}

	return
}

// UpdateCommissionCuntAndEvaluate 更新接稿计数
func UpdateCommissionCuntAndEvaluate(senderId, receiveId string, nextStatus, nowStatus int8, e *m.Evaluate) (err error) {
	// 开启一个事务
	tx, err := db.Begin()
	if err != nil {
		err = errors.Wrap(err, "transaction begin failed")
		return
	}
	// 函数关闭时 如果出错 则回滚，没出错则 提交
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
		} else if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
			return
		}
	}()
	//下一阶段 个数 +1
	var columnName string
	switch nextStatus {
	case 0:
		columnName = "wait"
	case 1:
		columnName = "talk"
	case 2:
		columnName = "ing"
	case 3:
		columnName = "finish"
	case -1, -2:
		columnName = "close"
	}

	tempStr := `UPDATE commission_count
				SET receive_%[1]s = IF(user_id = %[2]s, receive_%[1]s + %[4]d, receive_%[1]s),
    			send_%[1]s = IF(user_id = %[3]s, send_%[1]s + %[4]d, send_%[1]s)
				WHERE user_id IN (%[2]s, %[3]s);`

	sqlStr := fmt.Sprintf(tempStr, columnName, receiveId, senderId, 1)

	_, err = tx.Exec(sqlStr)

	// 上一阶段 -1
	if nextStatus != 0 {
		switch nowStatus {
		case 0:
			columnName = "wait"
		case 1:
			columnName = "talk"
		case 2:
			columnName = "ing"
		}

		tempStr = `UPDATE commission_count
				SET receive_%[1]s = IF(user_id = %[2]s, receive_%[1]s - %[4]d, receive_%[1]s),
    			send_%[1]s = IF(user_id = %[3]s, send_%[1]s - %[4]d, send_%[1]s)
				WHERE user_id IN (%[2]s, %[3]s);`

		sqlStr = fmt.Sprintf(tempStr, columnName, receiveId, senderId, 1)
		_, err = tx.Exec(sqlStr)

	}

	if e != nil {
		e.EvaluateId = snowflake.CreateID()
		sql1 := `INSERT INTO commission_evaluate (evaluate_id,invite_id, invite_own ,receiver, sender,text ,rate_1, rate_2, rate_3, total_rating) 
			values (?,?,?,?,?,?,?,?,?,?)`

		_, err = tx.Exec(sql1, e.EvaluateId, e.InviteId, e.InviteOwn, e.Receiver, e.Sender, e.Text, e.Rate1, e.Rate2, e.Rate3, e.Score)
	}

	return
}

// GetPlanEvaluate 获取方案评论
func GetPlanEvaluate(planId string) (evaluate []m.Evaluate, err error) {
	sql1 := `SELECT invite_id,invite_own,receiver,sender,total_rating,rate_1,rate_2,rate_3,text,is_delete,createAT FROM commission_evaluate WHERE invite_id = ?`
	err = db.Select(&evaluate, sql1, planId)

	return
}

// GetNoEvaluateID 查找没有评价的约稿方案id
func GetNoEvaluateID(inviteIds []int64, userId string) (res map[int64]struct{}, err error) {
	if len(inviteIds) == 0 {
		return
	}

	sqlStr1 := `(SELECT invite_id FROM commission_evaluate WHERE invite_id = ? AND sender = ?)`

	sql1Strings := make([]string, 0, len(inviteIds))
	findIds := make([]interface{}, 0, len(inviteIds))
	for _, id := range inviteIds {
		sql1Strings = append(sql1Strings, sqlStr1)
		findIds = append(findIds, id, userId)
	}
	sqlStr1 = strings.Join(sql1Strings, " UNION ALL ")

	var haveEvaluateID []int64
	err = db.Select(&haveEvaluateID, sqlStr1, findIds...)
	if err != nil {
		err = errors.Wrap(err, "GetNoEvaluateID: sql1 get fail")
	}

	noEvaluate := formatTools.SliceDiff(inviteIds, haveEvaluateID)
	res = make(map[int64]struct{})
	for _, id := range noEvaluate {
		res[id] = struct{}{}
	}
	return
}

// SaveOneEvaluate 保存一条评论
func SaveOneEvaluate(e *m.Evaluate) (err error) {
	sql1 := `INSERT INTO commission_evaluate (evaluate_id,invite_id, invite_own ,receiver, sender,text ,rate_1, rate_2, rate_3, total_rating) 
			values (?,?,?,?,?,?,?,?,?,?)`

	_, err = db.Exec(sql1, e.EvaluateId, e.InviteId, e.InviteOwn, e.Receiver, e.Sender, e.Text, e.Rate1, e.Rate2, e.Rate3, e.Score)
	if err != nil {
		err = errors.Wrap(err, "mysql SaveOneEvaluate fail")
	}
	return
}

// GetUserReceiveEvaluate 获取用户收到的评论
func GetUserReceiveEvaluate(userId string, page uint8) (evaluate []m.EvaluateShow, err error) {
	// 查找已经相互评价且没有上传的评价
	sqlStr := ` SELECT a.evaluate_id, a.invite_id,a.sender,a.text,a.total_rating,a.createAT
				FROM commission_evaluate a
				JOIN commission_evaluate b
				ON a.invite_id = b.invite_id AND a.receiver = b.sender AND a.sender = b.receiver
				WHERE a.receiver = ? AND a.is_delete = 0 AND b.is_delete = 0
				ORDER BY a.createAT DESC
				limit ?,15`

	evaluate = make([]m.EvaluateShow, 0)
	err = db.Select(&evaluate, sqlStr, userId, 15*page)
	if err != nil {
		err = errors.Wrap(err, "mysql GetUserReceiveEvaluate fail")
	}

	return
}

// UpdateCommissionStatus 更新约稿状态
func UpdateCommissionStatus(isOpen bool, userId string) (err error) {
	sqlStr := `update user_profile set commission = ?,have_plan = 1 where user_id = ?`
	_, err = db.Exec(sqlStr, isOpen, userId)
	if err != nil {
		err = errors.Wrap(err, "mysql UpdateCommissionStatus fail")
	}
	return
}

// CheckCreatAcceptPermission 检测是否可以创建计划
func CheckCreatAcceptPermission(userId string) (ok bool, artCount int, err error) {

	sqlStr := `select count(*) from artwork where user_id = ? and is_delete = 0 and whoSee ='public' and state = 0`
	err = db.Get(&artCount, sqlStr, userId)
	if err != nil {
		err = errors.Wrap(err, "mysql CheckCreatAcceptPermission fail")
	}

	if artCount >= 5 {
		ok = true
	}

	return
}
