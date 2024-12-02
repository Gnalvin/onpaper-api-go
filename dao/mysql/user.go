package mysql

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"onpaper-api-go/models"
	"onpaper-api-go/settings"
	"onpaper-api-go/utils/encrypt"
	"strconv"
	"strings"
)

// GetUserByPhone 通过用户名查询用户
func GetUserByPhone(phone string) (userInfo models.UserTableInfo, err error) {
	sqlStr := `SELECT snow_id,user.userName,password,phone,email,forbid,avatar_name FROM user 
			  left join user_profile up on user.snow_id = up.user_id
			  WHERE phone = ?;`
	// 查询 数据库
	err = db.Get(&userInfo, sqlStr, phone)
	if err != nil {
		// 返回错误信息
		err = errors.Wrap(err, "GetUserByPhone: sql get fail")
		return
	}
	return
}

// GetUserByEmail 通过邮箱查询用户
func GetUserByEmail(userEmail string) (userInfo models.UserTableInfo, err error) {
	sqlStr := `SELECT snow_id,user.userName,password,phone,email,forbid,avatar_name  FROM user 
               left join user_profile up on user.snow_id = up.user_id              
               WHERE email = ?;`

	// 查询 数据库
	err = db.Get(&userInfo, sqlStr, userEmail)
	if err != nil {
		err = errors.Wrap(err, "GetUserByName: sql get fail")
		return
	}
	return
}

// CreatUserInfo 创建用户
func CreatUserInfo(info *models.LoginForm) (err error) {
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

	var eg errgroup.Group

	eg.Go(func() (mErr error) {
		sqlStr1 := `INSERT INTO user (snow_id,username,password,phone) VALUES (?,?,?,?);`
		_, mErr = tx.Exec(sqlStr1, info.SnowId, info.UserName, info.Password, info.Phone)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr1 into fail")
		}
		return
	})

	eg.Go(func() (mErr error) {
		sqlStr2 := `INSERT INTO user_profile (user_id,username) VALUES (?,?);`
		_, mErr = tx.Exec(sqlStr2, info.SnowId, info.UserName)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr2 into fail")
		}
		return
	})

	eg.Go(func() (mErr error) {
		sqlStr3 := `INSERT INTO user_intro (user_id) VALUES (?);`
		_, mErr = tx.Exec(sqlStr3, info.SnowId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr3 into fail")
		}
		return
	})

	eg.Go(func() (mErr error) {
		sqlStr4 := `INSERT INTO user_count (user_id) VALUES (?);`
		_, mErr = tx.Exec(sqlStr4, info.SnowId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr4 into fail")
		}

		sqlStr9 := `INSERT INTO commission_count (user_id) VALUES (?);`
		_, mErr = tx.Exec(sqlStr9, info.SnowId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr6 into fail")
		}
		return
	})

	eg.Go(func() (mErr error) {
		sqlStr5 := `INSERT INTO avatar (user_id) VALUES (?);`
		_, mErr = tx.Exec(sqlStr5, info.SnowId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr5 into fail")
		}

		sqlStr6 := `INSERT INTO banner (user_id) VALUES (?);`
		_, mErr = tx.Exec(sqlStr6, info.SnowId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr6 into fail")
		}

		return
	})

	// 邀请码
	eg.Go(func() (mErr error) {
		code1 := encrypt.RandStr(7, "all")
		code2 := encrypt.RandStr(7, "all")
		code3 := encrypt.RandStr(7, "all")
		sqlStr7 := `INSERT INTO invite_code (owner, code) VALUES (?,?),(?,?),(?,?);`
		_, mErr = tx.Exec(sqlStr7, info.SnowId, code1, info.SnowId, code2, info.SnowId, code3)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr7 into fail")
		}
		// 如果使用的是万能马
		if info.InviteCode == settings.Conf.MagicCode {
			sqlStr8 := `insert into invite_code (used,code,owner) values (?,?,?)`
			_, mErr = tx.Exec(sqlStr8, info.SnowId, "magic", info.SnowId)
			if mErr != nil {
				mErr = errors.Wrap(mErr, "CreatUserInfo magicCode into fail")
			}
		} else if info.InviteCode != "" {
			sqlStr8 := `update invite_code set used = ? where code = ?`
			_, mErr = tx.Exec(sqlStr8, info.SnowId, info.InviteCode)
			if mErr != nil {
				mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr8 into fail")
			}
		}
		return
	})

	eg.Go(func() (mErr error) {
		sqlStr9 := `INSERT INTO notify_config (user_id) VALUES (?);`
		_, mErr = tx.Exec(sqlStr9, info.SnowId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr9 into fail")
		}

		sqlStr10 := `INSERT INTO notify_unread_count (user_id) VALUES (?);`
		_, mErr = tx.Exec(sqlStr10, info.SnowId)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "CreatUserInfo sqlStr10 into fail")
		}
		return
	})

	err = eg.Wait()
	if err != nil {
		return
	}

	// 记录创建日志
	zap.L().Info("creat a new user", zap.Any("userInfo", info))
	return
}

// GetUserProfileById 通过 用户id 查询 profile 资料
func GetUserProfileById(userId string) (models.UserProfileTableInfo, error) {
	var profile models.UserProfileTableInfo
	var eg errgroup.Group

	eg.Go(func() error {
		sqlStr2 := `SELECT up.username,work_email,birthday,sex, user_id, QQ,Weibo,Twitter,Pixiv,Bilibili,WeChat,
       			address,expect_work,software,create_style,banner_name,avatar_name,v_status,v_tag,commission,email
		 		FROM user_profile as up
		 		left join user as u on u.snow_id = up.user_id
		 		WHERE user_id = ?;`
		// 查询 数据库
		err := db.Get(&profile, sqlStr2, userId)
		if err != nil {
			// 返回错误信息
			err = errors.Wrap(err, "GetUserProfileById: sqlStr2 get fail")
		}
		return err
	})

	eg.Go(func() (err error) {
		profile.Count, err = GetUserCount(userId)
		if err != nil {
			// 返回错误信息
			err = errors.Wrap(err, "GetUserProfileById: sqlStr3 get fail")
		}
		return err
	})

	eg.Go(func() error {
		// 查询个人简介
		sqlStr4 := `SELECT introduce from user_intro WHERE user_id = ?`
		err := db.Get(&profile, sqlStr4, userId)
		if err != nil {
			// 返回错误信息
			err = errors.Wrap(err, "GetUserProfileById: sqlStr4 get fail")
		}
		return err
	})

	err := eg.Wait()

	return profile, err
}

// GetUserFocusUserId 查询用户所有已关注的用户id
func GetUserFocusUserId(userId string) (focusList []string, err error) {
	sqlStr := `SELECT focus_id FROM user_focus WHERE user_id = ?  and is_cancel = 0`
	_ = db.Select(&focusList, sqlStr, userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserFocusUserId sqlStr into fail")
		return
	}
	return
}

// UpdateUserName 修改用户名
func UpdateUserName(userName string, userId string) (err error) {
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
	// 更改 user表
	sqlStr1 := `
		UPDATE user
		SET username=?
		WHERE snow_id=?;`
	_, err = tx.Exec(sqlStr1, userName, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserName sqlStr1 into fail")
		return
	}
	// 更改 user_profile 表
	sqlStr2 := `
		UPDATE user_profile
		SET username=?
		WHERE user_id=?;`
	_, err = tx.Exec(sqlStr2, userName, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserName sqlStr2 into fail")
		return
	}

	return
}

// UpdateUserSex 修改用户性别
func UpdateUserSex(userSex string, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET sex=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, userSex, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserSex sqlStr into fail")
		return
	}

	return
}

// UpdateUserSns 更新社交链接
func UpdateUserSns(userSns models.SnsLinkData, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET qq=?,weibo=?,twitter=?,pixiv=?,WeChat=?,Bilibili=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, userSns.QQ, userSns.Weibo, userSns.Twitter, userSns.Pixiv, userSns.WeChat, userSns.Bilibili, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserSns sqlStr into fail")
		return
	}

	return
}

// UpdateUserWorkEmail 保存用户工作邮箱
func UpdateUserWorkEmail(workEmail string, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET work_email=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, workEmail, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserWorkEmail sqlStr into fail")
		return
	}

	return
}

// UpdateUserBirthday 保存用户的生日
func UpdateUserBirthday(birthday string, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET birthday=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, birthday, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserBirthday sqlStr into fail")
		return
	}

	return
}

// UpdateUserIntroduce 保存用户的 其他介绍
func UpdateUserIntroduce(introduce string, userId string) (err error) {
	sqlStr := `
		UPDATE user_intro
		SET introduce=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, introduce, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserIntroduce sqlStr into fail")
		return
	}

	return
}

// UpdateUserAddress 保存用户的居住区域
func UpdateUserAddress(address string, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET address=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, address, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserAddress sqlStr into fail")
		return
	}

	return
}

// UpdateUserExpectWork 更新用户期望工作
func UpdateUserExpectWork(expectWork string, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET expect_work=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, expectWork, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserExpectWork sqlStr into fail")
		return
	}

	return
}

// UpdateUserCreateStyle 更新用户创作风格
func UpdateUserCreateStyle(createStyle string, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET create_style=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, createStyle, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserCreateStyle sqlStr into fail")
		return
	}

	return
}

// UpdateUserSoftware 更新用户常用软件
func UpdateUserSoftware(software string, userId string) (err error) {
	sqlStr := `
		UPDATE user_profile
		SET software=?
		WHERE user_id=?;`
	_, err = db.Exec(sqlStr, software, userId)
	if err != nil {
		err = errors.Wrap(err, "UpdateUserSoftware sqlStr into fail")
		return
	}

	return
}

// GetUserNavDataById 查询用户导航栏信息
func GetUserNavDataById(userId string) (userNavData models.UserNavData, err error) {
	sqlStr1 := `SELECT avatar_name,banner_name,username,up.user_id,fans,following,likes FROM user_profile as up
                left join user_count uc on up.user_id = uc.user_id                        
                WHERE uc.user_id = ?;`
	err = db.Get(&userNavData, sqlStr1, userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserNavDataById: sql1 get fail")
		return
	}

	var unRead models.NotifyUnreadCount
	sqlStr2 := "SELECT comment,`like`,collect,follow,commission,at FROM notify_unread_count WHERE user_id = ?;"
	err = db.Get(&unRead, sqlStr2, userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserNavDataById: sql2 get fail")
		return
	}
	userNavData.NotifyUnread = unRead.Count()
	return
}

// SaveUserFocus 到数据库保存用户关注信息
func SaveUserFocus(focusData models.VerifyUserFocus, userId string) (isChange bool, err error) {
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

	sqlStr1 := `INSERT INTO user_focus (user_id,focus_id,is_cancel) VALUES (?,?,?)  
  				ON DUPLICATE KEY UPDATE is_cancel=?; `
	result, err := tx.Exec(sqlStr1, userId, focusData.FocusId, focusData.IsCancel, focusData.IsCancel)
	if err != nil {
		err = errors.Wrap(err, "SaveUserFocus: sql1 get fail")
		return
	}
	// 影响的行数 如果没有影响行数 说明请求没有实际变化
	count, _ := result.RowsAffected()

	if count == 0 {
		return
	} else {
		isChange = true
	}

	return
}

// GetUserRankData 获取用户排名数据
func GetUserRankData(rankType string) (userData []models.UserBigCard, err error) {
	var sqlStr1 string

	switch rankType {
	case "new":
		sqlStr1 = `SELECT user_id from rank_user_new
              	   WHERE rank_date = (SELECT rank_date FROM rank_user_new ORDER BY rank_date DESC LIMIT 1)
				   order by score desc ,collects desc ,likes desc 
				   limit 100 `
	case "girl":
		sqlStr1 = `SELECT user_id from rank_user_girl
               	   WHERE rank_date = (SELECT rank_date FROM rank_user_girl ORDER BY rank_date DESC LIMIT 1)		
				   order by score desc ,collects desc ,likes desc 
				   limit 100 `
	case "boy":
		sqlStr1 = `SELECT user_id from rank_user_boy
                   WHERE rank_date = (SELECT rank_date FROM rank_user_boy ORDER BY rank_date DESC LIMIT 1)		
				   order by score desc ,collects desc ,likes desc 
				   limit 100 `
	case "like":
		sqlStr1 = `SELECT user_id from rank_user_like
               	   WHERE rank_date = (SELECT rank_date FROM rank_user_like ORDER BY rank_date DESC LIMIT 1)
				   order by likes desc 
				   limit 100 `
	case "collect":
		sqlStr1 = `SELECT user_id from rank_user_collect
                   WHERE rank_date = (SELECT rank_date FROM rank_user_collect ORDER BY rank_date DESC LIMIT 1)
				   order by collects desc 
				   limit 100 `
	}

	var userIds []string
	// 查询 数据库
	err = db.Select(&userIds, sqlStr1)
	if err != nil {
		err = errors.Wrap(err, "GetUserRankData: sql1 get fail")
		return
	}

	userData, err = BatchGetUserAllInfo(userIds, 4)
	if err != nil {
		err = errors.Wrap(err, "GetUserRankData: BatchGetUserAllInfo get fail")
		return
	}

	return
}

// GetUserFocusIdList 获取用户关注名单
func GetUserFocusIdList(userId string, page int) (focusId []string, err error) {
	// 查询关注名单
	sqlStr1 := `select focus_id from user_focus  
                WHERE user_id=? and is_cancel= 0
				order by updateAt desc
				limit ?,50;`
	err = db.Select(&focusId, sqlStr1, userId, page*50)
	if err != nil {
		err = errors.Wrap(err, "GetUserFocusList sql1 fail")
		return
	}

	return
}

func GetUserFansList(userId string, page int) (fansId []string, err error) {
	// 查询关注名单
	sqlStr1 := `select user_id from user_focus  
                WHERE focus_id=? and is_cancel= 0
				order by updateAt desc
				limit ?,50;`
	err = db.Select(&fansId, sqlStr1, userId, page*50)
	if err != nil {
		err = errors.Wrap(err, "GetUserFocusList sql1 fail")
		return
	}
	return
}

// SearchUserByName 通过名字查找user
func SearchUserByName(searchText string) (searchData []models.UserSimpleInfo, likeData []models.UserSimpleInfoCount, err error) {

	var eg errgroup.Group
	eg.Go(func() (mErr error) {
		sql1 := `SELECT user_id,avatar_name,username,v_status,v_tag FROM user_profile WHERE username  = ? LIMIT 1`
		mErr = db.Select(&searchData, sql1, searchText)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "SearchUserByName: sql1 get fail")
		}
		return
	})

	eg.Go(func() (mErr error) {
		sql2 := `SELECT up.user_id,avatar_name,username,likes,v_status,v_tag FROM user_profile as up
			LEFT JOIN user_count as uc ON  uc.user_id = up.user_id
			WHERE username  like ? and  username != ?
			ORDER BY likes DESC
			limit 10`

		mErr = db.Select(&likeData, sql2, searchText+"%", searchText)
		if mErr != nil {
			mErr = errors.Wrap(mErr, "SearchUserByName: sql2 get fail")
		}
		return
	})

	err = eg.Wait()

	return
}

// SearchOurFocus 精准查找自己关注的用户
func SearchOurFocus(searchText, userId string) (searchData []models.UserSimpleInfo, err error) {
	sql := `SELECT user_id,username,avatar_name FROM user_profile 
			WHERE username like ? and user_id in (
			SELECT focus_id FROM user_focus WHERE user_id = ? AND is_cancel = 0)`

	err = db.Select(&searchData, sql, searchText+"%", userId)
	if err != nil {
		err = errors.Wrap(err, "SearchOurFocus: get fail")
	}

	return
}

// GetUserPanel 查询用户面板信息
func GetUserPanel(userId string) (userInfo models.UserPanel, err error) {
	var eg errgroup.Group

	eg.Go(func() error {
		sql1 := `select avatar_name,up.user_id,username,banner_name,likes,collects,fans,v_status,v_tag,commission,have_plan from user_profile as up 
			left join user_count uc on up.user_id = uc.user_id
			where uc.user_id = ?
			`
		_err := db.Get(&userInfo, sql1, userId)
		if _err != nil {
			_err = errors.Wrap(_err, "GetUserPanel sql1 fail")
		}
		return _err
	})

	eg.Go(func() error {
		sql2 := `select left(introduce,30) as introduce from user_intro where user_id = ?`
		_err := db.Get(&userInfo, sql2, userId)
		if _err != nil {
			_err = errors.Wrap(_err, "GetUserPanel sql2 fail")
		}
		return _err
	})

	eg.Go(func() error {
		sql3 := `SELECT artwork_id,cover,user_id from artwork 
            WHERE user_id = ? and is_delete = 0 and whoSee ='public' and state = 0
			ORDER BY artwork_id DESC LIMIT 3`
		_err := db.Select(&userInfo.Artworks, sql3, userId)
		if _err != nil {
			_err = errors.Wrap(_err, "GetUserPanel sql3 fail")
		}
		return _err
	})

	eg.Go(func() error {
		sql4 := `select rating from commission_count where user_id = ?`
		_err := db.Get(&userInfo, sql4, userId)
		if _err != nil {
			_err = errors.Wrap(_err, "GetUserPanel sql4 fail")
		}
		return _err
	})

	err = eg.Wait()

	return
}

// GetUserCount 获取用户统计数据
func GetUserCount(userId string) (userCount models.UserAllCount, err error) {
	sql := `select likes,collects,fans,following,art_count,trend_count,collect_count from user_count where user_id = ?`
	err = db.Get(&userCount, sql, userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserCount sql fail")
	}
	return
}

// CheckUserFocus 查询某个用户是否关注过登陆的用户
func CheckUserFocus(checkList []string, userId string) (res []models.UserIsFocus, err error) {
	if len(checkList) == 0 {
		return
	}
	var sqlStrList []string
	var uIds []interface{}

	for _, uid := range checkList {
		str := fmt.Sprintf(`(SELECT COALESCE(user_id,%s) as user_id, SIGN(count(*)) as is_focus 
  		FROM user_focus WHERE user_id = %s and focus_id = %s AND is_cancel = 0)`, uid, uid, userId)
		sqlStrList = append(sqlStrList, str)
	}

	sqlStr := strings.Join(sqlStrList, " UNION ALL ")

	// 查询 数据库
	err = db.Select(&res, sqlStr, uIds...)
	if err != nil {
		err = errors.Wrap(err, "CheckUserFocus get fail")
		return
	}

	return
}

// GetUserInvitationCode 查找用户邀请码
func GetUserInvitationCode(userId string) (res []models.InvitationCode, err error) {
	sqlStr := `SELECT code,used,IFNULL(avatar_name,'') as avatar,IFNULL(username,'') as userName FROM invite_code 
				LEFT JOIN user_profile
				on invite_code.used = user_profile.user_id
				WHERE owner = ? and used != ?`
	// 查询 数据库
	err = db.Select(&res, sqlStr, userId, userId)
	if err != nil {
		err = errors.Wrap(err, "GetUserInvitationCode get fail")
		return
	}

	return
}

// GetUserFans 获取用户粉丝
func GetUserFans(focusId string, nextId string, limit int) (fans []string, err error) {
	sqlStr := `SELECT user_id 
			   FROM user_focus WHERE  focus_id = ? and is_cancel = 0 and user_id > ? 
			   ORDER BY user_id LIMIT ?`
	// 查询 数据库
	err = db.Select(&fans, sqlStr, focusId, nextId, limit)
	if err != nil {
		err = errors.Wrap(err, "GetUserInvitationCode get fail")
		return
	}

	return
}

// GetAllUserShowId 获取需要展示的用户id
func GetAllUserShowId(query models.AllUserShowQuery) (userData []models.BigCardUserId, err error) {
	var sqlStr = "select user_id,score,'' as createAT  from user_count where art_count >= 1"
	if query.Type == "new" {
		if query.Next != "0" {
			sqlStr = sqlStr + " and user_id < ?"
		}
		sqlStr = sqlStr + ` order by user_id desc limit 20`
	} else if query.Type == "hot" {
		if query.Next != "0" {
			page, _ := strconv.Atoi(query.Next)
			query.Next = strconv.Itoa(page * 20)
			sqlStr = sqlStr + ` order by score desc,likes desc limit ` + `?,20;`
		} else {
			sqlStr = sqlStr + ` order by score desc,likes desc limit 0,20;`
		}
	} else {
		if query.Next != "0" {
			sqlStr = `SELECT snow_id as user_id, MAX(artwork_id) as createAT FROM user
						inner join artwork ak on user.snow_id = ak.user_id
						WHERE is_delete = 0 And whoSee = 'public' and state = 0 and artwork_id < ?
						GROUP BY snow_id
						ORDER BY createAT DESC
						LIMIT 20`
		} else {
			sqlStr = `SELECT snow_id as user_id, MAX(artwork_id) as createAT FROM user
						inner join artwork ak on user.snow_id = ak.user_id
						WHERE is_delete = 0 And whoSee = 'public' and state = 0 
						GROUP BY snow_id
						ORDER BY createAT DESC
						LIMIT 20`
		}
	}

	if query.Next != "0" {
		err = db.Select(&userData, sqlStr, query.Next)
	} else {
		err = db.Select(&userData, sqlStr)
	}

	if err != nil {
		err = errors.Wrap(err, "GetAllUserShowInfo sqlStr fail")
		return
	}

	return
}
